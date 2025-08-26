package calc

import (
	"context"
	"fmt"
	"html"
	"log"
	"math"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v3"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
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

// ---- Selectors ----
const (
	createEstimateBtnXPath    = `//button[.//span[normalize-space()='Create estimate' or normalize-space()='Create Estimate']] | //span[normalize-space()='Create estimate' or normalize-space()='Create Estimate']/ancestor::button`
	findServiceInputCSS       = `input[aria-label="Find Service"]`
	ec2ConfigureXPath         = `//*[@data-cy="Amazon EC2 -button"]//button | //button[.//span[contains(normalize-space(.),'Configure Amazon EC2')]]`
	ec2ConfigHeaderXPath      = `//h1[contains(normalize-space(),'Configure Amazon EC2')]`
	numberInstancesInputXPath = `
		(
		  //input[contains(translate(@aria-label,'ABCDEFGHIJKLMNOPQRSTUVWXYZ','abcdefghijklmnopqrstuvwxyz'),'number of instances')]
		)[1]`
	instanceSearchInputCSS       = `input[aria-label*="Search instance types"], input[placeholder*="Search instance types"], input[aria-label*="Search by instance name"]`
	anyMemoryTriggerXPath        = `//*[@id='ec2enhancement']//*[contains(normalize-space(),'Any Memory')]/ancestor::*[self::button or self::div[contains(@class,'trigger')]] | //*[contains(@id,'trigger-content') and contains(normalize-space(),'Any Memory')]`
	anyVcpuTriggerXPath          = `//*[@id='ec2enhancement']//*[contains(normalize-space(),'Any vCPUs')]/ancestor::*[self::button or self::div[contains(@class,'trigger')]] | //*[contains(@id,'trigger-content') and contains(normalize-space(),'Any vCPUs')]`
	onDemandOptionXPath          = `//label[.//text()[contains(.,'On-Demand')]] | //input[@type='radio' and (contains(@value,'On-Demand') or contains(@aria-label,'On-Demand'))]/ancestor::label`
	appFooterCSS                 = `.appFooter`
	saveAndAddXPath              = `//button[@data-cy='Save and add service-button' and not(@disabled)] | //button[.//span[normalize-space()='Save and add service'] and not(@disabled)]`
	saveAndAddBtnFooterCSS       = `div.appFooter [data-cy='Save and add service-button']`
	viewSummaryBtnCSS            = `#estimate-button`
	editNameLinkCSS              = `a[data-cy="edit-estimate-name"]`
	editNameLinkXPath            = `//*[@data-cy='edit-estimate-name'] | //*[normalize-space()='Edit' and (self::button or self::span or self::a)]/ancestor::a | //*[contains(@class,'myEstimate')]//*[normalize-space()='Edit']`
	nameInputCSS                 = `input[aria-label="Enter Name"]`
	saveNameBtnXPath             = `//button[.//span[normalize-space()='Save'] and not(@disabled)]`
	numberInstancesInputCSSExact = `input[aria-label="Number of instances Enter amount"]`

	// Share
	shareBtnXPath            = `//*[@data-cy='save-and-share'] | //button[.//span[contains(normalize-space(.),'Share')]]`
	shareConsentDialogXPath  = `//div[@role='dialog' or contains(@class,'awsui-modal-root') or contains(@class,'awsui_modal-root')]`
	shareModalTitleXPath     = `(` + shareConsentDialogXPath + `)//*[self::h1 or self::h2][normalize-space()='Save estimate']`
	copyPublicLinkXPath      = `//button[.//span[normalize-space()='Copy public link']]`
	shareLinkInputCSS        = `div.clipboard-inputfield > input`
	shareLinkInputXPath      = `//div[contains(@class,'clipboard-inputfield')]//input | //input[@aria-label='Copy public link' or @aria-label='Public share link'] | //input[contains(@value,'calculator.aws') and contains(@value,'/estimate')]`
	shareAgreeContinueBtnCSS = `button[data-id="agree-continue"], button[aria-label="Agree and continue"], button[title="Agree and continue"]`
)

func (o *Orchestrator) Run(ctx context.Context) (Result, error) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	if fp, err := os.OpenFile("seidor-tools.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644); err == nil {
		log.SetOutput(fp)
		defer func(fp *os.File) {
			err := fp.Close()
			if err != nil {
				log.Printf("Error: %s", err)
			}
		}(fp)
	}

	// 1) Launch Chrome
	log.Printf("[1/10] Launching Chrome (headful=%v)...", o.Headful)
	allocOpts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("headless", !o.Headful),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("window-size", "1400,1000"),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
	}
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, allocOpts...)
	defer cancelAlloc()
	bctx, cancelBrowser := chromedp.NewContext(allocCtx, chromedp.WithLogf(func(format string, args ...interface{}) {
		log.Printf("[chromedp] "+format, args...)
	}))
	defer cancelBrowser()
	defer dumpHTML(bctx, "(always end)")

	// 2) Navigate
	base := "https://calculator.aws"
	navURL := buildURL(base, o.RegionCode, "")
	log.Printf("[2/10] Navigating to %s ...", navURL)
	if err := chromedp.Run(bctx, chromedp.Navigate(navURL)); err != nil {
		return Result{}, err
	}
	dismissCookieBanner(bctx)

	// 3) Optional "Create estimate"
	log.Printf("[3/10] Trying optional 'Create estimate' click if present...")
	_ = clickAny(bctx, []selector{{s: createEstimateBtnXPath, by: byXPath}})
	log.Printf("        Current URL: %s", currentURL(bctx))

	// 4) Ensure "Find Service"
	log.Printf("[4/10] Ensuring 'Find Service' input...")
	if err := chromedp.Run(bctx, chromedp.WaitVisible(findServiceInputCSS, chromedp.ByQuery)); err != nil {
		return Result{}, err
	}

	// 5) Type EC2
	log.Printf("[5/10] Typing 'EC2' into service finder...")
	if err := typeInto(bctx, findServiceInputCSS, byCSS, "EC2"); err != nil {
		log.Printf("       typing failed, JS-fallback...")
		if err2 := setInputValueJS(bctx, findServiceInputCSS, byCSS, "EC2"); err2 != nil {
			return Result{}, fmt.Errorf("cannot type 'EC2': %v / fb: %v", err, err2)
		}
		_ = chromedp.Run(bctx, chromedp.SendKeys(findServiceInputCSS, kb.Enter, chromedp.ByQuery))
	}

	// 6) Click "Configure Amazon EC2"
	log.Printf("[6/10] Selecting 'Configure Amazon EC2'...")
	if !clickAny(bctx, []selector{{s: ec2ConfigureXPath, by: byXPath}}) {
		_ = chromedp.Run(bctx, chromedp.SendKeys(findServiceInputCSS, kb.Enter, chromedp.ByQuery))
		if !clickAny(bctx, []selector{{s: ec2ConfigureXPath, by: byXPath}}) {
			return Result{}, fmt.Errorf("could not find 'Configure Amazon EC2'")
		}
	}
	log.Printf("        Current URL after configure: %s", currentURL(bctx))
	dismissCookieBanner(bctx)
	_ = waitVisibleWithTimeout(bctx, ec2ConfigHeaderXPath, byXPath, 5*time.Second)
	_ = waitVisibleWithTimeout(bctx, numberInstancesInputXPath, byXPath, 5*time.Second)

	// ---- PLANO (greedy desc, ignorando *.nano) ----
	plan := planGreedyEC2(o.TargetMRR /*mrr*/, o.Tolerance)
	totalApprox := 0.0
	for _, it := range plan {
		totalApprox += float64(it.Count) * it.Monthly
	}

	log.Printf("        PLAN (greedy desc, no nano) — target MRR=%.2f:", o.TargetMRR)
	for _, it := range plan {
		log.Printf("          - %d x %s  (~$%.2f/mo cada, ~$%.2f total)", it.Count, it.Name, it.Monthly, float64(it.Count)*it.Monthly)
	}
	log.Printf("        PLAN total ~= $%.2f/mo", totalApprox)

	for idx, it := range plan {
		// If not the first item, reopen the EC2 configurator
		if idx > 0 {
			_ = waitVisibleWithTimeout(bctx, findServiceInputCSS, byCSS, 5*time.Second)
			_ = typeInto(bctx, findServiceInputCSS, byCSS, "EC2")
			_ = chromedp.Run(bctx, chromedp.Sleep(200*time.Millisecond))
			if !clickAny(bctx, []selector{{s: ec2ConfigureXPath, by: byXPath}}) {
				_ = chromedp.Run(bctx, chromedp.SendKeys(findServiceInputCSS, kb.Enter, chromedp.ByQuery))
				if !clickAny(bctx, []selector{{s: ec2ConfigureXPath, by: byXPath}}) {
					return Result{}, fmt.Errorf("could not re-open 'Configure Amazon EC2' for planned item %d", idx+1)
				}
			}
			_ = waitVisibleWithTimeout(bctx, ec2ConfigHeaderXPath, byXPath, 5*time.Second)
			_ = waitVisibleWithTimeout(bctx, numberInstancesInputXPath, byXPath, 5*time.Second)
			dismissCookieBanner(bctx)
		}

		// Filters/price model
		_ = ensureAnyFilters(bctx, 5*time.Second)
		_ = ensureOnDemand(bctx, 5*time.Second)

		if err := setInstanceCount(bctx, it.Count); err != nil {
			dumpHTML(bctx, fmt.Sprintf("(set count %s failed)", it.Name))
			return Result{}, fmt.Errorf("could not set count for %q: %w", it.Name, err)
		}

		// Select instance by exact name, then set instance count
		if err := selectInstanceByName(bctx, it.Name, 5*time.Second); err != nil {
			dumpHTML(bctx, fmt.Sprintf("(select %s failed)", it.Name))
			return Result{}, fmt.Errorf("could not select instance %q: %w", it.Name, err)
		}

		// Save and add service, then continue to the next planned item
		if err := clickSaveAndAddService(bctx); err != nil {
			dumpHTML(bctx, fmt.Sprintf("(save/add %s failed)", it.Name))
			return Result{}, fmt.Errorf("could not save/add EC2 %s: %w", it.Name, err)
		}
		dumpHTML(bctx, fmt.Sprintf("(after save/add %s)", it.Name))
	}

	// After adding every planned service, only then view the summary
	if err := clickViewSummary(bctx); err != nil {
		dumpHTML(bctx, "(view summary failed)")
		return Result{}, fmt.Errorf("could not click 'View summary': %w", err)
	}
	_ = waitVisibleWithTimeout(bctx, shareBtnXPath, byXPath, 5*time.Second)

	// 8) Rename (optional)
	name := strings.TrimSpace(o.EstimateName)
	if name == "" {
		name = "Estimate-" + time.Now().Format("20060102-150405")
	}
	log.Printf("[8/10] Renaming estimate to %q (if controls are present)...", name)
	clickAny(bctx, []selector{{s: editNameLinkCSS, by: byCSS}, {s: editNameLinkXPath, by: byXPath}})
	if exists(bctx, nameInputCSS, byCSS) {
		_ = typeInto(bctx, nameInputCSS, byCSS, name)
		clickAny(bctx, []selector{{s: saveNameBtnXPath, by: byXPath}})
	}

	// 9) Open Share and handle consent
	log.Printf("[9/10] Opening 'Share' modal...")
	_ = scrollToTop(bctx)
	if err := clickRobust(bctx, shareBtnXPath, byXPath); err != nil {
		return Result{}, fmt.Errorf("could not open Share dialog/button: %w", err)
	}
	log.Printf("        Share clicked, handling consent if present...")
	dismissCookieBanner(bctx)

	shareURL := handleShareConsent(bctx)

	// 10) Buscar link (backoff 1s → 3s → 5s; reabrir Share se necessário)
	log.Printf("[10/10] Waiting for public link...")
	if !looksLikeShareURL(shareURL) {
		for _, d := range []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second} {
			_ = chromedp.Run(bctx, chromedp.Sleep(d))
			shareURL = getShareURLFromInputs(bctx)
			if looksLikeShareURL(shareURL) {
				break
			}
		}
	}
	if !looksLikeShareURL(shareURL) {
		// reabrir Share e tentar de novo rapidamente
		_ = clickWithTimeout(bctx, shareBtnXPath, byXPath, 3*time.Second)
		_ = chromedp.Run(bctx, chromedp.Sleep(800*time.Millisecond))
		shareURL = getShareURLFromInputs(bctx)
	}
	if !looksLikeShareURL(shareURL) {
		shareURL = getShareURLAnywhere(bctx)
	}
	if !looksLikeShareURL(shareURL) {
		shareURL = waitForShareLink(bctx, 3, 300*time.Millisecond)
	}
	dumpHTML(bctx, "(final snapshot)")

	if strings.TrimSpace(shareURL) == "" {
		return Result{}, fmt.Errorf("share link did not appear; see tmp.html")
	}
	log.Printf("[DONE] Share URL: %s", shareURL)

	relErr := 0.0
	if o.TargetMRR > 0 {
		relErr = math.Abs(totalApprox-o.TargetMRR) / o.TargetMRR
	}

	return Result{
		ShareURL:      shareURL,
		RegionLabel:   regionLabelFromCode(o.RegionCode),
		AchievedMRR:   totalApprox,
		RelativeError: relErr,
	}, nil
}

