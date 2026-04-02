package queue

import (
	"bolsadeaposta-bot/internal/crawler"
	"bolsadeaposta-bot/internal/models"
	"bolsadeaposta-bot/internal/betting"
	"bolsadeaposta-bot/internal/config"
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

	// Polling Loop
	for {
		if tip.Status != models.StatusPending {
			break
		}

		if time.Since(tip.CreatedAt) > 10*time.Minute {
			log.Printf("🗑️ Tip expirou (10min passados)! %s vs %s. Cancelando aposta e fechando aba.", tip.Team1, tip.Team2)
			tip.Status = models.StatusCancelled
			break
		}

		// Optional: We can check live score on the dedicated page to validate
		liveScore, err := crawler.FetchLiveScore(page)
		if err == nil && liveScore != "" {
			liveScoreFmt := strings.ReplaceAll(liveScore, " ", "")
			if tip.Score != "" && liveScoreFmt != tip.Score {
				log.Printf("⏳ [Aba #%s] Placar atual da partida (%s) diferente da tip (%s). Aguardando... (Tentativas duram 10min)", tip.Team1, liveScoreFmt, tip.Score)
				time.Sleep(2 * time.Second)
				continue
			}
		}

		// Attempt Goal Bet validation directly
		log.Printf("🔍 [Aba #%s] Checando odd atual...", tip.Team1)
		
		err = betting.PrepareGoalBet(page, tip, config.DefaultStake)
		if err != nil {
			log.Printf("🔄 [Aba #%s] Odd insuficiente ou mercado não disponível. Motivo: %v | Tentando novamente em 2s.", tip.Team1, err)
		} else {
			log.Printf("✅🚀 [Aba #%s] Aposta finalizada com sucesso! Odd cobriu a da Tip.", tip.Team1)
			tip.Status = models.StatusPlaced
			break // Sai do loop para fechar a aba
		}

		time.Sleep(2 * time.Second)
	}
}
