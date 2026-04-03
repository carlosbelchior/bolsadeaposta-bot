package crawler

import (
	"bolsadeaposta-bot/internal/config"
	"bolsadeaposta-bot/internal/models"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

func FindLeague(page *rod.Page) (*rod.Element, error) {
	log.Printf("⏳ Buscando pela liga '%s'...", config.TargetLeagueName)
	start := time.Now()

	for time.Since(start) < config.TimeoutSearch {
		// 🔹 Procura na página principal
		targetLeagues, err := page.ElementsX(config.XPathLeagueGT)
		if err == nil {
			for _, league := range targetLeagues {
				text, _ := league.Text()
				if strings.Contains(strings.ToLower(text), strings.ToLower(config.TargetLeagueName)) {
					log.Printf("✅ Liga '%s' encontrada na página principal.", config.TargetLeagueName)
					_ = league.ScrollIntoView()
					return league, nil
				}
			}
		}

		// 🔹 Procura nos iframes
		iframes, _ := page.Elements("iframe")
		for _, iframe := range iframes {
			src, _ := iframe.Attribute("src")
			if src != nil && strings.Contains(*src, config.TargetIframeDomain) {
				f, err := iframe.Frame()
				if err == nil {
					targetLeaguesInFrame, _ := f.ElementsX(config.XPathLeagueGT)
					for _, league := range targetLeaguesInFrame {
						text, _ := league.Text()
						if strings.Contains(strings.ToLower(text), strings.ToLower(config.TargetLeagueName)) {
							log.Printf("✅ Liga '%s' encontrada dentro de iframe.", config.TargetLeagueName)
							_ = league.ScrollIntoView()
							return league, nil
						}
					}
				}
			}
		}

		// Scroll para forçar carregamento
		_, _ = page.Eval(`() => {
			let scrollContainer = document.querySelector('.master_fe_ViewStyles_desktopWrapperEventListIndependentScroll, .ViewStyles_desktopWrapperEventListIndependentScroll, [class*="EventListIndependentScroll"]');
			if (scrollContainer) {
				scrollContainer.scrollBy(0, 800);
			} else {
				window.scrollBy(0, 800);
			}
		}`)

		allLeagues, _ := page.Elements(`[class*="LeagueItemDesktop_header"], [class*="leagueHeader"], [class*="LeagueHeader"], [class*="LeagueItem"]`)
		if len(allLeagues) > 0 {
			_ = allLeagues[len(allLeagues)-1].ScrollIntoView()
		}

		log.Printf("⏳ Procurando '%s'... (%v decorridos)", config.TargetLeagueName, time.Since(start).Round(time.Second))
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("liga '%s' não encontrada após %v", config.TargetLeagueName, config.TimeoutSearch)
}

func ExpandLeague(league *rod.Element) error {
	log.Println("⏳ Tentando expandir a liga...")
	xpath := `ancestor::div[contains(@class,"header") or contains(@class,"Header")]//button | ancestor::div[contains(@class,"header") or contains(@class,"Header")]//i[contains(@class, "arrow") or contains(@class, "Arrow")] | .`
	btn, err := league.ElementX(xpath)
	if err != nil {
		log.Println("⚠️ Botão da liga não encontrado explicitamente, tentando clicar no elemento da liga.")
		btn = league
	}

	if err := btn.ScrollIntoView(); err != nil {
		return fmt.Errorf("não foi possível dar scroll para a liga: %w", err)
	}
	if err := btn.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("não foi possível clicar na liga: %w", err)
	}

	log.Println("✅ Clique na liga realizado.")
	time.Sleep(config.DelayNavigation)

	// Dá um scroll para baixo para exibir os jogos
	_, _ = league.Page().Eval(`() => {
			let scrollContainer = document.querySelector('.master_fe_ViewStyles_desktopWrapperEventListIndependentScroll, .ViewStyles_desktopWrapperEventListIndependentScroll, [class*="EventListIndependentScroll"]');
			if (scrollContainer) {
				scrollContainer.scrollBy(0, 400);
			} else {
				window.scrollBy(0, 400);
			}
		}`)
	log.Println("📜 Rolando para exibir partidas...")
	return nil
}

func GetMatches(league *rod.Element, player1, player2 string) []models.Match {
	var matches []models.Match

	container, err := league.ElementX(`ancestor::div[contains(@class,"header") or contains(@class,"Header")]/parent::div`)
	if err != nil {
		log.Println("⚠️ Não foi possível localizar o container da liga:", err)
		return matches
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

	seen := make(map[string]bool)

	for _, event := range events {
		teams, _ := event.Elements(`[class*="teamNameText"], [class*="participantName"]`)
		if len(teams) < 2 {
			continue
		}

		teamOneName, _ := teams[0].Text()
		teamTwoName, _ := teams[1].Text()

		matched, t1, t2 := IsMatchTarget(teamOneName, teamTwoName, player1, player2)
		if !matched {
			continue
		}
		teamOneName = t1
		teamTwoName = t2

		key := teamOneName + "|" + teamTwoName
		if seen[key] {
			continue
		}
		seen[key] = true

		scores, _ := event.Elements(`[class*="mainScore"]`)
		timeText := ""
		timeEl, err := event.Element(`[class*="LiveEventCounter_wrapper"], [class*="Time"]`)
		if err == nil {
			timeText, _ = timeEl.Text()
		}

		oddsEls, _ := event.Elements(`[class*="Selection_odds"], [class*="oddsText"]`)

		match := models.Match{
			HomeTeam: teamOneName,
			AwayTeam: teamTwoName,
		}

		if len(scores) >= 2 {
			match.HomeScore, _ = scores[0].Text()
			match.AwayScore, _ = scores[1].Text()
		}

		if len(oddsEls) >= 3 {
			match.HomeOdd, _ = oddsEls[0].Text()
			match.DrawOdd, _ = oddsEls[1].Text()
			match.AwayOdd, _ = oddsEls[2].Text()
		}

		timeText = strings.ReplaceAll(timeText, "\n", " ")
		match.Time = strings.TrimSpace(timeText)

		log.Printf("🔍 Entrando na partida: %s vs %s", teamOneName, teamTwoName)

		clickTarget, err := teams[0].ElementX(`ancestor::div[contains(@class, "participant") or contains(@class, "Team")][1]`)
		if err != nil {
			clickTarget = teams[0]
		}

		_ = clickTarget.ScrollIntoView()
		_ = clickTarget.Click(proto.InputMouseButtonLeft, 1)

		// Aguarda o container de mercados carregar
		_, _ = event.Page().Timeout(5 * time.Second).Element(".eventdetails_eu_fe_ViewStyles_scrollContainer, .market_fe_MarketList_wrapper, .MarketList_wrapper")

		// Volta para a lista de jogos
		backBtn, err := event.Page().ElementX(config.SelectorBackBtn)
		if err == nil {
			_ = backBtn.Click(proto.InputMouseButtonLeft, 1)
			time.Sleep(config.DelayAction)
		}

		matches = append(matches, match)
	}

	return matches
}

// IsMatchTarget checks if the extracted team names match the expected player/team names.
func IsMatchTarget(extractedTeam1, extractedTeam2, target1, target2 string) (bool, string, string) {
	t1 := strings.ToLower(strings.TrimSpace(extractedTeam1))
	t2 := strings.ToLower(strings.TrimSpace(extractedTeam2))
	p1 := strings.ToLower(target1)
	p2 := strings.ToLower(target2)

	match1 := strings.Contains(t1, p1) && strings.Contains(t2, p2)
	if match1 {
		return true, strings.TrimSpace(extractedTeam1), strings.TrimSpace(extractedTeam2)
	}
	match2 := strings.Contains(t1, p2) && strings.Contains(t2, p1)
	if match2 {
		return true, strings.TrimSpace(extractedTeam1), strings.TrimSpace(extractedTeam2)
	}
	return false, "", ""
}

// IsScoreMatch compares the current live score with the target score from the tip.
func IsScoreMatch(current, target string) bool {
	if target == "" {
		return true
	}
	c := strings.ReplaceAll(strings.ReplaceAll(current, " ", ""), "-", "")
	t := strings.ReplaceAll(strings.ReplaceAll(target, " ", ""), "-", "")
	return c == t
}
