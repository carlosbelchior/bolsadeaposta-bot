package config

import (
	"log"
	"os"
	"strconv"
)

var (
	TelegramAPIID          int
	TelegramAPIHash        string
	TelegramTargetUsername string // Apenas ouvirá desta peer (ex: @botprovedor ou o ID do chat)
)

func loadTelegramConfig() {
	// API ID
	apiIDStr := os.Getenv("TELEGRAM_API_ID")
	if apiIDStr == "" {
		log.Fatal("TELEGRAM_API_ID is not set")
	}
	id, err := strconv.Atoi(apiIDStr)
	if err != nil {
		log.Fatalf("TELEGRAM_API_ID deve ser um número inteiro")
	}
	TelegramAPIID = id

	// API HASH
	TelegramAPIHash = os.Getenv("TELEGRAM_API_HASH")
	if TelegramAPIHash == "" {
		log.Fatal("TELEGRAM_API_HASH is not set")
	}

	// Target Bot Username/Peer
	TelegramTargetUsername = os.Getenv("TELEGRAM_TARGET_USERNAME")
	if TelegramTargetUsername == "" {
		log.Fatal("TELEGRAM_TARGET_USERNAME is not set (indique de quem vou ler as mensagens)")
	}
}
