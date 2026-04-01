package config

import "time"

const (
	BaseURL = "https://bolsadeaposta.bet.br/fbook/br-pt/spbk?selectedDefaultTab=Live&selectedLiveSport=1"
	
	// Selectors
	SelectorUsernameInput = `input[placeholder="Username"], input[type="text"], input[type="email"], input[placeholder*="Usuário"], input[placeholder*="Email"]`
	SelectorPasswordInput = `input[placeholder="Password"], input[type="password"]`
	SelectorUserActions   = `.useractions .user-name strong`
	SelectorAgeYesBtn     = `#cdk-overlay-0 button`
	SelectorCookieBtn     = `div[id^="cdk-overlay"] button`
	SelectorHandicapBtns  = ".eventpage_fe_HandicapSelection_line, button[class*='HandicapSelection_line']"
	SelectorHandicapPoints = ".eventpage_fe_HandicapSelection_points, [class*='HandicapSelection_points']"
	SelectorHandicapOdds   = ".eventpage_fe_HandicapSelection_odds, [class*='HandicapSelection_odds']"
	SelectorBackBtn       = `//div[contains(@class, "eventdetails")]//div[contains(@class, "backButton")] | //div[contains(@class, "MarketHeader_back")] | //div[contains(@class, "navigation_eu_fe_Breadcrumbs_backButton")]`
	
	// Betting Selectors
	SelectorIframeFSSB       = "iframe[src*='fssb.io']"
	SelectorBetslipStakeInput = "input#counter"
	SelectorBetslipPlaceBetBtn = "button#place-bets"
	SelectorBetslipRemoveBtn  = "button.betslip_fe_BetInformationSecondary_betInformation__removeButton"
	SelectorBetslipTab        = "div.navigation_eu_fe_Breadcrumbs_betslipTab"
	SelectorAHButton          = "button.eventpage_fe_HandicapSelection_line"
	
	// XPaths
	XPathLeagueGT        = `//*[contains(translate(text(), 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz'), 'gt leagues')]`
	XPathHandicapMarket  = `//*[contains(@class, "MarketHeader_marketName") or contains(@class, "eventpage_fe_Markets_marketName") or contains(@class, "MarketName") or contains(@class, "header")][contains(text(), "Asian") or contains(text(), "Asiático")]/ancestor::div[contains(@class, "MarketGroup_wrapper") or contains(@class, "MarketGroup") or contains(@class, "marketsList_marketGroup")]`
	
	// Timeouts and Delays
	TimeoutModal    = 10 * time.Second
	TimeoutSearch   = 30 * time.Second
	DelayAction     = 1 * time.Second
	DelayNavigation = 3 * time.Second

	// Default Stake
	DefaultStake = "5"
)
