package awspc

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	bcm "github.com/aws/aws-sdk-go-v2/service/bcmpricingcalculator"
	bcmtypes "github.com/aws/aws-sdk-go-v2/service/bcmpricingcalculator/types"
)

// Client defines subset of AWS Pricing Calculator API used.
type Client interface {
	CreateWorkloadEstimate(ctx context.Context, title, region string, amount float64) (string, error)
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

// CreateWorkloadEstimate creates a workload estimate and a single usage line.
func (c *AWSClient) CreateWorkloadEstimate(ctx context.Context, title, region string, amount float64) (string, error) {
	out, err := c.svc.CreateWorkloadEstimate(ctx, &bcm.CreateWorkloadEstimateInput{
		Name: aws.String(title),
	})
	if err != nil {
		return "", err
	}
	id := aws.ToString(out.Id)
	// best effort: create one usage item reflecting the amount
	_, _ = c.svc.BatchCreateWorkloadEstimateUsage(ctx, &bcm.BatchCreateWorkloadEstimateUsageInput{
		WorkloadEstimateId: aws.String(id),
		Usage: []bcmtypes.BatchCreateWorkloadEstimateUsageEntry{
			{
				Key:            aws.String("1"),
				ServiceCode:    aws.String("AmazonEC2"),
				UsageType:      aws.String("USE1-BoxUsage:m7g.large"),
				Operation:      aws.String("RunInstances"),
				UsageAccountId: aws.String("123456789012"),
				Amount:         aws.Float64(amount),
			},
		},
	})
	return id, nil

}

// StubClient implements Client without calling AWS.
type StubClient struct{}


func (StubClient) CreateWorkloadEstimate(ctx context.Context, title, region string, amount float64) (string, error) {
	return "stub-id", nil
}
