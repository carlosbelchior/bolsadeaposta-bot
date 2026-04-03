package betting

import (
	"bolsadeaposta-bot/internal/config"
	"bolsadeaposta-bot/internal/logger"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// PrepareHandicapBet simula a adição de uma aposta ao boletim e preenchimento do valor.
// Clica no botão final de confirmação.
func PrepareHandicapBet(page *rod.Page, teamName string, handicapLine string, amount string) error {
	log.Printf("🎯 Iniciando simulação de aposta: %s (%s) | Valor: %s", teamName, handicapLine, amount)

	// 1. Localiza o iframe do Sportsbook
	sportsbookIframeElement, err := page.Element(config.SelectorIframeFSSB)
	if err != nil {
		return fmt.Errorf("iframe do sportsbook não encontrado: %w", err)
	}

	frame, err := sportsbookIframeElement.Frame()
	if err != nil {
		return fmt.Errorf("não foi possível acessar o contexto do iframe: %w", err)
	}

	// 2. Limpa o boletim atual para evitar conflitos (opcional mas recomendado)
	_ = ClearBetslip(frame)

	// 3. Localiza e clica na odd correta
	odds, err := frame.Elements(config.SelectorAHButton)
	if err != nil {
		return fmt.Errorf("não foi possível encontrar botões de odds: %w", err)
	}

	var targetOdd *rod.Element
	for _, odd := range odds {
		text, _ := odd.Text()
		lowerText := strings.ToLower(text)
		if strings.Contains(lowerText, strings.ToLower(teamName)) && strings.Contains(lowerText, handicapLine) {
			targetOdd = odd
			break
		}
	}

	if targetOdd == nil {
		return fmt.Errorf("odd para %s com linha %s não encontrada", teamName, handicapLine)
	}

	log.Println("✅ Odd encontrada. Adicionando ao boletim...")
	if err := targetOdd.ScrollIntoView(); err != nil {
		log.Printf("⚠️ Erro ao dar scroll na odd (prosseguindo): %v", err)
	}
	if err := targetOdd.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("erro ao clicar na odd: %w", err)
	}

	// 4. Aguarda e interage com o betslip
	log.Println("⏳ Aguardando boletim de apostas...")
	stakeInput, err := frame.Timeout(10 * time.Second).Element(config.SelectorBetslipStakeInput)
	if err != nil {
		return fmt.Errorf("campo de valor (stake) não encontrado no boletim: %w", err)
	}

	log.Printf("✍️ Preenchendo valor: %s", amount)
	if err := stakeInput.Input(amount); err != nil {
		return fmt.Errorf("erro ao preencher valor da aposta: %w", err)
	}

	// 5. Verifica presença do botão de confirmação e clica
	placeBtn, err := frame.Element(config.SelectorBetslipPlaceBetBtn)
	if err == nil {
		btnText, _ := placeBtn.Text()
		log.Printf("🚀 Aposta preparada! Clicando no botão: [%s]", strings.TrimSpace(btnText))
		if err := placeBtn.Click(proto.InputMouseButtonLeft, 1); err != nil {
			return fmt.Errorf("erro ao confirmar aposta no boletim: %w", err)
		}
		log.Println("✅ Aposta confirmada com sucesso!")
		
		// Registrar aposta no log diário
		oddValStr := "-"
		text, _ := targetOdd.Text()
		lines := strings.Split(text, "\n")
		if len(lines) > 0 {
			oddValStr = strings.TrimSpace(lines[len(lines)-1])
		}
		
		if err := logger.LogBet(teamName, handicapLine, oddValStr, amount); err != nil {
			log.Printf("⚠️ Erro ao salvar log da aposta: %v", err)
		}

	} else {
		return fmt.Errorf("botão de confirmação não identificado no boletim")
	}

	return nil
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
