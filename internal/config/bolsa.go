package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

var (
	// Bolsa Settings
	BolsaUsername      string
	BolsaPassword      string
	BetAmount          int
	TargetLeagueName   string
	TargetIframeDomain string
)

func loadBolsaConfig() error {
	BolsaUsername = os.Getenv("BOLSA_USERNAME")
	BolsaPassword = os.Getenv("BOLSA_PASSWORD")
	TargetLeagueName = os.Getenv("TARGET_LEAGUE_NAME")
	if TargetLeagueName == "" {
		TargetLeagueName = "GT Leagues"
	}

	TargetIframeDomain = os.Getenv("TARGET_IFRAME_DOMAIN")
	if TargetIframeDomain == "" {
		TargetIframeDomain = "fssb.io"
	}

	amountStr := os.Getenv("BET_AMOUNT")
	if amountStr == "" {
		return fmt.Errorf("BET_AMOUNT não configurado")
	}
	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		return fmt.Errorf("BET_AMOUNT inválido: %w", err)
	}
	BetAmount = amount

	return nil
}

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
	SelectorAHButton          = "button.eventpage_fe_HandicapSelection_line"
	SelectorRealityCheckBtn   = "app-reality-check-dialog button.btn--color--transparent"
	
	// XPaths
	XPathLeagueGT        = `//*[contains(translate(text(), 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz'), 'gt leagues')]`
	XPathHandicapMarket  = `//*[contains(@class, "MarketHeader_marketName") or contains(@class, "eventpage_fe_Markets_marketName") or contains(@class, "MarketName") or contains(@class, "header")][contains(text(), "Asian") or contains(text(), "Asiático")]/ancestor::div[contains(@class, "MarketGroup_wrapper") or contains(@class, "MarketGroup") or contains(@class, "marketsList_marketGroup")]`
	XPathGoalMarket      = `//*[contains(@class, "MarketHeader_marketName") or contains(@class, "eventpage_fe_Markets_marketName") or contains(@class, "MarketName") or contains(@class, "header")][contains(text(), "Aposta Ao Vivo Mais/Menos") or contains(translate(text(), 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz'), "mais/menos")]/ancestor::div[contains(@class, "MarketGroup_wrapper") or contains(@class, "MarketGroup") or contains(@class, "marketsList_marketGroup")]`
)

var (
	TimeoutModal    = 10 * time.Second
	TimeoutSearch   = 30 * time.Second
	DelayAction     = 1 * time.Second
	DelayNavigation = 3 * time.Second
)
