package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"bolsadeaposta-bot/internal/browser"
	"bolsadeaposta-bot/internal/config"
	"bolsadeaposta-bot/internal/queue"
	"bolsadeaposta-bot/internal/telegram"
)

func main() {
	config.Load()

	browserInstance, _, err := browser.LoadPageFlow()
	if err != nil {
		fmt.Printf("❌ Erro fatal ao carregar o navegador: %v\n", err)
		os.Exit(1)
	}
	defer browserInstance.MustClose()

	log.Println("Navegador inicializado e logado (se necessário). Iniciando fila em background e bot do Telegram...")

	worker := queue.NewWorker(browserInstance)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := telegram.StartUserbot(ctx, worker); err != nil {
			log.Fatalf("Erro ao rodar cliente Telegram MTProto: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Encerrando bot e fechando navegador...")
	cancel() // signals MTProto to shut down cleanly
}
