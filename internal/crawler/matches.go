package crawler

import (
	"betstake-webscrap/internal/config"
	"betstake-webscrap/internal/models"
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

func FindLeague(page *rod.Page) (*rod.Element, error) {
	fmt.Println("⏳ Buscando pela liga 'GT Leagues'...")
	start := time.Now()

	for time.Since(start) < config.TimeoutSearch {
		// 🔹 Procura na página principal
		targetLeagues, err := page.ElementsX(config.XPathLeagueGT)
		if err == nil {
			for _, league := range targetLeagues {
				text, _ := league.Text()
				if strings.Contains(strings.ToLower(text), "gt leagues") {
					fmt.Println("✅ Liga 'GT Leagues' encontrada na página principal.")
					_ = league.ScrollIntoView()
					return league, nil
				}
			}
		}

		// 🔹 Procura nos iframes
		iframes, _ := page.Elements("iframe")
		for _, iframe := range iframes {
			src, _ := iframe.Attribute("src")
			if src != nil && strings.Contains(*src, "fssb.io") {
				f, err := iframe.Frame()
				if err == nil {
					targetLeaguesInFrame, _ := f.ElementsX(config.XPathLeagueGT)
					for _, league := range targetLeaguesInFrame {
						text, _ := league.Text()
						if strings.Contains(strings.ToLower(text), "gt leagues") {
							fmt.Println("✅ Liga 'GT Leagues' encontrada dentro de iframe.")
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

		fmt.Printf("⏳ Procurando 'GT Leagues'... (%v decorridos)\n", time.Since(start).Round(time.Second))
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("liga 'GT Leagues' não encontrada após %v", config.TimeoutSearch)
}

func ExpandLeague(league *rod.Element) error {
	fmt.Println("⏳ Tentando expandir a liga...")
	xpath := `ancestor::div[contains(@class,"header") or contains(@class,"Header")]//button | ancestor::div[contains(@class,"header") or contains(@class,"Header")]//i[contains(@class, "arrow") or contains(@class, "Arrow")] | .`
	btn, err := league.ElementX(xpath)
	if err != nil {
		fmt.Println("⚠️ Botão da liga não encontrado explicitamente, tentando clicar no elemento da liga.")
		btn = league
	}

	if err := btn.ScrollIntoView(); err != nil {
		return fmt.Errorf("não foi possível dar scroll para a liga: %w", err)
	}
	if err := btn.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("não foi possível clicar na liga: %w", err)
	}

	fmt.Println("✅ Clique na liga realizado.")
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
	fmt.Println("📜 Rolando para exibir partidas...")
	return nil
}

func GetMatches(league *rod.Element, player1, player2 string) []models.Match {
	var matches []models.Match

	container, err := league.ElementX(`ancestor::div[contains(@class,"header") or contains(@class,"Header")]/parent::div`)
	if err != nil {
		fmt.Println("⚠️ Não foi possível localizar o container da liga:", err)
		return matches
	}

	var events rod.Elements
	for i := 0; i < 15; i++ {
		evs, err := container.Elements(`.eventlist_eu_fe_EventItemDesktop_wrapper, [class*="EventItemDesktop_wrapper"], [class*="EventItem"]`)
		if err == nil && len(evs) > 0 {
			events = evs
			break
		}
		time.Sleep(1 * time.Second)
	}

	seen := make(map[string]bool)

	for _, e := range events {
		teams, _ := e.Elements(`[class*="teamNameText"], [class*="participantName"]`)
		if len(teams) < 2 {
			continue
		}

		t1, _ := teams[0].Text()
		t2, _ := teams[1].Text()
		t1 = strings.TrimSpace(t1)
		t2 = strings.TrimSpace(t2)

		matchP1 := strings.Contains(strings.ToLower(t1), strings.ToLower(player1))
		matchP2 := strings.Contains(strings.ToLower(t2), strings.ToLower(player2))

		if !matchP1 || !matchP2 {
			matchP1 = strings.Contains(strings.ToLower(t1), strings.ToLower(player2))
			matchP2 = strings.Contains(strings.ToLower(t2), strings.ToLower(player1))
		}

		if !matchP1 || !matchP2 {
			continue
		}

		key := t1 + "|" + t2
		if seen[key] {
			continue
		}
		seen[key] = true

		scores, _ := e.Elements(`[class*="mainScore"]`)
		timeText := ""
		timeEl, err := e.Element(`[class*="LiveEventCounter_wrapper"], [class*="Time"]`)
		if err == nil {
			timeText, _ = timeEl.Text()
		}

		oddsEls, _ := e.Elements(`[class*="Selection_odds"], [class*="oddsText"]`)

		match := models.Match{
			Team1: t1,
			Team2: t2,
		}

		if len(scores) >= 2 {
			match.Score1, _ = scores[0].Text()
			match.Score2, _ = scores[1].Text()
		}

		if len(oddsEls) >= 3 {
			match.Odd1, _ = oddsEls[0].Text()
			match.OddX, _ = oddsEls[1].Text()
			match.Odd2, _ = oddsEls[2].Text()
		}

		timeText = strings.ReplaceAll(timeText, "\n", " ")
		match.Time = strings.TrimSpace(timeText)

		fmt.Printf("🔍 Entrando na partida: %s vs %s\n", t1, t2)

		clickTarget, err := teams[0].ElementX(`ancestor::div[contains(@class, "participant") or contains(@class, "Team")][1]`)
		if err != nil {
			clickTarget = teams[0]
		}

		_ = clickTarget.ScrollIntoView()
		_ = clickTarget.Click(proto.InputMouseButtonLeft, 1)

		// Aguarda o container de mercados carregar
		_, _ = e.Page().Timeout(5 * time.Second).Element(".eventdetails_eu_fe_ViewStyles_scrollContainer, .market_fe_MarketList_wrapper, .MarketList_wrapper")

		match.HTHandicap1, match.HTHandicap2, _ = fetchHandicap(e.Page())

		// Volta
		backBtn, err := e.Page().ElementX(config.SelectorBackBtn)
		if err == nil {
			_ = backBtn.Click(proto.InputMouseButtonLeft, 1)
			time.Sleep(config.DelayAction)
		}

		matches = append(matches, match)
	}

	return matches
}

func fetchHandicap(page *rod.Page) ([]models.HandicapLine, []models.HandicapLine, bool) {
	var h1, h2 []models.HandicapLine
	found := false

	// Busca pelo mercado de Handicap
	var marketContainer *rod.Element
	for i := 0; i < 5; i++ {
		el, err := page.ElementX(config.XPathHandicapMarket)
		if err == nil {
			marketContainer = el
			break
		}
		// Tenta em iframes
		iframes, _ := page.Elements("iframe")
		for _, iframe := range iframes {
			if f, err := iframe.Frame(); err == nil {
				el, err = f.ElementX(config.XPathHandicapMarket)
				if err == nil {
					marketContainer = el
					break
				}
			}
		}
		if marketContainer != nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	if marketContainer != nil {
		fmt.Println("✅ Mercado de Handicap encontrado.")
		buttons, err := marketContainer.Elements(config.SelectorHandicapBtns)
		if err == nil && len(buttons) >= 2 {
			found = true
			for i := 0; i < len(buttons); i += 2 {
				if i+1 >= len(buttons) {
					break
				}

				// Time 1
				line1, _ := buttons[i].Element(config.SelectorHandicapPoints)
				odd1, _ := buttons[i].Element(config.SelectorHandicapOdds)
				if line1 != nil && odd1 != nil {
					l1, _ := line1.Text()
					o1, _ := odd1.Text()
					o1 = strings.ReplaceAll(o1, "▲", "")
					o1 = strings.ReplaceAll(o1, "▼", "")
					h1 = append(h1, models.HandicapLine{Line: strings.TrimSpace(l1), Odd: strings.TrimSpace(o1)})
				}

				// Time 2
				line2, _ := buttons[i+1].Element(config.SelectorHandicapPoints)
				odd2, _ := buttons[i+1].Element(config.SelectorHandicapOdds)
				if line2 != nil && odd2 != nil {
					l2, _ := line2.Text()
					o2, _ := odd2.Text()
					o2 = strings.ReplaceAll(o2, "▲", "")
					o2 = strings.ReplaceAll(o2, "▼", "")
					h2 = append(h2, models.HandicapLine{Line: strings.TrimSpace(l2), Odd: strings.TrimSpace(o2)})
				}
			}
		}
	}

	return h1, h2, found
}
