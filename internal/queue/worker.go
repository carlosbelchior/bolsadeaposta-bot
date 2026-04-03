package queue

import (
	"bolsadeaposta-bot/internal/betting"
	"bolsadeaposta-bot/internal/browser"
	"bolsadeaposta-bot/internal/config"
	"bolsadeaposta-bot/internal/crawler"
	"bolsadeaposta-bot/internal/models"
	"context"
	"fmt"
	"log"
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
	log.Printf("📥 Nova tip recebida para processamento: %s vs %s", tip.HomeTeam, tip.AwayTeam)
	go w.processTipInDedicatedPage(tip)
}

func (w *Worker) processTipInDedicatedPage(tip *models.Tip) {
	log.Printf("🌐 Abrindo aba para o jogo: %s vs %s...", tip.HomeTeam, tip.AwayTeam)

	timeoutDuration := 10*time.Minute - time.Since(tip.CreatedAt)
	if timeoutDuration <= 0 {
		log.Printf("🗑️ Tip expirada (10min): %s vs %s", tip.HomeTeam, tip.AwayTeam)
		tip.Status = models.StatusCancelled
		return
	}

	page, err := crawler.NavigateToMatch(w.browser, tip.HomeTeam, tip.AwayTeam)
	if err != nil {
		log.Printf("❌ Falha ao navegar para o jogo %s vs %s: %v", tip.HomeTeam, tip.AwayTeam, err)
		tip.Status = models.StatusCancelled
		if page != nil {
			page.Close()
		}
		return
	}

	defer func() {
		log.Printf("🧹 Fechando aba dedicada: %s vs %s", tip.HomeTeam, tip.AwayTeam)
		page.Close()
	}()

	log.Printf("👁️ Monitorando partida: %s vs %s", tip.HomeTeam, tip.AwayTeam)

	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("🗑️ Tempo de validade da tip encerrado via Contexto: %s vs %s", tip.HomeTeam, tip.AwayTeam)
			tip.Status = models.StatusCancelled
			return

		case <-ticker.C:
			if tip.Status != models.StatusPending {
				return
			}

			browser.CheckAndDismissPopups(page)

			liveScore, err := crawler.FetchLiveScore(page)
			if err == nil && liveScore != "" {
				if !crawler.IsScoreMatch(liveScore, tip.Score) {
					log.Printf("⏳ [Aba %s] Placar atual (%s) diferente da tip (%s). Aguardando...", tip.HomeTeam, liveScore, tip.Score)
					continue
				}
			}

			log.Printf("🔍 [Aba %s] Checando odd atual...", tip.HomeTeam)

			err = betting.PrepareGoalBet(page, tip, fmt.Sprintf("%d", config.BetAmount))
			if err != nil {
				log.Printf("🔄 [Aba %s] Critérios não atendidos: %v | Tentando novamente...", tip.HomeTeam, err)
			} else {
				log.Printf("✅🚀 [Aba %s] Aposta finalizada com sucesso!", tip.HomeTeam)
				tip.Status = models.StatusPlaced
				return
			}
		}
	}
}