// ---- View summary helper ----

func clickViewSummary(ctx context.Context) error {
	_ = scrollToBottom(ctx)
	if err := clickWithTimeout(ctx, viewSummaryBtnCSS, byCSS, 5*time.Second); err == nil {
		_ = chromedp.Run(ctx, chromedp.Sleep(500*time.Millisecond))
		return nil
	}
	fallback := `//*[@id='estimate-button'] | //*[normalize-space()='View summary']/ancestor::*[self::button or self::a]`
	if err := clickWithTimeout(ctx, fallback, byXPath, 5*time.Second); err != nil {
		return err
	}
	_ = chromedp.Run(ctx, chromedp.Sleep(500*time.Millisecond))
	return nil
}

// ---- Consent / Share helpers ----

// Abre (ou confirma aberto) o modal de Share com timeout curto.
func ensureShareModalOpen(ctx context.Context) bool {
	// já está aberto?
	if exists(ctx, shareModalTitleXPath, byXPath) || exists(ctx, shareLinkInputXPath, byXPath) {
		return true
	}
	// tenta abrir
	_ = scrollToTop(ctx)
	if err := clickWithTimeout(ctx, shareBtnXPath, byXPath, 3*time.Second); err != nil {
		return false
	}
	_ = waitVisibleWithTimeout(ctx, shareConsentDialogXPath, byXPath, 3*time.Second)
	// título do modal ou campo de link visível?
	return exists(ctx, shareModalTitleXPath, byXPath) || exists(ctx, shareLinkInputXPath, byXPath)
}

