package config

import (
	"github.com/joho/godotenv"
)

// Load initializes all configuration structures
func Load() {
	_ = godotenv.Load() // Ignore error as .env may not exist in production
	
	loadTelegramConfig()
	loadBolsaConfig()
}
