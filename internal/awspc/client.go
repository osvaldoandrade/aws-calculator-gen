package awspc

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	bcm "github.com/aws/aws-sdk-go-v2/service/bcmpricingcalculator"
	bcmtypes "github.com/aws/aws-sdk-go-v2/service/bcmpricingcalculator/types"
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
	return &AWSClient{svc: bcm.NewFromConfig(cfg)}, nil
}

// AWSClient calls the real AWS Pricing Calculator API.
type AWSClient struct {
	svc *bcm.Client
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
	entries := defaultEntries(prefix, profile)
	if len(entries) == 0 {
		return id, nil
	}

	portion := amount / float64(len(entries))
	for i := range entries {
		entries[i].Key = aws.String(strconv.Itoa(i + 1))
		entries[i].UsageAccountId = aws.String("123456789012")
		entries[i].Amount = aws.Float64(portion)
	}
	_, err = c.svc.BatchCreateWorkloadEstimateUsage(ctx, &bcm.BatchCreateWorkloadEstimateUsageInput{
		WorkloadEstimateId: aws.String(id),
		Usage:              entries,
	})
	if err != nil {
		return "", err
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

func defaultEntries(prefix, profile string) []bcmtypes.BatchCreateWorkloadEstimateUsageEntry {
	if profile == "lake" {
		return []bcmtypes.BatchCreateWorkloadEstimateUsageEntry{
			{
				ServiceCode: aws.String("AmazonDynamoDB"),
				UsageType:   aws.String(prefix + "-TimedStorage-ByteHrs"),
			},
			{
				ServiceCode: aws.String("AmazonRedshift"),
				UsageType:   aws.String(prefix + "-Redshift:ServerlessUsage"),
			},
			{
				ServiceCode: aws.String("AWSGlue"),
				UsageType:   aws.String(prefix + "-ETL-Flex-DPU-Hour"),
			},
			{
				ServiceCode: aws.String("AmazonAthena"),
				UsageType:   aws.String(prefix + "-DataScannedInTB"),
			},
		}
	}
	// default transactional profile
	return []bcmtypes.BatchCreateWorkloadEstimateUsageEntry{
		{
			ServiceCode: aws.String("AmazonEC2"),
			UsageType:   aws.String(prefix + "-BoxUsage:m7g.large"),
			Operation:   aws.String("RunInstances"),
		},
		{
			ServiceCode: aws.String("AmazonS3"),
			UsageType:   aws.String(prefix + "-TimedStorage-ByteHrs"),
			Operation:   aws.String("PutObject"),
		},
		{
			ServiceCode: aws.String("AmazonRDS"),
			UsageType:   aws.String(prefix + "-InstanceUsage:db.m7g.large"),
			Operation:   aws.String("CreateDBInstance"),
		},
		{
			ServiceCode: aws.String("AWSLambda"),
			UsageType:   aws.String(prefix + "-Lambda-GB-Second"),
			Operation:   aws.String("Invoke"),
		},
	}
}