// Tenta ler o link diretamente do(s) input(s) do modal (sem clicar em "Copy").
func tryReadShareURLFromModal(ctx context.Context) string {
	if v := getShareURLFromInputs(ctx); looksLikeShareURL(v) {
		return v
	}
	// às vezes o input está em um shadow root; tentar varredura ampla
	if v := deepScanShareInput(ctx); looksLikeShareURL(v) {
		return v
	}
	return ""
}

// Fluxo robusto: prioriza ler o link *antes* do clique; se fechar o modal, reabre e lê.
func handleShareConsent(ctx context.Context) string {
	// 1) Aceita consentimento (se existir) — best effort
	_ = clickWithTimeout(ctx, shareAgreeContinueBtnCSS, byCSS, 5*time.Second)
	dumpHTML(ctx, "(post-agree)")
	_ = chromedp.Run(ctx, chromedp.Sleep(800*time.Millisecond))

	// 2) Garante modal aberto e tenta ler sem clicar
	_ = ensureShareModalOpen(ctx)
	if sharedUrl := tryReadShareURLFromModal(ctx); looksLikeShareURL(sharedUrl) {
		return sharedUrl
	}

	// 3) Até 3 tentativas: clicar em "Copy", reabrir modal (se fechar) e ler o input
	for i := 1; i <= 3; i++ {
		log.Printf("        [share] try %d/3: copy, reopen-if-needed, read input", i)

		// clica no botão "Copy public link" se estiver visível; caso contrário tenta por texto
		if exists(ctx, copyPublicLinkXPath, byXPath) {
			_ = clickRobust(ctx, copyPublicLinkXPath, byXPath)
		} else {
			_ = deepClickButtonByText(ctx, "Copy public link")
		}

		// pequena espera para o possível fechamento do modal
		_ = chromedp.Run(ctx, chromedp.Sleep(800*time.Millisecond))

		// tenta ler de imediato (caso o modal continue aberto)
		if sharedUrl := tryReadShareURLFromModal(ctx); looksLikeShareURL(sharedUrl) {
			return sharedUrl
		}

		// se o modal fechou, reabra e leia
		if ensureShareModalOpen(ctx) {
			if sharedUrl := tryReadShareURLFromModal(ctx); looksLikeShareURL(sharedUrl) {
				return sharedUrl
			}
		}

		// fallback: varredura no DOM e snapshot
		if sharedUrl := getShareURLAnywhere(ctx); looksLikeShareURL(sharedUrl) {
			return sharedUrl
		}
		if snapURL := dumpAndExtract(ctx, fmt.Sprintf("(share try %d)", i)); looksLikeShareURL(snapURL) {
			return snapURL
		}
	}

	// Última cartada: reabrir o modal e tentar novamente a leitura direta
	if ensureShareModalOpen(ctx) {
		if sharedUrl := tryReadShareURLFromModal(ctx); looksLikeShareURL(sharedUrl) {
			return sharedUrl
		}
	}
	return ""
}

