package calc

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

// The functions in this file provide light abstractions over common DOM
// interactions. They intentionally avoid the more advanced polling helpers from
// chromedp, as some of those rely on browser features such as
// SharedArrayBuffer which are not available on every page.

// Click dispatches a click event using either a CSS selector or XPath query.
func Click(sel string, isXPath bool) chromedp.Action {
	if isXPath {
		return chromedp.Click(sel, chromedp.BySearch)
	}
	return chromedp.Click(sel)
}

// SetValue assigns a value to an element located via CSS or XPath.
func SetValue(sel, value string, isXPath bool) chromedp.Action {
	if isXPath {
		return chromedp.SetValue(sel, value, chromedp.BySearch)
	}
	return chromedp.SetValue(sel, value)
}

// WaitVisible polls for an element to be present and visible in the DOM. It
// implements the polling on the Go side to avoid relying on page helpers that
// may use SharedArrayBuffer/Atomics, which can be unavailable when the target
// page is not cross‑origin isolated.
func WaitVisible(sel string, isXPath bool) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		query := fmt.Sprintf(`document.querySelector(%q)`, sel)
		if isXPath {
			query = fmt.Sprintf(`document.evaluate(%q, document, null, XPathResult.FIRST_ORDERED_NODE_TYPE, null).singleNodeValue`, sel)
		}
		script := fmt.Sprintf(`(() => { const el = %s; return el && (el.offsetWidth || el.offsetHeight || el.getClientRects().length); })()`, query)
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			var visible bool
			err := chromedp.Evaluate(script, &visible).Do(ctx)
			if err == nil && visible {
				return nil
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
			}
		}
	})
}
