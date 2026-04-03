package crawler

import (
	"bolsadeaposta-bot/internal/config"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// NavigateToMatch Dedicates a new browser tab to find and enter a specific match.
// It returns the *rod.Page focused on the match details.
func NavigateToMatch(browser *rod.Browser, player1, player2 string) (*rod.Page, error) {
	page, err := browser.Page(proto.TargetCreateTarget{URL: config.BaseURL})
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir aba para o jogo: %w", err)
	}

	if err := page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width:             1920,
		Height:            1080,
		DeviceScaleFactor: 1,
		Mobile:            false,
	}); err != nil {
		page.Close()
		return nil, err
	}

	page.MustWaitLoad()

	league, err := FindLeague(page)
	if err != nil {
		page.Close()
		return nil, fmt.Errorf("liga não encontrada na nova aba: %w", err)
	}

	if err := ExpandLeague(league); err != nil {
		page.Close()
		return nil, fmt.Errorf("erro ao expandir a liga: %w", err)
	}

	// We look for the specific match
	container, err := league.ElementX(`ancestor::div[contains(@class,"header") or contains(@class,"Header")]/parent::div`)
	if err != nil {
		page.Close()
		return nil, fmt.Errorf("container da liga não localizado")
	}

	var events rod.Elements
	for i := 0; i < 15; i++ {
		eventElements, err := container.Elements(`.eventlist_eu_fe_EventItemDesktop_wrapper, [class*="EventItemDesktop_wrapper"], [class*="EventItem"]`)
		if err == nil && len(eventElements) > 0 {
			events = eventElements
			break
		}
		time.Sleep(1 * time.Second)
	}

	var targetMatch *rod.Element
	var team1Name, team2Name string

	for _, event := range events {
		teams, _ := event.Elements(`[class*="teamNameText"], [class*="participantName"]`)
		if len(teams) < 2 {
			continue
		}

		extractedTeam1, _ := teams[0].Text()
		extractedTeam2, _ := teams[1].Text()
		
		matched, t1, t2 := IsMatchTarget(extractedTeam1, extractedTeam2, player1, player2)
		if matched {
			targetMatch = event
			team1Name = t1
			team2Name = t2
			break
		}
	}

	if targetMatch == nil {
		page.Close()
		return nil, fmt.Errorf("partida %s vs %s não encontrada ativamente", player1, player2)
	}

	log.Printf("🎯 Instando acompanhamento Multi-Aba: %s vs %s", team1Name, team2Name)

	teams, _ := targetMatch.Elements(`[class*="teamNameText"], [class*="participantName"]`)
	clickTarget, err := teams[0].ElementX(`ancestor::div[contains(@class, "participant") or contains(@class, "Team")][1]`)
	if err != nil {
		clickTarget = teams[0]
	}

	_ = clickTarget.ScrollIntoView()
	_ = clickTarget.Click(proto.InputMouseButtonLeft, 1)

	// Aguarda o container de mercados carregar
	_, _ = page.Timeout(10 * time.Second).Element(".eventdetails_eu_fe_ViewStyles_scrollContainer, .market_fe_MarketList_wrapper, .MarketList_wrapper")

	time.Sleep(2 * time.Second) // additional buffer to let components render fully
	return page, nil
}

// FetchLiveScore extracts the live score from the current match details page
func FetchLiveScore(page *rod.Page) (string, error) {
	// Inside match details, score usually has different classes, e.g. "Score_wrapper", "MatchScore". Look for common score blocks
	scoreEl, err := page.Element(`[class*="ScoreWrapper"], [class*="MatchScore"], [class*="score-live"], [class*="score"]`)
	if err != nil {
		// Try fallback to just typical number-number pattern if needed.
		return "", err
	}
	text, _ := scoreEl.Text()
	
	return ParseLiveScore(text), nil
}

// ParseLiveScore converts raw score text into a "S1-S2" format.
func ParseLiveScore(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}

	// Format could be "1\n2" or "1 - 2" or "1:2"
	lines := strings.Split(text, "\n")
	var validNumbers []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			validNumbers = append(validNumbers, l)
		}
	}

	if len(validNumbers) >= 2 {
		return fmt.Sprintf("%s-%s", validNumbers[0], validNumbers[1])
	}

	return text
}