func clickWithTimeout(ctx context.Context, sel string, by queryBy, d time.Duration) error {
	c2, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	return clickRobust(c2, sel, by)
}

func waitVisibleWithTimeout(ctx context.Context, sel string, by queryBy, d time.Duration) error {
	c2, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	opts := queryOpts(by)
	return chromedp.Run(c2, chromedp.WaitVisible(sel, opts...))
}

func getShareURLAnywhere(ctx context.Context) string {
	if v := getShareURLFromInputs(ctx); looksLikeShareURL(v) {
		return v
	}
	if v := scanAnyShareInput(ctx); looksLikeShareURL(v) {
		return html.UnescapeString(strings.TrimSpace(v))
	}
	if v := deepScanShareInput(ctx); looksLikeShareURL(v) {
		return html.UnescapeString(strings.TrimSpace(v))
	}
	return ""
}

func getShareURLFromInputs(ctx context.Context) string {
	if v, ok := readInputValue(ctx, `input[aria-label="Copy public link"]`, byCSS); ok {
		v = html.UnescapeString(strings.TrimSpace(v))
		if looksLikeShareURL(v) {
			return v
		}
	}
	if v, ok := readInputValue(ctx, shareLinkInputCSS, byCSS); ok {
		v = html.UnescapeString(strings.TrimSpace(v))
		if looksLikeShareURL(v) {
			return v
		}
	}
	if v, ok := readInputValue(ctx, shareLinkInputXPath, byXPath); ok {
		v = html.UnescapeString(strings.TrimSpace(v))
		if looksLikeShareURL(v) {
			return v
		}
	}
	return ""
}

func waitForShareLink(ctx context.Context, tries int, delay time.Duration) string {
	for i := 0; i < tries; i++ {
		if v := getShareURLFromInputs(ctx); looksLikeShareURL(v) {
			return v
		}
		if exists(ctx, copyPublicLinkXPath, byXPath) {
			_ = clickRobust(ctx, copyPublicLinkXPath, byXPath)
		} else {
			_ = deepClickButtonByText(ctx, "Copy public link")
		}
		if delay > 0 {
			_ = chromedp.Run(ctx, chromedp.Sleep(delay))
		}
	}
	return ""
}

