package config

import (
	"fmt"
	"os"
	"strconv"
)

var (
	TelegramAPIID          int
	TelegramAPIHash        string
	TelegramTargetUsername string // Apenas ouvirá desta peer (ex: @botprovedor ou o ID do chat)
)

func loadTelegramConfig() error {
	// API ID
	apiIDStr := os.Getenv("TELEGRAM_API_ID")
	if apiIDStr == "" {
		return fmt.Errorf("TELEGRAM_API_ID is not set")
	}
	id, err := strconv.Atoi(apiIDStr)
	if err != nil {
		return fmt.Errorf("TELEGRAM_API_ID must be an integer")
	}
	TelegramAPIID = id

	// API HASH
	TelegramAPIHash = os.Getenv("TELEGRAM_API_HASH")
	if TelegramAPIHash == "" {
		return fmt.Errorf("TELEGRAM_API_HASH is not set")
	}

	// Target Bot Username/Peer
	TelegramTargetUsername = os.Getenv("TELEGRAM_TARGET_USERNAME")
	if TelegramTargetUsername == "" {
		return fmt.Errorf("TELEGRAM_TARGET_USERNAME is not set")
	}
	return nil
}
