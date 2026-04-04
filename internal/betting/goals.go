package betting

import (
	"bolsadeaposta-bot/internal/browser"
	"bolsadeaposta-bot/internal/config"
	"bolsadeaposta-bot/internal/logger"
	"bolsadeaposta-bot/internal/models"
	"fmt"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// ParseOddString extracts line and odd values from a text string using regex and split logic.
func ParseOddString(text string) (line, odd float64, ok bool) {
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
		return
	}
	return
}

// parseOddEl uses ParseOddString and element-specific logic to extract betting info.
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

	line, odd, ok = ParseOddString(text)
	return
}

// ValidateBet checks if the slip line and odd are acceptable for a given tip.
func ValidateBet(tip *models.Tip, slipLine, slipOdd float64) (bool, string) {
	tipIsOver := strings.Contains(strings.ToLower(tip.Market), "mais") || strings.Contains(strings.ToLower(tip.Market), "over")
	tipIsUnder := strings.Contains(strings.ToLower(tip.Market), "menos") || strings.Contains(strings.ToLower(tip.Market), "under")
	tipLineF, _ := strconv.ParseFloat(tip.Line, 64)

	if tipIsOver {
		// for "Mais de", market line must be <= bot line (safer for us)
		if slipLine <= tipLineF && slipOdd >= tip.TargetOdd {
			return true, ""
		}
		return false, fmt.Sprintf("Aposta Over: Linha bilhete %.2f (Esperado <= %.2f), Odd bilhete %.2f (Esperado >= %.2f)", slipLine, tipLineF, slipOdd, tip.TargetOdd)
	} else if tipIsUnder {
		// for "Menos de", market line must be >= bot line (safer for us)
		if slipLine >= tipLineF && slipOdd >= tip.TargetOdd {
			return true, ""
		}
		return false, fmt.Sprintf("Aposta Under: Linha bilhete %.2f (Esperado >= %.2f), Odd bilhete %.2f (Esperado >= %.2f)", slipLine, tipLineF, slipOdd, tip.TargetOdd)
	}
	return false, "Tipo de mercado não identificado (Over/Under)"
}

