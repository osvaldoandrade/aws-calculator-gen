package calc

import "github.com/chromedp/chromedp"

// The functions in this file are placeholders and should be expanded with
// robust DOM interaction helpers.

func Click(sel string, isXPath bool) chromedp.Action {
	if isXPath {
		return chromedp.Click(sel, chromedp.BySearch)
	}
	return chromedp.Click(sel)
}

func SetValue(sel, value string, isXPath bool) chromedp.Action {
	if isXPath {
		return chromedp.SetValue(sel, value, chromedp.BySearch)
	}
	return chromedp.SetValue(sel, value)
}

func WaitVisible(sel string, isXPath bool) chromedp.Action {
	if isXPath {
		return chromedp.WaitVisible(sel, chromedp.BySearch)
	}
	return chromedp.WaitVisible(sel)
}
