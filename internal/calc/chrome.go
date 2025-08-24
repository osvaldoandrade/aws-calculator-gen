package calc

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
)

// newBrowser creates a chromedp context bound to a parent context and optional
// timeout. The implementation is minimal and intended as a placeholder for a
// more complete setup with timeouts and flags.
func newBrowser(parent context.Context, headful bool, timeout time.Duration) (context.Context, context.CancelFunc, error) {
	opts := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", !headful),
		chromedp.DisableGPU,
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
	}
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(parent, opts...)
	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	if timeout > 0 {
		var cancelTimeout context.CancelFunc
		ctx, cancelTimeout = context.WithTimeout(ctx, timeout)
		cancel := func() {
			cancelTimeout()
			cancelCtx()
			cancelAlloc()
		}
		return ctx, cancel, nil
	}
	cancel := func() {
		cancelCtx()
		cancelAlloc()
	}
	return ctx, cancel, nil
}