func deepScanShareInput(ctx context.Context) string {
	js := `(function(){
		const seen = new Set();
		const looks = v => typeof v==='string' && v.startsWith('http') && v.includes('/estimate');
		function dfs(root){
			if(!root || seen.has(root)) return '';
			seen.add(root);
			const ins = root.querySelectorAll ? root.querySelectorAll('input, textarea') : [];
			for(const el of ins){
				const v = (el.value||'').trim();
				if(looks(v)) return v;
			}
			const as = root.querySelectorAll ? root.querySelectorAll('a') : [];
			for (const a of as){
				const href = (a.href||'').trim();
				if(looks(href)) return href;
			}
			const nodes = root.querySelectorAll ? root.querySelectorAll('*') : [];
			for(const el of nodes){
				if(el.shadowRoot){
					const r = dfs(el.shadowRoot);
					if(r) return r;
				}
				const t = (el.innerText||el.textContent||'').trim();
				if(looks(t)) return t;
			}
			return '';
		}
		return dfs(document);
	})()`
	var val string
	_ = chromedp.Run(ctx, chromedp.Evaluate(js, &val))
	return strings.TrimSpace(val)
}

func deepClickButtonByText(ctx context.Context, label string) bool {
	js := fmt.Sprintf(`(function(){
		const target = %q.toLowerCase();
		function text(el){ return (el.innerText||el.textContent||'').trim().toLowerCase(); }
		function clickIfMatch(el){
			const t = text(el);
			if(t===target || t.includes(target)) { el.click(); return true; }
			return false;
		}
		function dfs(root){
			if(!root) return false;
			let list = [];
			if(root.querySelectorAll){
				list = Array.from(root.querySelectorAll('button,[role="button"]'));
			}
			for(const el of list){
				if(clickIfMatch(el)) return true;
			}
			if(root.querySelectorAll){
				for(const el of Array.from(root.querySelectorAll('*'))){
					if(el.shadowRoot && dfs(el.shadowRoot)) return true;
				}
			}
			return false;
		}
		return dfs(document);
	})()`, label)
	var ok bool
	_ = chromedp.Run(ctx, chromedp.Evaluate(js, &ok))
	if ok {
		log.Printf("   clicked(deep): %s", label)
	}
	return ok
}

// ---- Robust click helper ----

type queryBy int

const (
	byCSS queryBy = iota
	byXPath
)

type selector struct {
	s  string
	by queryBy
}

func clickRobust(ctx context.Context, sel string, by queryBy) error {
	opts := queryOpts(by)

	var nodes []*cdp.Node
	if err := chromedp.Run(ctx, chromedp.Nodes(sel, &nodes, opts...)); err != nil || len(nodes) == 0 {
		if by == byXPath && strings.Contains(strings.ToLower(sel), "agree and continue") {
			if deepClickButtonByText(ctx, "Agree and continue") {
				return nil
			}
		}
		return fmt.Errorf("not found: %s", sel)
	}

	if err := chromedp.Run(ctx,
		chromedp.ScrollIntoView(sel, opts...),
		chromedp.WaitVisible(sel, opts...),
		chromedp.Click(sel, opts...),
	); err == nil {
		log.Printf("   clicked: %s", shortSel(sel))
		return nil
	}

	if err := chromedp.Run(ctx, chromedp.MouseClickNode(nodes[0])); err == nil {
		log.Printf("   clicked(mouse): %s", shortSel(sel))
		return nil
	}

	var js string
	switch by {
	case byCSS:
		js = fmt.Sprintf(`(function(){var e=document.querySelector(%q); if(!e) return "no-el"; e.click(); return "ok";})()`, sel)
	default:
		js = fmt.Sprintf(`(function(){var it=document.evaluate(%q,document,null,XPathResult.FIRST_ORDERED_NODE_TYPE,null); var e=it.singleNodeValue; if(!e) return "no-el"; e.click(); return "ok";})()`, sel)
	}
	var res string
	if err := chromedp.Run(ctx, chromedp.Evaluate(js, &res)); err == nil && res == "ok" {
		log.Printf("   clicked(js): %s", shortSel(sel))
		return nil
	}

	_ = chromedp.Run(ctx, chromedp.Focus(sel, opts...), chromedp.SendKeys(sel, kb.Enter, opts...))
	_ = chromedp.Run(ctx, chromedp.Sleep(120*time.Millisecond))

	return fmt.Errorf("click failed: %s", sel)
}

// ---- Generic helpers ----

func buildURL(base, region, hash string) string {
	if strings.TrimSpace(region) == "" {
		return base + "/" + hash
	}
	return fmt.Sprintf("%s/?region=%s%s", base, url.QueryEscape(region), hash)
}

func currentURL(ctx context.Context) string {
	var loc string
	_ = chromedp.Run(ctx, chromedp.Location(&loc))
	return loc
}

func clickAny(ctx context.Context, sels []selector) bool {
	for _, sel := range sels {
		if clickIfExists(ctx, sel.s, sel.by) {
			return true
		}
	}
	return false
}

func clickIfExists(ctx context.Context, sel string, by queryBy) bool {
	var nodes []*cdp.Node
	opts := queryOpts(by)
	if err := chromedp.Run(ctx, chromedp.Nodes(sel, &nodes, opts...)); err != nil || len(nodes) == 0 {
		return false
	}
	_ = chromedp.Run(ctx, chromedp.ScrollIntoView(sel, opts...))
	if err := chromedp.Run(ctx, chromedp.Click(sel, opts...)); err == nil {
		log.Printf("   clicked: %s", shortSel(sel))
		return true
	}
	return false
}

