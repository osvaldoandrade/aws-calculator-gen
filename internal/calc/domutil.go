package calc

import "github.com/chromedp/chromedp"

// The functions in this file provide light abstractions over common DOM
// interactions. They intentionally rely on chromedp's "search" based querying
// to traverse shadow roots without injecting scripts that depend on browser
// features such as SharedArrayBuffer.

// Click dispatches a click event using a selector that can be CSS, XPath or
// plain text.
func Click(sel string) chromedp.Action {
	return chromedp.Click(sel, chromedp.BySearch)
}

// SetValue assigns a value to an element located via CSS, XPath or text
// selector.
func SetValue(sel, value string) chromedp.Action {
	return chromedp.SetValue(sel, value, chromedp.BySearch)
}

// WaitVisible waits until the element matching the selector is visible.
func WaitVisible(sel string) chromedp.Action {
	return chromedp.WaitVisible(sel, chromedp.BySearch)
}
