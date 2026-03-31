package main

import (
	"fmt"
	"os"
	"strings"

	"betstake-webscrap/internal/betting"
	"betstake-webscrap/internal/browser"
	"betstake-webscrap/internal/crawler"
)

func main() {
	browserInstance, page, err := browser.LoadPageFlow()
	if err != nil {
		fmt.Printf("❌ Erro fatal: %v\n", err)
		os.Exit(1)
	}
	defer browserInstance.MustClose()

	league, err := crawler.FindLeague(page)
	if err != nil {
		fmt.Printf("❌ Erro ao buscar liga: %v\n", err)
		return
	}

	if err := crawler.ExpandLeague(league); err != nil {
		fmt.Printf("❌ Erro ao expandir liga: %v\n", err)
		return
	}

	var p1, p2 string
	fmt.Print("Digite o nome do primeiro jogador: ")
	fmt.Scanln(&p1)
	fmt.Print("Digite o nome do segundo jogador: ")
	fmt.Scanln(&p2)

	matches := crawler.GetMatches(league, p1, p2)
	fmt.Printf("Encontradas %d partidas:\n", len(matches))
	for _, m := range matches {
		fmt.Printf("- %s vs %s (Placar: %s-%s, Tempo: %s)\n", m.Team1, m.Team2, m.Score1, m.Score2, m.Time)
		if len(m.HTHandicap1) > 0 {
			fmt.Println("  [Handicap Asiático 1º Tempo - Casa]")
			for _, h := range m.HTHandicap1 {
				fmt.Printf("    Linha: %s | Odd: %s\n", h.Line, h.Odd)
			}
		}
		if len(m.HTHandicap2) > 0 {
			fmt.Println("  [Handicap Asiático 1º Tempo - Fora]")
			for _, h := range m.HTHandicap2 {
				fmt.Printf("    Linha: %s | Odd: %s\n", h.Line, h.Odd)
			}
		}
	}

	if len(matches) > 0 {
		fmt.Printf("\n💸 Deseja simular uma aposta para alguma partida encontrada? (s/n): ")
		var resp string
		fmt.Scanln(&resp)

		if strings.ToLower(resp) == "s" {
			var matchIdx int
			fmt.Printf("Selecione o número da partida (1-%d): ", len(matches))
			fmt.Scanln(&matchIdx)

			if matchIdx >= 1 && matchIdx <= len(matches) {
				m := matches[matchIdx-1]
				fmt.Printf("Escolha o time (1 for %s, 2 for %s): ", m.Team1, m.Team2)
				var teamSelect int
				fmt.Scanln(&teamSelect)

				teamName := m.Team1
				if teamSelect == 2 {
					teamName = m.Team2
				}

				fmt.Printf("Digite a linha de handicap (ex: -0.5, 0.0, +0.25): ")
				var hl string
				fmt.Scanln(&hl)

				fmt.Printf("Digite o valor da aposta (ex: 1.00): ")
				var amt string
				fmt.Scanln(&amt)

				if err := betting.PrepareHandicapBet(page, teamName, hl, amt); err != nil {
					fmt.Printf("❌ Falha na simulação: %v\n", err)
				}
			}
		}
	}

	fmt.Println("🚀 Processo finalizado.")
}