func exists(ctx context.Context, sel string, by queryBy) bool {
	var nodes []*cdp.Node
	opts := queryOpts(by)
	if err := chromedp.Run(ctx, chromedp.Nodes(sel, &nodes, opts...)); err != nil {
		return false
	}
	return len(nodes) > 0
}

func typeInto(ctx context.Context, sel string, by queryBy, text string) error {
	opts := queryOpts(by)
	_ = chromedp.Run(ctx,
		chromedp.Focus(sel, opts...),
		chromedp.Click(sel, opts...),
		chromedp.SendKeys(sel, kb.Control+"a", opts...),
		chromedp.SendKeys(sel, kb.Backspace, opts...),
	)
	_ = setInputValueJS(ctx, sel, by, "")
	return chromedp.Run(ctx,
		chromedp.SendKeys(sel, text, opts...),
		chromedp.SendKeys(sel, kb.Enter, opts...),
	)
}

func setInputValueJS(ctx context.Context, sel string, by queryBy, val string) error {
	var expr string
	switch by {
	case byCSS:
		expr = fmt.Sprintf(`(function(){
			const el = document.querySelector(%q);
			if(!el) return "no-el";
			const proto = Object.getPrototypeOf(el);
			const desc = Object.getOwnPropertyDescriptor(proto, 'value') || Object.getOwnPropertyDescriptor(HTMLInputElement.prototype,'value');
			if(desc && desc.set) desc.set.call(el, %q); else el.value=%q;
			el.dispatchEvent(new Event('input', {bubbles:true}));
			el.dispatchEvent(new Event('change', {bubbles:true}));
			return "ok";
		})()`, sel, val, val)
	case byXPath:
		expr = fmt.Sprintf(`(function(){
			const it = document.evaluate(%q, document, null, XPathResult.FIRST_ORDERED_NODE_TYPE, null);
			const el = it.singleNodeValue;
			if(!el) return "no-el";
			const proto = Object.getPrototypeOf(el);
			const desc = Object.getOwnPropertyDescriptor(proto, 'value') || Object.getOwnPropertyDescriptor(HTMLInputElement.prototype,'value');
			if(desc && desc.set) desc.set.call(el, %q); else el.value=%q;
			el.dispatchEvent(new Event('input', {bubbles:true}));
			el.dispatchEvent(new Event('change', {bubbles:true}));
			return "ok";
		})()`, sel, val, val)
	}
	var res string
	if err := chromedp.Run(ctx, chromedp.Evaluate(expr, &res)); err != nil {
		return err
	}
	if res != "ok" {
		return fmt.Errorf("js set value failed: %s", res)
	}
	return nil
}

func readInputValue(ctx context.Context, sel string, by queryBy) (string, bool) {
	var js string
	switch by {
	case byCSS:
		js = fmt.Sprintf(`(function(){const el=document.querySelector(%q); return el ? (el.value||"") : "";})()`, sel)
	default:
		js = fmt.Sprintf(`(function(){const it=document.evaluate(%q,document,null,XPathResult.FIRST_ORDERED_NODE_TYPE,null); const el=it.singleNodeValue; return el ? (el.value||"") : "";})()`, sel)
	}
	var val string
	_ = chromedp.Run(ctx, chromedp.Evaluate(js, &val))
	val = strings.TrimSpace(val)
	if val != "" {
		return val, true
	}
	return "", false
}

func scanAnyShareInput(ctx context.Context) string {
	c2, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	js := `(function(){
  		const looks = v => typeof v==='string' && v.startsWith('http') && v.includes('/estimate');
  		const nodes = Array.from(document.querySelectorAll('input,textarea,a'));
  		for (const el of nodes) {
    		const v = ((el.value !== undefined ? el.value : '') || (el.href || '')).trim();
    		if (looks(v)) return v;
  		}
  		return '';
	})()`
	var val string
	_ = chromedp.Run(c2, chromedp.Evaluate(js, &val))
	return strings.TrimSpace(val)
}

func looksLikeShareURL(v string) bool {
	v = strings.TrimSpace(html.UnescapeString(v))
	if v == "" || !strings.HasPrefix(v, "http") {
		return false
	}
	if !strings.Contains(v, "calculator.aws") {
		return false
	}
	return strings.Contains(v, "/estimate")
}

func fullHTML(ctx context.Context) (string, error) {
	c2, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var htmlStr string
	if err := chromedp.Run(c2, chromedp.Evaluate(`document.documentElement.outerHTML`, &htmlStr)); err == nil && strings.TrimSpace(htmlStr) != "" {
		return htmlStr, nil
	}
	htmlStr = ""
	if err := chromedp.Run(c2, chromedp.OuterHTML("html", &htmlStr, chromedp.ByQuery)); err != nil {
		return "", err
	}
	if strings.TrimSpace(htmlStr) == "" {
		return "", fmt.Errorf("empty html")
	}
	return htmlStr, nil
}

