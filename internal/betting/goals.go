package betting

import (
	"bolsadeaposta-bot/internal/config"
	"bolsadeaposta-bot/internal/models"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// Helper function to extract info from an odd button
func parseOddEl(oddEl *rod.Element, index int) (isOver, isUnder bool, line, odd float64, ok bool) {
	text, _ := oddEl.Text()
	lower := strings.ToLower(text)

	isOver = strings.Contains(lower, "mais") || strings.Contains(lower, "over") || strings.Contains(lower, "acima") || strings.Contains(lower, " o ") || strings.HasPrefix(lower, "o ") || strings.Contains(lower, "mais de")
	isUnder = strings.Contains(lower, "menos") || strings.Contains(lower, "under") || strings.Contains(lower, "abaixo") || strings.Contains(lower, " u ") || strings.HasPrefix(lower, "u ") || strings.Contains(lower, "menos de")

	if !isOver && !isUnder {
		if index%2 == 0 {
			isOver = true
		} else {
			isUnder = true
		}
	}

	ptsEl, err1 := oddEl.Element(`[class*='points'], [class*='Points'], [class*='header'], [class*='line'], [class*='Line']`)
	oddsEl, err2 := oddEl.Element(`[class*='odds'], [class*='Odds']`)

	if err1 == nil && err2 == nil {
		tLine, _ := ptsEl.Text()
		tOdd, _ := oddsEl.Text()
		re := regexp.MustCompile(`\d+(?:\.\d+)?`)
		mLine := re.FindString(tLine)
		mOdd := re.FindString(strings.ReplaceAll(strings.ReplaceAll(tOdd, "▲", ""), "▼", ""))

		if mLine != "" && mOdd != "" {
			line, _ = strconv.ParseFloat(mLine, 64)
			odd, _ = strconv.ParseFloat(mOdd, 64)
			ok = true
			return
		}
	}

	lines := strings.Split(text, "\n")
	var numbers []float64
	re := regexp.MustCompile(`\d+(?:\.\d+)?`)
	for _, l := range lines {
		l = strings.ReplaceAll(strings.ReplaceAll(l, "▲", ""), "▼", "")
		matches := re.FindAllString(l, -1)
		for _, m := range matches {
			if val, err := strconv.ParseFloat(m, 64); err == nil {
				numbers = append(numbers, val)
			}
		}
	}

	if len(numbers) >= 2 {
		line = numbers[0]
		odd = numbers[len(numbers)-1]
		ok = true
	} else if len(numbers) == 1 {
		// Possibly text format "4.5 1.80"
		reCombo := regexp.MustCompile(`(\d+(?:\.\d+)?)\s+(\d+(?:\.\d+)?)`)
		matches := reCombo.FindStringSubmatch(text)
		if len(matches) == 3 {
			line, _ = strconv.ParseFloat(matches[1], 64)
			odd, _ = strconv.ParseFloat(matches[2], 64)
			ok = true
		}
	}
	return
}

// PrepareGoalBet specifically looks for goals logic and validates the odd minimum
func PrepareGoalBet(page *rod.Page, tip *models.Tip, amount string) error {
	sportsbookIframeElement, err := page.Element(config.SelectorIframeFSSB)
	if err != nil {
		return fmt.Errorf("iframe do sportsbook não encontrado")
	}

	frame, err := sportsbookIframeElement.Frame()
	if err != nil {
		return fmt.Errorf("erro frame")
	}

	_ = ClearBetslip(frame)

	// Busca pelo mercado de Aposta Ao Vivo Mais/Menos
	var marketContainer *rod.Element
	for i := 0; i < 5; i++ {
		el, err := frame.ElementX(config.XPathGoalMarket)
		if err == nil {
			marketContainer = el
			break
		}
		time.Sleep(1 * time.Second)
	}

	if marketContainer == nil {
		return fmt.Errorf("mercado 'Aposta Ao Vivo Mais/Menos' não encontrado")
	}

	odds, err := marketContainer.Elements(`button, [class*='Selection_line']`)
	if err != nil || len(odds) == 0 {
		return fmt.Errorf("não encontrou botões no mercado de gols")
	}

	tipIsOver := strings.Contains(strings.ToLower(tip.Market), "mais") || strings.Contains(strings.ToLower(tip.Market), "over")
	tipIsUnder := strings.Contains(strings.ToLower(tip.Market), "menos") || strings.Contains(strings.ToLower(tip.Market), "under")
	tipLineF, _ := strconv.ParseFloat(tip.Line, 64)

	var targetOddElement *rod.Element
	var bestLine float64
	var bestOdd float64
	var foundValidButton bool

	for i, oddEl := range odds {
		isOver, isUnder, lineVal, oddVal, ok := parseOddEl(oddEl, i)
		if !ok {
			continue
		}

		if tipIsOver && isOver {
			// for "Mais de", market line must be <= bot line (safer for us)
			if lineVal <= tipLineF && oddVal >= tip.TargetOdd {
				// We want the most advantageous line. For "Mais de", smaller line is better.
				if !foundValidButton || lineVal < bestLine {
					bestLine = lineVal
					bestOdd = oddVal
					targetOddElement = oddEl
					foundValidButton = true
				}
			}
		} else if tipIsUnder && isUnder {
			// for "Menos de", market line must be >= bot line (safer for us)
			if lineVal >= tipLineF && oddVal >= tip.TargetOdd {
				// For "Menos de", larger line is better.
				if !foundValidButton || lineVal > bestLine {
					bestLine = lineVal
					bestOdd = oddVal
					targetOddElement = oddEl
					foundValidButton = true
				}
			}
		}
	}

	if !foundValidButton {
		return fmt.Errorf("nenhuma linha válida encontrada que atenda aos critérios (tipo over=%v, linha bot: %.2f, odd mínima: %.2f)", tipIsOver, tipLineF, tip.TargetOdd)
	}

	_ = targetOddElement.ScrollIntoView()
	time.Sleep(500 * time.Millisecond)
	_ = targetOddElement.Click(proto.InputMouseButtonLeft, 1)

	stakeInput, err := frame.Timeout(5 * time.Second).Element(config.SelectorBetslipStakeInput)
	if err != nil {
		return fmt.Errorf("campo de stake ausente")
	}

	_ = stakeInput.Input(amount)
	
	// Confirma
	placeBtn, err := frame.Element(config.SelectorBetslipPlaceBetBtn)
	if err == nil {
		btnText, _ := placeBtn.Text()
		log.Printf("🚀 Aposta preparada! Odd final: %.2f (Linha %.2f) | Clicando no botão: [%s]", bestOdd, bestLine, strings.TrimSpace(btnText))
		if err := placeBtn.Click(proto.InputMouseButtonLeft, 1); err != nil {
			return fmt.Errorf("erro ao confirmar aposta no boletim: %w", err)
		}
		log.Println("✅ Aposta confirmada com sucesso!")
	} else {
		return fmt.Errorf("botão de confirmação não encontrado no boletim")
	}

	return nil
}
