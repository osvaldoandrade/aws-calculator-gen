package calc

import (
	"context"

	"github.com/chromedp/chromedp"
)

// newBrowser creates a chromedp context. The implementation is minimal and
// intended as a placeholder for a more complete setup with timeouts and flags.
func newBrowser(headful bool) (context.Context, context.CancelFunc, error) {
	opts := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", !headful),
		chromedp.DisableGPU,
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
	}
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	cancel := func() {
		cancelCtx()
		cancelAlloc()
	}
	return ctx, cancel, nil
}