func dumpHTML(ctx context.Context, note string) {
	if htmlStr, err := fullHTML(ctx); err == nil {
		_ = os.WriteFile("tmp.html", []byte(htmlStr), 0644)
		if note != "" {
			log.Printf("        Wrote HTML snapshot to tmp.html %s", note)
		} else {
			log.Printf("        Wrote HTML snapshot to tmp.html")
		}
	} else {
		if note != "" {
			log.Printf("        Could not capture HTML snapshot %s: %v", note, err)
		} else {
			log.Printf("        Could not capture HTML snapshot: %v", err)
		}
	}
}

func dismissCookieBanner(ctx context.Context) {
	js := `(function(){
		var root = document.querySelector('#awsccc-cb-c');
		if (!root) return 'none';
		var btn = root.querySelector('button.awsccc-u-btn-primary, button[title="Agree and continue"], button[aria-label="Agree and continue"]');
		if (btn) { btn.click(); return 'clicked'; }
		root.style.display = 'none';
		return 'hidden';
	})()`
	var res string
	_ = chromedp.Run(ctx, chromedp.Evaluate(js, &res))
	if res != "" {
		log.Printf("        Cookie banner: %s", res)
	}
}

func scrollToBottom(ctx context.Context) error {
	return chromedp.Run(ctx, chromedp.Evaluate(`window.scrollTo({top: document.body.scrollHeight, behavior: 'instant'});`, nil))
}

func scrollToTop(ctx context.Context) error {
	return chromedp.Run(ctx, chromedp.Evaluate(`window.scrollTo({top: 0, behavior: 'instant'});`, nil))
}

func queryOpts(by queryBy) []chromedp.QueryOption {
	switch by {
	case byXPath:
		return []chromedp.QueryOption{chromedp.BySearch}
	default:
		return []chromedp.QueryOption{chromedp.ByQuery}
	}
}

func shortSel(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > 96 {
		return s[:96] + "…"
	}
	return s
}

func dumpAndExtract(ctx context.Context, note string) string {
	htmlStr, err := fullHTML(ctx)
	if err == nil {
		_ = os.WriteFile("tmp.html", []byte(htmlStr), 0644)
		if note != "" {
			log.Printf("        Wrote HTML snapshot to tmp.html %s", note)
		} else {
			log.Printf("        Wrote HTML snapshot to tmp.html")
		}
		if v := extractShareURLFromHTML(htmlStr); looksLikeShareURL(v) {
			return html.UnescapeString(strings.TrimSpace(v))
		}
	} else {
		if note != "" {
			log.Printf("        Could not capture HTML snapshot %s: %v", note, err)
		} else {
			log.Printf("        Could not capture HTML snapshot: %v", err)
		}
	}
	return ""
}

func extractShareURLFromHTML(s string) string {
	re := regexp.MustCompile(`https://calculator\.aws/[^\s"'<>]+`)
	m := re.FindString(s)
	return m
}

func ensureOnDemand(ctx context.Context, d time.Duration) error {
	if err := clickWithTimeout(ctx, onDemandOptionXPath, byXPath, d); err != nil {
		_ = deepClickButtonByText(ctx, "On-Demand")
	}
	return nil
}

// (REPOSTO) Garante que filtros "Any Memory" e "Any vCPUs" sejam ativados
func ensureAnyFilters(ctx context.Context, d time.Duration) error {
	_ = clickWithTimeout(ctx, anyMemoryTriggerXPath, byXPath, d)
	_ = chromedp.Run(ctx, chromedp.Sleep(100*time.Millisecond))
	_ = clickWithTimeout(ctx, anyVcpuTriggerXPath, byXPath, d)
	_ = chromedp.Run(ctx, chromedp.Sleep(100*time.Millisecond))
	return nil
}

// ---- Save/Add helper ----

func clickSaveAndAddService(ctx context.Context) error {
	for i := 1; i <= 3; i++ {
		log.Printf("        [save/add] attempt %d/3", i)
		_ = scrollToBottom(ctx)
		_ = chromedp.Run(ctx, chromedp.WaitVisible(appFooterCSS, chromedp.ByQuery))
		if err := clickWithTimeout(ctx, saveAndAddBtnFooterCSS, byCSS, 5*time.Second); err != nil {
			_ = clickWithTimeout(ctx, saveAndAddXPath, byXPath, 5*time.Second)
		}
		_ = chromedp.Run(ctx, chromedp.Sleep(500*time.Millisecond))
		if err := waitVisibleWithTimeout(ctx, findServiceInputCSS, byCSS, 5*time.Second); err == nil {
			log.Printf("        [save/add] service added (Find Service visible)")
			return nil
		}
		dumpHTML(ctx, fmt.Sprintf("(after save/add attempt %d)", i))
	}
	return fmt.Errorf("Save and add service did not complete")
}

// ---- Planning (greedy desc, ignorando *.nano) ----

type ec2Option struct {
	Name    string
	Hourly  float64
	Monthly float64
}

type planItem struct {
	Name    string
	Hourly  float64
	Monthly float64
	Count   int
}

// YAML model
type pricingDoc struct {
	EC2 []struct {
		Name      string  `yaml:"name"`
		Hourly    float64 `yaml:"hourly,omitempty"`
		Monthly   float64 `yaml:"monthly,omitempty"`
		VCpus     int     `yaml:"vcpus,omitempty"`
		MemoryGiB string  `yaml:"memory_gib,omitempty"`
	} `yaml:"ec2"`
}