// PrepareGoalBet especificamente busca a lógica de gols e valida a odd mínima
func PrepareGoalBet(page *rod.Page, tip *models.Tip, amount string) error {
	// Limpa popups na página principal antes de qualquer ação
	browser.CheckAndDismissPopups(page)

	frame, err := GetBetslipFrame(page)
	if err != nil {
		return err
	}

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

	var targetSelection *rod.Element
	var bestLine float64
	var bestOdd float64
	var foundValidButton bool

	for i, oddEl := range odds {
		isOver, isUnder, lineVal, oddVal, ok := parseOddEl(oddEl, i)
		if !ok {
			continue
		}

		if tipIsOver && isOver {
			if lineVal <= tipLineF && oddVal >= tip.TargetOdd {
				if !foundValidButton || lineVal < bestLine {
					bestLine = lineVal
					bestOdd = oddVal
					targetSelection = oddEl
					foundValidButton = true
				}
			}
		} else if tipIsUnder && isUnder {
			if lineVal >= tipLineF && oddVal >= tip.TargetOdd {
				if !foundValidButton || lineVal > bestLine {
					bestLine = lineVal
					bestOdd = oddVal
					targetSelection = oddEl
					foundValidButton = true
				}
			}
		}
	}

	if !foundValidButton {
		return fmt.Errorf("nenhuma linha válida encontrada (over=%v, linha bot: %.2f, odd min: %.2f)", tipIsOver, tipLineF, tip.TargetOdd)
	}

	// Realiza o fluxo do bilhete
	if err := PerformBetslipFlow(frame, targetSelection, amount); err != nil {
		return err
	}

	// Validar a linha e odd no bilhete antes de clicar em confirmar
	time.Sleep(1 * time.Second)

	var slipLine float64 = bestLine
	var slipOdd float64 = bestOdd
	var foundSlipInfo bool

	removeBtn, err := frame.Element(config.SelectorBetslipRemoveBtn)
	if err == nil {
		parentContainer := removeBtn
		for i := 0; i < 5; i++ {
			if p, err := parentContainer.Parent(); err == nil && p != nil {
				parentContainer = p
			} else {
				break
			}
		}

		slipText, _ := parentContainer.Text()
		lineEls, _ := parentContainer.Elements(`[class*='line'], [class*='Line'], [class*='Points'], [class*='points'], [class*='selectionName']`)
		oddsEls, _ := parentContainer.Elements(`[class*='odds'], [class*='Odds']`)

		var tempLine, tempOdd float64
		reNum := regexp.MustCompile(`\d+(?:\.\d+)?`)

		if len(lineEls) > 0 {
			for _, el := range lineEls {
				t, _ := el.Text()
				m := reNum.FindString(t)
				if m != "" {
					if val, err := strconv.ParseFloat(m, 64); err == nil {
						tempLine = val
					}
				}
			}
		}
		if len(oddsEls) > 0 {
			for _, el := range oddsEls {
				t, _ := el.Text()
				t = strings.ReplaceAll(strings.ReplaceAll(t, "▲", ""), "▼", "")
				m := reNum.FindString(t)
				if m != "" {
					if val, err := strconv.ParseFloat(m, 64); err == nil {
						tempOdd = val
					}
				}
			}
		}

		if tempLine > 0 && tempOdd > 0 {
			slipLine = tempLine
			slipOdd = tempOdd
			foundSlipInfo = true
		} else {
			matches := reNum.FindAllString(strings.ReplaceAll(strings.ReplaceAll(slipText, "▲", ""), "▼", ""), -1)
			if len(matches) >= 2 {
				var floats []float64
				for _, m := range matches {
					if v, err := strconv.ParseFloat(m, 64); err == nil {
						floats = append(floats, v)
					}
				}
				for _, f := range floats {
					if math.Abs(f-tipLineF) <= 5.0 && (math.Mod(f, 0.25) == 0 || math.Mod(f, 0.5) == 0) {
						if slipLine == bestLine || math.Abs(f-tipLineF) < math.Abs(slipLine-tipLineF) {
							slipLine = f
						}
					}
					if math.Abs(f-tip.TargetOdd) <= 3.0 && f > 1.0 {
						if slipOdd == bestOdd || math.Abs(f-tip.TargetOdd) < math.Abs(slipOdd-tip.TargetOdd) {
							slipOdd = f
						}
					}
				}
				foundSlipInfo = true
			}
		}
	}

	if foundSlipInfo {
		ok, msg := ValidateBet(tip, slipLine, slipOdd)
		if !ok {
			log.Printf("⚠️ Aposta abortada na validação do bilhete: %s", msg)
			return fmt.Errorf("validação falhou: %s", msg)
		}
		log.Printf("✅ Validação bilhete: Linha %.2f, Odd %.2f", slipLine, slipOdd)
	}

	// Nova verificação de popups antes do clique final
	browser.CheckAndDismissPopups(page)

	placeBtn, err := frame.Element(config.SelectorBetslipPlaceBetBtn)
	if err != nil {
		return fmt.Errorf("botão de confirmação não encontrado no bilhete")
	}

	btnText, _ := placeBtn.Text()
	log.Printf("🚀 Aposta preparada! Odd: %.2f (Linha %.2f) | Clicando: [%s]", slipOdd, slipLine, strings.TrimSpace(btnText))
	if err := placeBtn.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("erro ao confirmar aposta no bilhete: %w", err)
	}

	log.Println("✅ Aposta confirmada com sucesso!")

	// Registra log
	matchName := fmt.Sprintf("%s x %s", tip.HomeTeam, tip.AwayTeam)
	if err := logger.LogBet(matchName, fmt.Sprintf("%.2f", slipLine), fmt.Sprintf("%.2f", slipOdd), amount); err != nil {
		log.Printf("⚠️ Erro ao salvar log: %v", err)
	}

	return nil
}
