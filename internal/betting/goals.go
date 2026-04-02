package betting

import (
	"bolsadeaposta-bot/internal/config"
	"bolsadeaposta-bot/internal/models"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

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

	// Since we are inside the match, odds should be visible. We need to find the odd button.
	// For "Mais de 4.5", usually the button has the number 4.5 and the odd.
	// We'll search elements.
	odds, err := frame.Elements(config.SelectorAHButton)
	if err != nil {
		return fmt.Errorf("não encontrou botões")
	}

	var targetOddElement *rod.Element
	var currentOddValue float64

	for _, oddEl := range odds {
		text, _ := oddEl.Text()
		lowerText := strings.ToLower(text)
		
		// Se tip.Line for "4.5", buscaremos "4.5"
		if strings.Contains(lowerText, tip.Line) {
			// Extract odd value from the button
			oddLines := strings.Split(text, "\n")
			for _, lineStr := range oddLines {
				lineStr = strings.TrimSpace(lineStr)
				if val, err := strconv.ParseFloat(lineStr, 64); err == nil && val != tip.TargetOdd && val < 50.0 {
					currentOddValue = val
				}
			}
			
			// Simple fallback if odd extraction is tricky
			targetOddElement = oddEl
			break
		}
	}

	if targetOddElement == nil {
		return fmt.Errorf("linha %s não encontrada. Disponíveis podem ter mudado", tip.Line)
	}

	// Validation
	if currentOddValue > 0 && currentOddValue < tip.TargetOdd {
		return fmt.Errorf("odd atual (%.2f) está Menor que a Odd da Tip (%.2f)", currentOddValue, tip.TargetOdd)
	}

	_ = targetOddElement.ScrollIntoView()
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
		fmt.Printf("🚀 Aposta preparada! Botão de confirmação visível: [%s]\n", strings.TrimSpace(btnText))
		fmt.Println("ℹ️ O botão final NÃO foi clicado pelo simulador.")
	}

	return nil
}