// Lê EC2_PRICING_YAML ou ./pricing.yaml. Filtra *.nano
func ec2Catalog() []ec2Option {
	path := os.Getenv("EC2_PRICING_YAML")
	if strings.TrimSpace(path) == "" {
		path = "pricing.yaml"
	}
	b, err := os.ReadFile(path)
	if err != nil {
		log.Printf("        could not read pricing file %q: %v", path, err)
		return nil
	}
	opts := parsePricingYAML(b)
	if len(opts) == 0 {
		log.Printf("        WARNING: no EC2 entries loaded from %s", path)
	} else {
		log.Printf("        loaded %d pricing entries from %s (nano filtered out)", len(opts), path)
	}
	return opts
}

func parsePricingYAML(b []byte) []ec2Option {
	const HPM = 730.0
	var doc pricingDoc
	if err := yaml.Unmarshal(b, &doc); err != nil {
		log.Printf("        YAML unmarshal failed: %v", err)
		return nil
	}
	m := map[string]ec2Option{}
	for _, it := range doc.EC2 {
		name := strings.TrimSpace(it.Name)
		if name == "" {
			continue
		}
		if strings.Contains(name, ".nano") {
			continue
		}
		hr := it.Hourly
		mo := it.Monthly
		if mo <= 0 && hr > 0 {
			mo = hr * HPM
		}
		if hr <= 0 && mo > 0 {
			hr = mo / HPM
		}
		if mo <= 0 {
			continue
		}
		if old, ok := m[name]; !ok || mo < old.Monthly {
			m[name] = ec2Option{Name: name, Hourly: hr, Monthly: mo}
		}
	}
	opts := make([]ec2Option, 0, len(m))
	for _, v := range m {
		opts = append(opts, v)
	}
	sort.Slice(opts, func(i, j int) bool { return opts[i].Monthly < opts[j].Monthly })
	return opts
}

// Heurística solicitada (descendente, while current+price < target)
func planGreedyEC2(targetMRR float64, _ float64) []planItem {
	opts := ec2Catalog()
	if len(opts) == 0 || targetMRR <= 0 {
		return nil
	}

	cp := append([]ec2Option(nil), opts...)
	sort.Slice(cp, func(i, j int) bool { return cp[i].Monthly > cp[j].Monthly })

	var plan []planItem
	current := 0.0
	for _, o := range cp {
		if o.Monthly <= 0 {
			continue
		}
		count := 0
		for (current + o.Monthly) < targetMRR {
			current += o.Monthly
			count++
		}
		if count > 0 {
			plan = append(plan, planItem{
				Name:    o.Name,
				Hourly:  o.Hourly,
				Monthly: o.Monthly,
				Count:   count,
			})
		}
		if current >= targetMRR {
			break
		}
	}
	return compactPlan(plan)
}

func compactPlan(in []planItem) []planItem {
	if len(in) == 0 {
		return in
	}
	m := map[string]*planItem{}
	for _, it := range in {
		if it.Count <= 0 {
			continue
		}
		if ex, ok := m[it.Name]; ok {
			ex.Count += it.Count
		} else {
			cp := it
			m[it.Name] = &cp
		}
	}
	out := make([]planItem, 0, len(m))
	for _, v := range m {
		out = append(out, *v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Monthly > out[j].Monthly })
	return out
}

// ---- UI helpers ----

func selectInstanceByName(ctx context.Context, instance string, timeout time.Duration) error {
	if exists(ctx, instanceSearchInputCSS, byCSS) {
		_ = typeInto(ctx, instanceSearchInputCSS, byCSS, instance)
		_ = chromedp.Run(ctx, chromedp.Sleep(200*time.Millisecond))
	}
	x := fmt.Sprintf(`//label[.//text()[contains(., '%s')]] | //span[contains(normalize-space(),'%s')]/ancestor::label | //tr[.//*[contains(normalize-space(),'%s')]]//label`, instance, instance, instance)
	return clickWithTimeout(ctx, x, byXPath, timeout)
}
func setInstanceCount(ctx context.Context, count int) error {
	if count < 1 {
		count = 1
	}
	val := fmt.Sprintf("%d", count)

	// 1) Mesmo alvo do seu Puppeteer
	if err := setInputValueJS(ctx, numberInstancesInputCSSExact, byCSS, val); err == nil {
		if v, ok := readInputValue(ctx, numberInstancesInputCSSExact, byCSS); ok && strings.TrimSpace(v) == val {
			log.Printf("        [count] set via Puppeteer-exact aria-label → %s", val)
			return nil
		}
	}

	// 2) Fallback: nosso XPath já existente para "Number of instances"
	if exists(ctx, numberInstancesInputXPath, byXPath) {
		if err := setInputValueJS(ctx, numberInstancesInputXPath, byXPath, val); err == nil {
			if v, ok := readInputValue(ctx, numberInstancesInputXPath, byXPath); ok && strings.TrimSpace(v) == val {
				log.Printf("        [count] set via XPath(aria-label contains 'Number of instances') → %s", val)
				return nil
			}
		}
	}

	return fmt.Errorf("could not set instance count to %s (no matching input)", val)
}
