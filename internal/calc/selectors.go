package calc

const (
	SearchInputCSS     = "input[placeholder*='Search' i],input[placeholder*='Find' i],input[aria-label*='search' i],input[aria-label*='find' i],input[type='search']"
	EC2ConfigureXPath  = "//div[.//h3[contains(., 'Amazon EC2')]]//button[contains(., 'Configure') or .//span[contains(., 'Configure')]]"
	SaveAndAddXPath    = "//button[contains(., 'Save and add service') or .//span[contains(., 'Save and add service')]]"
	EditNameLinkCSS    = "a[data-cy='edit-estimate-name']"
	EstimateNameInput  = "input[aria-label='Enter Name']"
	SaveNameButton     = "button[aria-label='Save']"
	ShareButtonXPath   = "//button[contains(., 'Share') or .//span[contains(., 'Share')]]"
	ShareLinkXPath     = "//input[contains(@value, 'https://calculator.aws/#/estimate/')]"
	NumberInstancesCSS = "input[aria-label*='Number of instances' i]"
)
