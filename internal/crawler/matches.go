package crawler

import (
	"fmt"
	"strings"
	"time"

	"betstake-webscrap/internal/models"

	"github.com/go-rod/rod"
)

func FindLeague(page *rod.Page) *rod.Element {
	fmt.Println("⏳ Buscando pela liga 'GT Leagues'...")
	start := time.Now()
	xpath := `//*[contains(translate(text(), 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz'), 'gt leagues')]`

	for time.Since(start) < 30*time.Second {
		targetLeagues, _ := page.ElementsX(xpath)
		for _, league := range targetLeagues {
			text := league.MustText()
			if strings.Contains(strings.ToLower(text), "gt leagues") {
				fmt.Println("✅ Liga 'GT Leagues' encontrada na página principal.")
				league.MustScrollIntoView()
				return league
			}
		}

		iframes, _ := page.Elements("iframe")
		for _, iframe := range iframes {
			src, _ := iframe.Attribute("src")
			if src != nil && strings.Contains(*src, "fssb.io") {
				f, err := iframe.Frame()
				if err == nil {
					targetLeaguesInFrame, _ := f.ElementsX(xpath)
					for _, league := range targetLeaguesInFrame {
						text := league.MustText()
						if strings.Contains(strings.ToLower(text), "gt leagues") {
							fmt.Println("✅ Liga 'GT Leagues' encontrada dentro de iframe.")
							league.MustScrollIntoView()
							return league
						}
					}
				}
			}
		}

		page.MustEval(`() => {
			let scrollContainer = document.querySelector('.master_fe_ViewStyles_desktopWrapperEventListIndependentScroll, .ViewStyles_desktopWrapperEventListIndependentScroll, [class*="EventListIndependentScroll"]');
			if (scrollContainer) {
				scrollContainer.scrollBy(0, 800);
			} else {
				window.scrollBy(0, 800);
			}
		}`)

		allLeagues, _ := page.Elements(`[class*="LeagueItemDesktop_header"], [class*="leagueHeader"], [class*="LeagueHeader"], [class*="LeagueItem"]`)
		if len(allLeagues) > 0 {
			allLeagues[len(allLeagues)-1].ScrollIntoView()
		}

		fmt.Printf("⏳ Procurando 'GT Leagues'... (%v decorridos)\n", time.Since(start).Round(time.Second))
		time.Sleep(2 * time.Second)
	}

	return nil
}

func ExpandLeague(league *rod.Element) {
	fmt.Println("⏳ Tentando expandir a liga...")
	btn, err := league.ElementX(`ancestor::div[contains(@class,"header") or contains(@class,"Header")]//button | ancestor::div[contains(@class,"header") or contains(@class,"Header")]//i[contains(@class, "arrow") or contains(@class, "Arrow")] | .`)
	if err != nil {
		fmt.Println("⚠️ Botão da liga não encontrado explicitamente, tentando clicar no elemento da liga.")
		btn = league
	}

	btn.MustScrollIntoView()
	btn.MustWaitVisible()
	btn.MustClick()
	fmt.Println("✅ Clique na liga realizado.")
	time.Sleep(3 * time.Second)
	btn.MustScrollIntoView()
}

