package telegram

import (
	"bolsadeaposta-bot/internal/models"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// Matches: 📈Mais de 4.5 - Gols @1.71
	marketOddRegex = regexp.MustCompile(`📈\s*(.*?)\s*@([\d\.]+)`)
	// Matches: Mais de 4.5, Menos de 2.0
	lineRegex      = regexp.MustCompile(`(?:Mais de|Menos de)\s*([\d\.]+)`)
	// Matches: 🅿️Placar da Tip: 0-0
	scoreRegex     = regexp.MustCompile(`Placar da Tip:\s*(\d+\s*-\s*\d+)`)
	// Matches: 🏃Habibi (Sporting) vs Jose (Galatasaray)
	teamsRegex     = regexp.MustCompile(`🏃(.*?)\s*\(.*?\)\s*vs\s*(.*?)\s*\(.*?\)`)
)

func ParseTipMessage(messageText string) (*models.Tip, error) {
	if !strings.Contains(messageText, "📈") {
		return nil, fmt.Errorf("mensagem não aparenta ser uma tip de aposta")
	}

	tip := &models.Tip{
		ID:        fmt.Sprintf("tip-%d", time.Now().UnixNano()),
		Status:    models.StatusPending,
		CreatedAt: time.Now(),
	}

	// 1. Extrair Mercado e Odd
	marketMatch := marketOddRegex.FindStringSubmatch(messageText)
	if len(marketMatch) >= 3 {
		tip.Market = strings.TrimSpace(marketMatch[1])
		
		oddParsed, err := strconv.ParseFloat(marketMatch[2], 64)
		if err == nil {
			tip.TargetOdd = oddParsed
		}
	} else {
		return nil, fmt.Errorf("não foi possível extrair a odd e o mercado")
	}

	// 2. Extrair Linha do Mercado (se houver)
	lineMatch := lineRegex.FindStringSubmatch(tip.Market)
	if len(lineMatch) >= 2 {
		tip.Line = lineMatch[1]
	}

	// 3. Extrair Placar
	scoreMatch := scoreRegex.FindStringSubmatch(messageText)
	if len(scoreMatch) >= 2 {
		scoreStr := strings.ReplaceAll(scoreMatch[1], " ", "") // ensure 0-0 Format
		tip.Score = scoreStr
	} else {
		// Default to 0-0 se falhar ou se for tip pré-live sem isso
		tip.Score = "0-0"
	}

	// 4. Extrair Times
	teamsMatch := teamsRegex.FindStringSubmatch(messageText)
	if len(teamsMatch) >= 3 {
		tip.HomeTeam = strings.TrimSpace(teamsMatch[1])
		tip.AwayTeam = strings.TrimSpace(teamsMatch[2])
	} else {
		return nil, fmt.Errorf("não foi possível extrair os nomes dos times da mensagem")
	}

	return tip, nil
}
