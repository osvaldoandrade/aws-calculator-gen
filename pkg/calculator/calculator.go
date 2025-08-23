package calculator

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/go-playground/validator/v10"
	"golang.org/x/exp/slog"
)

var validate = validator.New()

// Options carries configuration for the crawling client.
type Options struct {
	Headless bool
	Debug    bool
	Logger   *slog.Logger // required
}

// Client models interactions with the AWS Pricing Calculator UI.
type Client interface {
	NewEstimate(ctx context.Context, name, currency, region string) error
	AddS3(ctx context.Context, p S3Params) (LineItem, error)
	ExportShareURL(ctx context.Context) (string, error)
	Close(ctx context.Context) error
}

// calculatorClient is a minimal chromedp-based crawler used to drive the
// public AWS Pricing Calculator UI. Only the S3 service is implemented for now;
// other services can be added following the same pattern.
type calculatorClient struct {
	opts   Options
	ctx    context.Context
	cancel context.CancelFunc
}

// NewClient initializes a new crawler instance.
func NewClient(opts Options) (Client, error) {
	if opts.Logger == nil {
		return nil, fmt.Errorf("logger required")
	}
	allocOpts := chromedp.DefaultExecAllocatorOptions[:]
	if opts.Headless || !opts.Debug {
		allocOpts = append(allocOpts, chromedp.Flag("headless", true))
	}
	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), allocOpts...)
	ctx, cancel := chromedp.NewContext(allocCtx)
	return &calculatorClient{opts: opts, ctx: ctx, cancel: cancel}, nil
}

// Common carries fields shared across service parameter structs.
type Common struct {
	Region string `validate:"required"`
}

// S3Params defines parameters for Amazon S3.
type S3Params struct {
	Common
	StorageClass   string  `validate:"required"`
	StorageGB      float64 `validate:"gte=0"`
	PUTRequestsMln float64
	GETRequestsMln float64
	DataOutGB      float64
}

// LineItem represents a service entry in the estimate.
type LineItem struct {
	Service    string
	Title      string
	MonthlyUSD float64
}

// NewEstimate opens the calculator and creates a new estimate with the given
// title, currency and region.
func (c *calculatorClient) NewEstimate(ctx context.Context, name, currency, region string) error {
	if err := chromedp.Run(c.ctx, chromedp.Navigate("https://calculator.aws/")); err != nil {
		return err
	}
	if err := chromedp.Run(c.ctx, chromedp.Click(`text="Create estimate"`, chromedp.NodeVisible)); err != nil {
		return err
	}
	if currency != "" {
		_ = chromedp.Run(c.ctx, chromedp.SetValue(`#currency`, currency))
	}
	if name != "" {
		_ = chromedp.Run(c.ctx, chromedp.SetValue(`#estimateName`, name))
	}
	if region != "" {
		_ = chromedp.Run(c.ctx, chromedp.SetValue(`#region`, region))
	}
	return nil
}

// AddS3 adds a minimal S3 line item to the estimate.
func (c *calculatorClient) AddS3(ctx context.Context, p S3Params) (LineItem, error) {
	if err := validate.Struct(p); err != nil {
		return LineItem{}, err
	}
	url := "https://calculator.aws/#/createCalculator/S3"
	if err := chromedp.Run(c.ctx, chromedp.Navigate(url)); err != nil {
		return LineItem{}, err
	}
	if err := chromedp.Run(c.ctx, chromedp.SetValue(`#s3-storage-class`, p.StorageClass)); err != nil {
		return LineItem{}, err
	}
	if err := chromedp.Run(c.ctx, chromedp.SetValue(`#s3-storage-amount`, strconv.FormatFloat(p.StorageGB, 'f', -1, 64))); err != nil {
		return LineItem{}, err
	}
	if err := chromedp.Run(c.ctx, chromedp.Click(`text="Add to my estimate"`)); err != nil {
		return LineItem{}, err
	}
	return LineItem{Service: "S3", Title: "Amazon S3"}, nil
}

// ExportShareURL clicks through the UI to obtain a public share link.
func (c *calculatorClient) ExportShareURL(ctx context.Context) (string, error) {
	if err := chromedp.Run(c.ctx, chromedp.Navigate("https://calculator.aws/#/estimate")); err != nil {
		return "", err
	}
	if err := chromedp.Run(c.ctx, chromedp.Click(`text="Share"`, chromedp.NodeVisible)); err != nil {
		return "", err
	}
	time.Sleep(time.Second)
	var link string
	if err := chromedp.Run(c.ctx, chromedp.AttributeValue(`input[value^="https://calculator.aws"]`, "value", &link, nil)); err != nil {
		return "", err
	}
	if link == "" {
		return "", fmt.Errorf("share url not found")
	}
	return link, nil
}

// Close releases browser resources.
func (c *calculatorClient) Close(ctx context.Context) error {
	if c.cancel != nil {
		c.cancel()
	}
	return nil
}
