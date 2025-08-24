package calc

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/chromedp/chromedp"
)

type Orchestrator struct {
	EstimateName string
	RegionCode   string
	TargetMRR    float64
	Headful      bool
	Tolerance    float64
	Timeout      time.Duration
	MaxRetries   int
}

type Result struct {
	ShareURL      string
	RegionLabel   string
	InstanceType  string
	Count         int
	AchievedMRR   float64
	RelativeError float64
}

type option struct {
	typ   string
	price float64
}

var instanceOptions = []option{
	{typ: "t3.micro", price: 8.5},
	{typ: "t3.small", price: 17.0},
	{typ: "t3.medium", price: 34.0},
	{typ: "m7i.large", price: 102.0},
	{typ: "c7i.xlarge", price: 220.0},
}

// selectInstance chooses an instance type and count that best matches the
// desired monthly recurring revenue using a simple heuristic across different
// instance sizes.
func selectInstance(target float64) (string, int, float64, float64) {
	bestType := ""
	bestCount := 0
	bestAchieved := 0.0
	bestError := math.MaxFloat64
	for _, opt := range instanceOptions {
		count := int(ceil(target / opt.price))
		achieved := float64(count) * opt.price
		rel := abs(achieved-target) / target
		if rel < bestError {
			bestError = rel
			bestType = opt.typ
			bestCount = count
			bestAchieved = achieved
		}
	}
	return bestType, bestCount, bestAchieved, bestError
}

// Run executes the pricing calculator automation using the real AWS calculator.
func (o *Orchestrator) Run(ctx context.Context) (Result, error) {
	instType, count, achieved, relErr := selectInstance(o.TargetMRR)

	timeout := o.Timeout
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}

	bctx, cancel, err := newBrowser(ctx, o.Headful, timeout)
	if err != nil {
		return Result{}, err
	}
	defer cancel()

	var shareURL string

	url := "https://calculator.aws/#/addService"
	if o.RegionCode != "" {
		url = fmt.Sprintf("https://calculator.aws/#/addService?region=%s", o.RegionCode)
	}

	err = chromedp.Run(bctx,
		chromedp.Navigate("https://calculator.aws/#/"),
		WaitVisible(CreateEstimateXPath),
		Click(CreateEstimateXPath),
		chromedp.Navigate(url),
		WaitVisible(SearchInputCSS),
		SetValue(SearchInputCSS, "EC2"),
		chromedp.SendKeys(SearchInputCSS, "\n", chromedp.BySearch),
		WaitVisible(EC2ConfigureXPath),
		Click(EC2ConfigureXPath),
		WaitVisible(NumberInstancesCSS),
		SetValue(NumberInstancesCSS, fmt.Sprintf("%d", count)),
		SetValue(SearchInputCSS, instType),
		chromedp.SendKeys(SearchInputCSS, "\n", chromedp.BySearch),
		Click(SaveAndAddXPath),
		WaitVisible(ViewSummaryXPath),
		Click(ViewSummaryXPath),
		WaitVisible(EditNameLinkCSS),
		Click(EditNameLinkCSS),
		SetValue(EstimateNameInput, o.EstimateName),
		Click(SaveNameButton),
		Click(ShareButtonXPath),
		chromedp.Value(ShareLinkXPath, &shareURL, chromedp.BySearch),
	)
	if err != nil {
		return Result{}, err
	}

	return Result{
		ShareURL:      shareURL,
		RegionLabel:   regionLabelFromCode(o.RegionCode),
		InstanceType:  instType,
		Count:         count,
		AchievedMRR:   achieved,
		RelativeError: relErr,
	}, nil
}
