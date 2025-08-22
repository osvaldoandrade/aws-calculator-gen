package awspc

import (
	"context"
	"math"
	"sort"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	bcm "github.com/aws/aws-sdk-go-v2/service/bcmpricingcalculator"
	bcmtypes "github.com/aws/aws-sdk-go-v2/service/bcmpricingcalculator/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// Client defines subset of AWS Pricing Calculator API used.
type Client interface {
	CreateWorkloadEstimate(ctx context.Context, title, region, profile string, amount float64) (string, error)
}

// New tries to create a real AWS Pricing Calculator client.
func New(ctx context.Context) (Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	stsClient := sts.NewFromConfig(cfg)
	ident, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, err
	}
	return &AWSClient{svc: bcm.NewFromConfig(cfg), accountID: aws.ToString(ident.Account)}, nil
}

// AWSClient calls the real AWS Pricing Calculator API.
type AWSClient struct {
	svc       *bcm.Client
	accountID string
}

// CreateWorkloadEstimate creates a workload estimate and several usage lines.
func (c *AWSClient) CreateWorkloadEstimate(ctx context.Context, title, region, profile string, amount float64) (string, error) {
	out, err := c.svc.CreateWorkloadEstimate(ctx, &bcm.CreateWorkloadEstimateInput{
		Name: aws.String(title),
	})
	if err != nil {
		return "", err
	}
	id := aws.ToString(out.Id)

	prefix := regionPrefix(region)
	lines := defaultEntries(prefix, profile)
	if len(lines) == 0 {
		return id, nil
	}

	assignUsage(lines, amount)
	var usage []bcmtypes.BatchCreateWorkloadEstimateUsageEntry
	for i := range lines {
		if lines[i].Amount == nil || *lines[i].Amount == 0 {
			continue
		}
		lines[i].Key = aws.String(strconv.Itoa(len(usage) + 1))
		lines[i].UsageAccountId = aws.String(c.accountID)
		usage = append(usage, lines[i].BatchCreateWorkloadEstimateUsageEntry)
	}
	if len(usage) > 0 {
		_, err = c.svc.BatchCreateWorkloadEstimateUsage(ctx, &bcm.BatchCreateWorkloadEstimateUsageInput{
			WorkloadEstimateId: aws.String(id),
			Usage:              usage,
		})
		if err != nil {
			return "", err
		}
	}
	return id, nil
}

// StubClient implements Client without calling AWS.
type StubClient struct{}

func (StubClient) CreateWorkloadEstimate(ctx context.Context, title, region, profile string, amount float64) (string, error) {
	_ = region
	_ = profile
	return "stub-id", nil
}

func regionPrefix(region string) string {
	switch region {
	case "us-east-1":
		return "USE1"
	case "us-west-2":
		return "USW2"
	case "eu-west-1":
		return "EUW1"
	case "sa-east-1":
		return "SAE1"
	default:
		return "USE1"
	}
}

type usageLine struct {
	bcmtypes.BatchCreateWorkloadEstimateUsageEntry
	price float64
}

func defaultEntries(prefix, profile string) []usageLine {
	if profile == "lake" {
		return []usageLine{
			{
				BatchCreateWorkloadEstimateUsageEntry: bcmtypes.BatchCreateWorkloadEstimateUsageEntry{
					ServiceCode: aws.String("AmazonDynamoDB"),
					UsageType:   aws.String(prefix + "-TimedStorage-ByteHrs"),
					Operation:   aws.String("CreateTable"),
				},
				price: 0.25, // per GB-month
			},
			{
				BatchCreateWorkloadEstimateUsageEntry: bcmtypes.BatchCreateWorkloadEstimateUsageEntry{
					ServiceCode: aws.String("AmazonRedshift"),
					UsageType:   aws.String(prefix + "-Redshift:ServerlessUsage"),
					Operation:   aws.String("CreateWorkgroup"),
				},
				price: 0.5, // per RPU-hour
			},
			{
				BatchCreateWorkloadEstimateUsageEntry: bcmtypes.BatchCreateWorkloadEstimateUsageEntry{
					ServiceCode: aws.String("AWSGlue"),
					UsageType:   aws.String(prefix + "-ETL-Flex-DPU-Hour"),
					Operation:   aws.String("StartJobRun"),
				},
				price: 0.44, // per DPU-hour
			},
			{
				BatchCreateWorkloadEstimateUsageEntry: bcmtypes.BatchCreateWorkloadEstimateUsageEntry{
					ServiceCode: aws.String("AmazonAthena"),
					UsageType:   aws.String(prefix + "-DataScannedInTB"),
					Operation:   aws.String("RunQuery"),
				},
				price: 5.0, // per TB scanned
			},
		}
	}
	// default transactional profile
	return []usageLine{
		{
			BatchCreateWorkloadEstimateUsageEntry: bcmtypes.BatchCreateWorkloadEstimateUsageEntry{
				ServiceCode: aws.String("AmazonEC2"),
				UsageType:   aws.String(prefix + "-BoxUsage:m7g.large"),
				Operation:   aws.String("RunInstances"),
			},
			price: 0.096, // per hour
		},
		{
			BatchCreateWorkloadEstimateUsageEntry: bcmtypes.BatchCreateWorkloadEstimateUsageEntry{
				ServiceCode: aws.String("AmazonS3"),
				UsageType:   aws.String(prefix + "-TimedStorage-ByteHrs"),
				Operation:   aws.String("PutObject"),
			},
			price: 0.023, // per GB-month
		},
		{
			BatchCreateWorkloadEstimateUsageEntry: bcmtypes.BatchCreateWorkloadEstimateUsageEntry{
				ServiceCode: aws.String("AmazonRDS"),
				UsageType:   aws.String(prefix + "-InstanceUsage:db.m7g.large"),
				Operation:   aws.String("CreateDBInstance"),
			},
			price: 0.206, // per hour
		},
		{
			BatchCreateWorkloadEstimateUsageEntry: bcmtypes.BatchCreateWorkloadEstimateUsageEntry{
				ServiceCode: aws.String("AWSLambda"),
				UsageType:   aws.String(prefix + "-Lambda-GB-Second"),
				Operation:   aws.String("Invoke"),
			},
			price: 0.0000166667, // per GB-second
		},
	}
}

func assignUsage(lines []usageLine, amount float64) {
	if amount <= 0 {
		return
	}
	remaining := amount
	for i := range lines {
		if remaining >= lines[i].price {
			lines[i].Amount = aws.Float64(1)
			remaining -= lines[i].price
		}
	}
	sort.Slice(lines, func(i, j int) bool { return lines[i].price > lines[j].price })
	for i := range lines {
		if remaining < lines[i].price {
			continue
		}
		units := math.Floor(remaining / lines[i].price)
		if units > 0 {
			if lines[i].Amount == nil {
				lines[i].Amount = aws.Float64(units)
			} else {
				*lines[i].Amount += units
			}
			remaining -= units * lines[i].price
		}
	}
	if remaining > 0 {
		if lines[0].Amount == nil {
			lines[0].Amount = aws.Float64(remaining / lines[0].price)
		} else {
			*lines[0].Amount += remaining / lines[0].price
		}
	}
}
