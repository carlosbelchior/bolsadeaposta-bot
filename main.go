package main

import (
	"fmt"

	"betstake-webscrap/internal/browser"
	"betstake-webscrap/internal/crawler"
)

func main() {
	browserInstance, page := browser.LoadPageFlow()
	defer browserInstance.MustClose()

	league := crawler.FindLeague(page)
	if league != nil {
		crawler.ExpandLeague(league)

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
	} else {
		fmt.Println("❌ GT Leagues não encontrada na página.")
	}

	fmt.Println("🚀 Processo finalizado.")
}
