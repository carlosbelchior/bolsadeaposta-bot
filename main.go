package main

import (
	"fmt"
	"os"
	"strings"

	"betstake-webscrap/internal/betting"
	"betstake-webscrap/internal/browser"
	"betstake-webscrap/internal/config"
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

	var playerOneName, playerTwoName string
	fmt.Print("Digite o nome do primeiro jogador: ")
	fmt.Scanln(&playerOneName)
	fmt.Print("Digite o nome do segundo jogador: ")
	fmt.Scanln(&playerTwoName)

	matches := crawler.GetMatches(league, playerOneName, playerTwoName)
	fmt.Printf("Encontradas %d partidas:\n", len(matches))
	for _, match := range matches {
		fmt.Printf("- %s vs %s (Placar: %s-%s, Tempo: %s)\n", match.Team1, match.Team2, match.Score1, match.Score2, match.Time)
		if len(match.HTHandicap1) > 0 {
			fmt.Println("  [Handicap Asiático 1º Tempo - Casa]")
			for _, handicap := range match.HTHandicap1 {
				fmt.Printf("    Linha: %s | Odd: %s\n", handicap.Line, handicap.Odd)
			}
		}
		if len(match.HTHandicap2) > 0 {
			fmt.Println("  [Handicap Asiático 1º Tempo - Fora]")
			for _, handicap := range match.HTHandicap2 {
				fmt.Printf("    Linha: %s | Odd: %s\n", handicap.Line, handicap.Odd)
			}
		}
	}

	if len(matches) > 0 {
		fmt.Printf("\n💸 Deseja simular uma aposta para alguma partida encontrada? (s/n): ")
		var userResponse string
		fmt.Scanln(&userResponse)

		if strings.ToLower(userResponse) == "s" {
			var selectedMatchIndex int
			fmt.Printf("Selecione o número da partida (1-%d): ", len(matches))
			fmt.Scanln(&selectedMatchIndex)

			if selectedMatchIndex >= 1 && selectedMatchIndex <= len(matches) {
				match := matches[selectedMatchIndex-1]
				fmt.Printf("Escolha o time (1 for %s, 2 for %s): ", match.Team1, match.Team2)
				var selectedTeamOption int
				fmt.Scanln(&selectedTeamOption)

				teamName := match.Team1
				if selectedTeamOption == 2 {
					teamName = match.Team2
				}

				fmt.Printf("Digite a linha de handicap (ex: -0.5, 0.0, +0.25): ")
				var handicapLine string
				fmt.Scanln(&handicapLine)

				if err := betting.PrepareHandicapBet(page, teamName, handicapLine, config.DefaultStake); err != nil {
					fmt.Printf("❌ Falha na simulação: %v\n", err)
				}
			}
		}
	}

	fmt.Println("🚀 Processo finalizado.")
}
