package betting

import (
	"bolsadeaposta-bot/internal/config"
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// PrepareHandicapBet simula a adição de uma aposta ao boletim e preenchimento do valor.
// Não clica no botão final de confirmação de acordo com a solicitação do usuário.
func PrepareHandicapBet(page *rod.Page, teamName string, handicapLine string, amount string) error {
	fmt.Printf("🎯 Iniciando simulação de aposta: %s (%s) | Valor: %s\n", teamName, handicapLine, amount)

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

	fmt.Printf("✅ Odd encontrada. Adicionando ao boletim...\n")
	if err := targetOdd.ScrollIntoView(); err != nil {
		fmt.Printf("⚠️ Erro ao dar scroll na odd (prosseguindo): %v\n", err)
	}
	if err := targetOdd.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("erro ao clicar na odd: %w", err)
	}

	// 4. Aguarda e interage com o betslip
	fmt.Println("⏳ Aguardando boletim de apostas...")
	stakeInput, err := frame.Timeout(10 * time.Second).Element(config.SelectorBetslipStakeInput)
	if err != nil {
		return fmt.Errorf("campo de valor (stake) não encontrado no boletim: %w", err)
	}

	fmt.Printf("✍️ Preenchendo valor: %s\n", amount)
	if err := stakeInput.Input(amount); err != nil {
		return fmt.Errorf("erro ao preencher valor da aposta: %w", err)
	}

	// 5. Verifica presença do botão de confirmação
	placeBtn, err := frame.Element(config.SelectorBetslipPlaceBetBtn)
	if err == nil {
		btnText, _ := placeBtn.Text()
		fmt.Printf("🚀 Aposta preparada! Botão de confirmação visível: [%s]\n", strings.TrimSpace(btnText))
		fmt.Println("ℹ️ Simulação concluída. O botão final NÃO foi clicado.")
	} else {
		fmt.Println("⚠️ Botão de confirmação não identificado, mas a seleção foi adicionada.")
	}

	return nil
}

// ClearBetslip tenta remover todas as seleções atuais do boletim
func ClearBetslip(frame *rod.Page) error {
	removeButtons, err := frame.Elements(config.SelectorBetslipRemoveBtn)
	if err != nil || len(removeButtons) == 0 {
		return nil
	}

	fmt.Printf("🗑️ Limpando %d seleções anteriores do boletim...\n", len(removeButtons))
	for _, removeButton := range removeButtons {
		_ = removeButton.Click(proto.InputMouseButtonLeft, 1)
		time.Sleep(200 * time.Millisecond)
	}
	return nil
}
