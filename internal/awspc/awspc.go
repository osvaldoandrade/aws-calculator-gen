package awspc

import (
	"context"
	"fmt"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	pc "github.com/aws/aws-sdk-go-v2/service/bcmpricingcalculator"
)

// Client defines operations needed from AWS Pricing Calculator.
type Client interface {
	CreateEstimate(ctx context.Context, title string) (string, error)
}

// AWSClient implements Client using AWS SDK.
type AWSClient struct {
	pc *pc.Client
}

// New creates AWSClient with default config.
func New(ctx context.Context) (*AWSClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &AWSClient{pc: pc.NewFromConfig(cfg)}, nil
}

// CreateEstimate creates workload estimate and returns id.
func (c *AWSClient) CreateEstimate(ctx context.Context, title string) (string, error) {
	out, err := c.pc.CreateWorkloadEstimate(ctx, &pc.CreateWorkloadEstimateInput{
		Name: aws.String(title),
	})
	if err != nil {
		return "", err
	}
	if out.Id == nil {
		return "", fmt.Errorf("missing id")
	}
	return aws.ToString(out.Id), nil
}

// URL returns console deep link for estimate.
func URL(id string) string {
	return fmt.Sprintf("https://console.aws.amazon.com/costmanagement/home#/pricing-calculator/workload-estimates/%s", id)
}