func GetMatches(league *rod.Element, player1, player2 string) []models.Match {
	var matches []models.Match

	container, err := league.ElementX(`ancestor::div[contains(@class,"header") or contains(@class,"Header")]/parent::div`)
	if err != nil {
		fmt.Println("⚠️ Não foi possível localizar o container da liga:", err)
		return matches
	}

	var events []*rod.Element
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

		t1 := strings.TrimSpace(teams[0].MustText())
		t2 := strings.TrimSpace(teams[1].MustText())

		matchP1 := strings.Contains(t1, "("+player1+")") || strings.Contains(t1, player1)
		matchP2 := strings.Contains(t2, "("+player2+")") || strings.Contains(t2, player2)

		if !matchP1 || !matchP2 {
			matchP1 = strings.Contains(t1, "("+player2+")") || strings.Contains(t1, player2)
			matchP2 = strings.Contains(t2, "("+player1+")") || strings.Contains(t2, player1)
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
			timeText = timeEl.MustText()
		}

		oddsEls, _ := e.Elements(`[class*="Selection_odds"], [class*="oddsText"]`)

		match := models.Match{
			Team1: t1,
			Team2: t2,
		}

		if len(scores) >= 2 {
			match.Score1 = scores[0].MustText()
			match.Score2 = scores[1].MustText()
		}

		if len(oddsEls) >= 3 {
			match.Odd1 = oddsEls[0].MustText()
			match.OddX = oddsEls[1].MustText()
			match.Odd2 = oddsEls[2].MustText()
		}

		timeText = strings.ReplaceAll(timeText, "\n", " ")
		match.Time = strings.TrimSpace(timeText)

		fmt.Printf("🔍 Entrando na partida: %s vs %s\n", t1, t2)

		clickTarget, err := teams[0].ElementX(`ancestor::div[contains(@class, "participant") or contains(@class, "Team")][1]`)
		if err != nil {
			clickTarget = teams[0]
		}

		clickTarget.MustScrollIntoView()
		clickTarget.MustClick()
		// Aguarda o container de mercados carregar com timeout para não travar
		_ = rod.Try(func() {
			e.Page().Timeout(5 * time.Second).MustElement(".eventdetails_eu_fe_ViewStyles_scrollContainer, .market_fe_MarketList_wrapper, .MarketList_wrapper")
		})

		var found bool
		match.HTHandicap1, match.HTHandicap2, found = fetchOdd(e)

		if !found {
			backBtn, err := e.Page().Element(".navigation_eu_fe_Breadcrumbs_backButton")
			if err == nil {
				backBtn.MustClick()
				time.Sleep(1 * time.Second)
			}
		}

		backBtn, err := e.Page().ElementX(`//div[contains(@class, "eventdetails")]//div[contains(@class, "backButton")] | //div[contains(@class, "MarketHeader_back")] | //div[contains(@class, "navigation_eu_fe_Breadcrumbs_backButton")]`)
		if err == nil {
			backBtn.MustClick()
			time.Sleep(1 * time.Second)
		}

		matches = append(matches, match)
	}

	return matches
}

