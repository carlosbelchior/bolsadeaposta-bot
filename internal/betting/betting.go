package betting

import (
	"bolsadeaposta-bot/internal/config"
	"fmt"
	"log"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// GetBetslipFrame localiza o iframe do Sportsbook e retorna seu contexto
func GetBetslipFrame(page *rod.Page) (*rod.Page, error) {
	el, err := page.Element(config.SelectorIframeFSSB)
	if err != nil {
		return nil, fmt.Errorf("iframe do sportsbook não encontrado: %w", err)
	}
	frame, err := el.Frame()
	if err != nil {
		return nil, fmt.Errorf("erro ao acessar contexto do iframe: %w", err)
	}
	return frame, nil
}

// PerformBetslipFlow limpa o boletim, clica na seleção e preenche o valor
func PerformBetslipFlow(frame *rod.Page, selection *rod.Element, amount string) error {
	_ = ClearBetslip(frame)

	log.Println("✅ Seleção encontrada. Adicionando ao boletim...")
	if err := selection.ScrollIntoView(); err != nil {
		log.Printf("⚠️ Erro ao dar scroll na seleção: %v", err)
	}
	if err := selection.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("erro ao clicar na seleção: %w", err)
	}

	log.Println("⏳ Aguardando boletim de apostas...")
	input, err := frame.Timeout(10 * time.Second).Element(config.SelectorBetslipStakeInput)
	if err != nil {
		return fmt.Errorf("campo de valor (stake) não encontrado: %w", err)
	}

	log.Printf("✍️ Preenchendo valor: %s", amount)
	return input.Input(amount)
}

// ClearBetslip tenta remover todas as seleções atuais do boletim
func ClearBetslip(frame *rod.Page) error {
	removeButtons, err := frame.Elements(config.SelectorBetslipRemoveBtn)
	if err != nil || len(removeButtons) == 0 {
		return nil
	}

	log.Printf("🗑️ Limpando %d seleções anteriores do boletim...", len(removeButtons))
	for _, removeButton := range removeButtons {
		_ = removeButton.Click(proto.InputMouseButtonLeft, 1)
		time.Sleep(200 * time.Millisecond)
	}
	return nil
}
