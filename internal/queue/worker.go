package queue

import (
	"bolsadeaposta-bot/internal/betting"
	"bolsadeaposta-bot/internal/config"
	"bolsadeaposta-bot/internal/crawler"
	"bolsadeaposta-bot/internal/models"
	"context"
	"log"
	"strings"
	"time"

	"github.com/go-rod/rod"
)

type Worker struct {
	browser *rod.Browser
}

func NewWorker(browser *rod.Browser) *Worker {
	return &Worker{
		browser: browser,
	}
}

// AddTip instantly spawns a dedicated goroutine and tab for the Tip
func (w *Worker) AddTip(tip *models.Tip) {
	log.Printf("📥 Nova tip distribuída para fila assíncrona (Aba Dedicada): %s vs %s", tip.Team1, tip.Team2)
	go w.processTipInDedicatedPage(tip)
}

func (w *Worker) processTipInDedicatedPage(tip *models.Tip) {
	log.Printf("🌐 Abrindo aba e viajando até o jogo: %s vs %s...", tip.Team1, tip.Team2)

	timeoutDuration := 10*time.Minute - time.Since(tip.CreatedAt)
	if timeoutDuration <= 0 {
		log.Printf("🗑️ Tip já expirou (10min) antes mesmo do loop rodar: %s vs %s", tip.Team1, tip.Team2)
		tip.Status = models.StatusCancelled
		return
	}

	page, err := crawler.NavigateToMatch(w.browser, tip.Team1, tip.Team2)
	if err != nil {
		log.Printf("❌ Falha crítica ao isolar o jogo %s vs %s: %v", tip.Team1, tip.Team2, err)
		tip.Status = models.StatusCancelled
		if page != nil {
			page.Close()
		}
		return
	}

	defer func() {
		log.Printf("🧹 Fechando a aba dedicada do jogo %s vs %s", tip.Team1, tip.Team2)
		page.Close()
	}()

	log.Printf("👁️ Olhos fixos no jogo %s vs %s. Monitorando a cada 2 segundos...", tip.Team1, tip.Team2)

	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Polling Loop with Select
	for {
		select {
		case <-ctx.Done():
			log.Printf("🗑️ Tip expirou o tempo de validade via Contexto! %s vs %s. Cancelando aposta e fechando aba.", tip.Team1, tip.Team2)
			tip.Status = models.StatusCancelled
			return

		case <-ticker.C:
			if tip.Status != models.StatusPending {
				return
			}

			// Optional: We can check live score on the dedicated page to validate
			liveScore, err := crawler.FetchLiveScore(page)
			if err == nil && liveScore != "" {
				liveScoreFmt := strings.ReplaceAll(liveScore, " ", "")
				if tip.Score != "" && liveScoreFmt != tip.Score {
					log.Printf("⏳ [Aba #%s] Placar atual da partida (%s) diferente da tip (%s). Aguardando...", tip.Team1, liveScoreFmt, tip.Score)
					continue
				}
			}

			// Attempt Goal Bet validation directly
			log.Printf("🔍 [Aba #%s] Checando odd atual...", tip.Team1)

			err = betting.PrepareGoalBet(page, tip, config.DefaultStake)
			if err != nil {
				log.Printf("🔄 [Aba #%s] Odd insuficiente ou mercado não disponível. Motivo: %v | Tentando novamente no próximo tick.", tip.Team1, err)
			} else {
				log.Printf("✅🚀 [Aba #%s] Aposta finalizada com sucesso! Odd cobriu a da Tip.", tip.Team1)
				tip.Status = models.StatusPlaced
				return // Retorna para acionar o fechamento da aba no defer
			}
		}
	}
}
