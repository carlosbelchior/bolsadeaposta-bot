package config

import (
	"github.com/joho/godotenv"
)

// Load initializes all configuration structures. It returns an error if a required environment variable is missing or invalid.
func Load() error {
	_ = godotenv.Load() // Ignore error as .env may not exist in production
	
	if err := loadTelegramConfig(); err != nil {
		return err
	}
	if err := loadBolsaConfig(); err != nil {
		return err
	}
	return nil
}
