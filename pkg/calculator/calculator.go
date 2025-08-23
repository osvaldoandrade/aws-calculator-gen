package calculator

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

// Client wraps a chromedp browser instance for interacting with the AWS pricing calculator.
type Client struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// Open launches a headless Chrome instance and returns a Client.
func Open(parent context.Context) (*Client, error) {
	// Allocate a new Chrome instance with sensible defaults for headless execution.
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Headless,
		chromedp.DisableGPU,
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(parent, opts...)
	ctx, cancelCtx := chromedp.NewContext(allocCtx)

	// Ensure the browser starts correctly.
	if err := chromedp.Run(ctx); err != nil {
		cancelCtx()
		cancelAlloc()
		return nil, err
	}

	return &Client{ctx: ctx, cancel: func() {
		cancelCtx()
		cancelAlloc()
	}}, nil
}

// Close closes the browser and releases all associated resources.
func (c *Client) Close() error {
	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
	}
	return nil
}

// NewEstimate opens the calculator site and starts a new estimate.
func (c *Client) NewEstimate() error {
	return chromedp.Run(c.ctx,
		chromedp.Navigate("https://calculator.aws/#/createCalculator"),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
	)
}

// S3Params describes the input parameters for adding an S3 service line.
type S3Params struct {
	StorageGB         float64
	PUTRequests       int
	GETRequests       int
	DataTransferOutGB float64
}

// AddS3 adds an S3 service line to the current estimate using the provided parameters.
func (c *Client) AddS3(p S3Params) error {
	tasks := chromedp.Tasks{
		// Open the add service dialog and choose S3.
		chromedp.Click(`button[aria-label="Add service"]`, chromedp.ByQuery),
		chromedp.Sleep(500 * time.Millisecond),
		chromedp.Click(`//span[contains(., "Amazon Simple Storage Service (S3)")]`, chromedp.BySearch),
		chromedp.Sleep(500 * time.Millisecond),

		// Fill out the form fields. Selectors may need to be adjusted if AWS updates the UI.
		chromedp.SetValue(`input[name="storageAmount"]`, fmt.Sprintf("%.2f", p.StorageGB), chromedp.ByQuery),
		chromedp.SetValue(`input[name="putRequests"]`, fmt.Sprintf("%d", p.PUTRequests), chromedp.ByQuery),
		chromedp.SetValue(`input[name="getRequests"]`, fmt.Sprintf("%d", p.GETRequests), chromedp.ByQuery),
		chromedp.SetValue(`input[name="dataTransferOut"]`, fmt.Sprintf("%.2f", p.DataTransferOutGB), chromedp.ByQuery),

		chromedp.Click(`button[aria-label="Add to my estimate"]`, chromedp.ByQuery),
	}
	return chromedp.Run(c.ctx, tasks)
}
