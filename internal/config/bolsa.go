package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

var (
	TargetLeagueName   = "gt leagues"
	TargetIframeDomain = "fssb.io"
	BolsaUsername      = ""
	BolsaPassword      = ""
	StakeAmount        int
)

func loadBolsaConfig() {
	if val := os.Getenv("TARGET_LEAGUE_NAME"); val != "" {
		TargetLeagueName = val
	}
	if val := os.Getenv("TARGET_IFRAME_DOMAIN"); val != "" {
		TargetIframeDomain = val
	}
	BolsaUsername = os.Getenv("BOLSA_USERNAME")
	BolsaPassword = os.Getenv("BOLSA_PASSWORD")

	val := os.Getenv("BET_AMOUNT")
	if val == "" {
		log.Fatalf("Erro: BET_AMOUNT é obrigatório no arquivo .env")
	}
	amt, err := strconv.Atoi(val)
	if err != nil {
		log.Fatalf("Erro: BET_AMOUNT no .env deve ser um valor inteiro, recebido: %s", val)
	}
	StakeAmount = amt
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
	SelectorBetslipTab        = "div.navigation_eu_fe_Breadcrumbs_betslipTab"
	SelectorAHButton          = "button.eventpage_fe_HandicapSelection_line"
	SelectorRealityCheckBtn   = "app-reality-check-dialog button.btn--color--transparent"
	
	// XPaths
	XPathLeagueGT        = `//*[contains(translate(text(), 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz'), 'gt leagues')]`
	XPathHandicapMarket  = `//*[contains(@class, "MarketHeader_marketName") or contains(@class, "eventpage_fe_Markets_marketName") or contains(@class, "MarketName") or contains(@class, "header")][contains(text(), "Asian") or contains(text(), "Asiático")]/ancestor::div[contains(@class, "MarketGroup_wrapper") or contains(@class, "MarketGroup") or contains(@class, "marketsList_marketGroup")]`
	XPathGoalMarket      = `//*[contains(@class, "MarketHeader_marketName") or contains(@class, "eventpage_fe_Markets_marketName") or contains(@class, "MarketName") or contains(@class, "header")][contains(text(), "Aposta Ao Vivo Mais/Menos") or contains(translate(text(), 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz'), "mais/menos")]/ancestor::div[contains(@class, "MarketGroup_wrapper") or contains(@class, "MarketGroup") or contains(@class, "marketsList_marketGroup")]`
	
	// Timeouts and Delays
	TimeoutModal    = 10 * time.Second
	TimeoutSearch   = 30 * time.Second
	DelayAction     = 1 * time.Second
	DelayNavigation = 3 * time.Second
)
