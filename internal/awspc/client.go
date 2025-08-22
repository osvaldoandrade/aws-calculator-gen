package awspc

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

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

	var prevIDs []string
	for attempt := 0; attempt < 5; attempt++ {
		assignUsage(lines, amount)
		svcByKey := map[string]string{}
		var usage []bcmtypes.BatchCreateWorkloadEstimateUsageEntry
		for i := range lines {
			if lines[i].Amount == nil || *lines[i].Amount == 0 {
				continue
			}
			cost := *lines[i].Amount * lines[i].price
			svc := ""
			if lines[i].ServiceCode != nil {
				svc = *lines[i].ServiceCode
			}
			fmt.Printf("adding service %s cost %.2f\n", svc, cost)
			key := strconv.Itoa(len(usage) + 1)
			lines[i].Key = aws.String(key)
			lines[i].UsageAccountId = aws.String(c.accountID)
			svcByKey[key] = svc
			usage = append(usage, lines[i].BatchCreateWorkloadEstimateUsageEntry)
		}

		if len(prevIDs) > 0 {
			_, err = c.svc.BatchDeleteWorkloadEstimateUsage(ctx, &bcm.BatchDeleteWorkloadEstimateUsageInput{
				WorkloadEstimateId: aws.String(id),
				Ids:                prevIDs,
			})
			if err != nil {
				return "", err
			}
		}

		prevIDs = prevIDs[:0]
		if len(usage) > 0 {
			out, err2 := c.svc.BatchCreateWorkloadEstimateUsage(ctx, &bcm.BatchCreateWorkloadEstimateUsageInput{
				WorkloadEstimateId: aws.String(id),
				Usage:              usage,
			})
			if err2 != nil {
				return "", err2
			}
			if len(out.Errors) > 0 {
				var msgs []string
				for _, e := range out.Errors {
					svc := svcByKey[aws.ToString(e.Key)]
					msgs = append(msgs, fmt.Sprintf("%s: %s", svc, aws.ToString(e.ErrorMessage)))
				}
				return "", fmt.Errorf("usage creation failed: %s", strings.Join(msgs, "; "))
			}
			for _, r := range out.Items {
				if r.Id != nil {
					prevIDs = append(prevIDs, aws.ToString(r.Id))
				}
			}
		}

		est, err := c.svc.GetWorkloadEstimate(ctx, &bcm.GetWorkloadEstimateInput{Identifier: aws.String(id)})
		if err != nil || est.TotalCost == nil {
			return "", err
		}
		fmt.Printf("calculator total %.2f\n", *est.TotalCost)
		diff := amount - *est.TotalCost
		if math.Abs(diff) < 0.01 {
			break
		}
		amount += diff
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
					ServiceCode: aws.String("AmazonS3"),
					UsageType:   aws.String(prefix + "-Requests-Tier1"),
					Operation:   aws.String("PutObject"),
				},
				// Tier 1 S3 requests are $0.005 per 1,000 requests.
				price: 0.000005, // per request

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
			{
				BatchCreateWorkloadEstimateUsageEntry: bcmtypes.BatchCreateWorkloadEstimateUsageEntry{
					ServiceCode: aws.String("AmazonAthena"),
					UsageType:   aws.String(prefix + "-DMLQueries"),
					Operation:   aws.String("RunQuery"),
				},
				price: 0.0005, // per DML query
			},
			{
				BatchCreateWorkloadEstimateUsageEntry: bcmtypes.BatchCreateWorkloadEstimateUsageEntry{
					ServiceCode: aws.String("AWSLambda"),
					UsageType:   aws.String(prefix + "-Lambda-GB-Second"),
					Operation:   aws.String("Invoke"),
				},
				price: 0.0000166667, // per GB-second
			},
			{
				BatchCreateWorkloadEstimateUsageEntry: bcmtypes.BatchCreateWorkloadEstimateUsageEntry{
					ServiceCode: aws.String("AmazonEC2"),
					UsageType:   aws.String(prefix + "-BoxUsage:m7g.large"),
					Operation:   aws.String("RunInstances"),
				},
				price: 0.096, // per hour
			},
		}
	}
	// transactional profile
	return []usageLine{
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
		{
			BatchCreateWorkloadEstimateUsageEntry: bcmtypes.BatchCreateWorkloadEstimateUsageEntry{
				ServiceCode: aws.String("AWSEvents"),
				UsageType:   aws.String(prefix + "-Event-64K-Chunks"),
				Operation:   aws.String("PutEvents"),
			},
			price: 0.000001, // per 64KB event chunk
		},
		{
			BatchCreateWorkloadEstimateUsageEntry: bcmtypes.BatchCreateWorkloadEstimateUsageEntry{
				ServiceCode: aws.String("AmazonStates"),
				UsageType:   aws.String(prefix + "-StateTransition"),
				Operation:   aws.String("StartExecution"),
			},
			price: 0.000025, // per state transition
		},
		{
			BatchCreateWorkloadEstimateUsageEntry: bcmtypes.BatchCreateWorkloadEstimateUsageEntry{
				ServiceCode: aws.String("AmazonElastiCache"),
				UsageType:   aws.String(prefix + "-NodeUsage:cache.t4g.small"),
				Operation:   aws.String("CreateCacheCluster"),
			},
			price: 0.034, // per node hour
		},
		{
			BatchCreateWorkloadEstimateUsageEntry: bcmtypes.BatchCreateWorkloadEstimateUsageEntry{
				ServiceCode: aws.String("AmazonS3"),
				UsageType:   aws.String(prefix + "-Requests-Tier1"),
				Operation:   aws.String("PutObject"),
			},
			// Tier 1 S3 requests cost $0.005 per 1,000 requests.
			price: 0.000005, // per request

		},
	}
}

func assignUsage(lines []usageLine, amount float64) {
	if amount <= 0 {
		return
	}
	// Group lines by service so each service receives an equal share
	// of the overall cost. If a service has multiple usage lines, split
	// that service's share evenly across them.
	services := map[string][]int{}
	for i := range lines {
		if lines[i].price <= 0 {
			continue
		}
		svc := ""
		if lines[i].ServiceCode != nil {
			svc = *lines[i].ServiceCode
		}
		services[svc] = append(services[svc], i)
	}
	if len(services) == 0 {
		return
	}
	perService := amount / float64(len(services))
	total := 0.0
	for _, idxs := range services {
		perLine := perService / float64(len(idxs))
		for _, i := range idxs {
			units := perLine / lines[i].price
			if units > 0 {
				lines[i].Amount = aws.Float64(units)
				total += units * lines[i].price
			}
		}
	}
	if diff := amount - total; math.Abs(diff) > 1e-6 {
		for i := range lines {
			if lines[i].price > 0 && lines[i].Amount != nil {
				*lines[i].Amount += diff / lines[i].price
				break
			}
		}
	}
}