func fetchHandicap(e *rod.Element) ([]models.HandicapLine, []models.HandicapLine, bool) {
	page := e.Page()
	var h1, h2 []models.HandicapLine
	found := false

	fmt.Printf("🔍 Iniciando busca de handicap para partida. Contexto inicial: %s\n", page.MustInfo().URL)

	// Tenta clicar na aba "Tempos" ou "Halves" se estiver disponível
	_ = rod.Try(func() {
		var halvesTab *rod.Element
		var activePage any = page // Pode ser *rod.Page ou *rod.Frame

		findTab := func(p any) *rod.Element {
			xpath := `//span[contains(translate(text(), 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz'), "tempos") or contains(translate(text(), 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz'), "halves")]`
			var spans rod.Elements
			var err error

			if pg, ok := p.(*rod.Page); ok {
				spans, err = pg.ElementsX(xpath)
			} else {
				// Use reflect ou cast para o tipo base se rod.Frame não estiver acessível diretamente como tipo,
				// mas aqui vamos tentar usar uma interface comum se possível ou apenas lidar com Page por enquanto
				// e tentar encontrar o Frame de outra forma se necessário.
				// Na verdade rod.Frame deve existir. Vamos ver se ele está no pacote.
			}

			if err == nil && len(spans) > 0 {
				for _, span := range spans {
					tab, err := span.ElementX(`ancestor::*[contains(@class, "marketsTabsItem")]`)
					if err == nil {
						return tab
					}
					tab, err = span.ElementX(`..`)
					if err == nil {
						return tab
					}
				}
			}
			return nil
		}

		for i := 0; i < 5; i++ {
			halvesTab = findTab(activePage)
			if halvesTab == nil {
				var iframes rod.Elements
				if pg, ok := activePage.(*rod.Page); ok {
					iframes, _ = pg.Elements("iframe")
				}

				for _, iframe := range iframes {
					src, _ := iframe.Attribute("src")
					if src != nil && strings.Contains(*src, "fssb.io") {
						if f, err := iframe.Frame(); err == nil {
							// Se encontrar o iframe, tenta procurar a aba nele
							xpath := `//span[contains(translate(text(), 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz'), "tempos") or contains(translate(text(), 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz'), "halves")]`
							spans, err := f.ElementsX(xpath)
							if err == nil && len(spans) > 0 {
								for _, span := range spans {
									tab, err := span.ElementX(`ancestor::*[contains(@class, "marketsTabsItem")]`)
									if err == nil {
										fmt.Println("📍 Aba 'Tempos/Halves' encontrada dentro de iframe.")
										halvesTab = tab
										// Gambiarra para manter a referência do frame se precisarmos pesquisar mais coisas nele
										// Como não conseguimos tipar activePage como *rod.Frame facilmente aqui sem erro de compilação aparente,
										// vamos usar o fato de que rod.Element (o tab) conhece seu contexto.
										break
									}
								}
							}
						}
					}
					if halvesTab != nil {
						break
					}
				}
			}
			if halvesTab != nil {
				break
			}
			fmt.Printf("⏳ Tentando encontrar aba 'Tempos/Halves'... (tentativa %d)\n", i+1)
			time.Sleep(1 * time.Second)
		}

		if halvesTab != nil {
			selected, _ := halvesTab.Attribute("aria-selected")
			if selected != nil && *selected == "true" {
				fmt.Println("✅ Aba 'Tempos/Halves' já selecionada.")
			} else {
				fmt.Println("⏳ Clicando na aba 'Tempos/Halves'...")
				halvesTab.MustScrollIntoView()
				halvesTab.MustWaitVisible()
				halvesTab.MustClick()
				time.Sleep(2 * time.Second)
			}
		} else {
			fmt.Println("⚠️ Aba 'Tempos/Halves' não encontrada em nenhum contexto.")
		}

		// Procura pelo mercado de Handicap na aba de Tempos
		var marketContainer *rod.Element
		findMarket := func() *rod.Element {

			xpath := `//*
				[
					contains(@class, "MarketHeader_marketName") 
					or contains(@class, "eventpage_fe_Markets_marketName") 
					or contains(@class, "MarketName") 
					or contains(@class, "header")
				]
				[
					contains(text(), "Asian") 
					or contains(text(), "Asiático")
				]
					/ancestor::div[
					contains(@class, "MarketGroup_wrapper") 
					or contains(@class, "MarketGroup") 
					or contains(@class, "marketsList_marketGroup")
				]`

			// 🔹 tenta na página principal
			el, err := page.ElementX(xpath)
			if err == nil {
				return el
			}

			// 🔹 tenta nos iframes
			iframes, _ := page.Elements("iframe")
			for _, iframe := range iframes {
				src, _ := iframe.Attribute("src")
				if src != nil && strings.Contains(*src, "fssb.io") {
					if f, err := iframe.Frame(); err == nil {
						el, err = f.ElementX(xpath)
						if err == nil {
							return el
						}
					}
				}
			}

			return nil
		}

		marketContainer = findMarket()

		if marketContainer != nil {
			fmt.Println("✅ Mercado de Handicap encontrado.")
			// Usando o seletor fornecido no exemplo HTML: eventpage_fe_HandicapSelection_line
			buttons, err := marketContainer.Elements(".eventpage_fe_HandicapSelection_line, button[class*='HandicapSelection_line']")
			if err == nil && len(buttons) >= 2 {
				fmt.Printf("📊 Encontradas %d opções de Handicap.\n", len(buttons))
				found = true
				// Supomos que os botões vêm em pares (Home/Away)
				for i := 0; i < len(buttons); i += 2 {
					if i+1 >= len(buttons) {
						break
					}

					// Time 1
					line1, _ := buttons[i].Element(".eventpage_fe_HandicapSelection_points, [class*='HandicapSelection_points']")
					odd1, _ := buttons[i].Element(".eventpage_fe_HandicapSelection_odds, [class*='HandicapSelection_odds']")
					if line1 != nil && odd1 != nil {
						l1 := strings.TrimSpace(line1.MustText())
						o1 := strings.TrimSpace(odd1.MustText())
						o1 = strings.ReplaceAll(o1, "▲", "")
						o1 = strings.ReplaceAll(o1, "▼", "")
						h1 = append(h1, models.HandicapLine{Line: l1, Odd: o1})
					}

					// Time 2
					line2, _ := buttons[i+1].Element(".eventpage_fe_HandicapSelection_points, [class*='HandicapSelection_points']")
					odd2, _ := buttons[i+1].Element(".eventpage_fe_HandicapSelection_odds, [class*='HandicapSelection_odds']")
					if line2 != nil && odd2 != nil {
						l2 := strings.TrimSpace(line2.MustText())
						o2 := strings.TrimSpace(odd2.MustText())
						o2 = strings.ReplaceAll(o2, "▲", "")
						o2 = strings.ReplaceAll(o2, "▼", "")
						h2 = append(h2, models.HandicapLine{Line: l2, Odd: o2})
					}
				}
			} else {
				fmt.Println("⚠️ Botões de Handicap não encontrados dentro do container.")
			}
		} else {
			fmt.Println("⚠️ Mercado de Handicap não encontrado.")
		}
	})

	return h1, h2, found
}

func fetchOdd(e *rod.Element) ([]models.HandicapLine, []models.HandicapLine, bool) {
	return fetchHandicap(e)
}
