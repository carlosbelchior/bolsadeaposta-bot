package config

import (
	"os"
	"testing"
)

func TestConfigLoad_Success(t *testing.T) {
	os.Clearenv()
	// Set required environment variables
	os.Setenv("TELEGRAM_API_ID", "12345")
	os.Setenv("TELEGRAM_API_HASH", "abcdef")
	os.Setenv("TELEGRAM_TARGET_USERNAME", "@testbot")
	os.Setenv("BET_AMOUNT", "10")
	os.Setenv("TARGET_LEAGUE_NAME", "Test League")
	os.Setenv("TARGET_IFRAME_DOMAIN", "test.io")

	err := Load()
	if err != nil {
		t.Fatalf("Load failed unexpectedly: %v", err)
	}

	if TelegramAPIID != 12345 {
		t.Errorf("TelegramAPIID incorrect: got %d, want 12345", TelegramAPIID)
	}
	if TelegramAPIHash != "abcdef" {
		t.Errorf("TelegramAPIHash incorrect: got %s, want abcdef", TelegramAPIHash)
	}
	if BetAmount != 10 {
		t.Errorf("BetAmount incorrect: got %d, want 10", BetAmount)
	}
	if TargetLeagueName != "Test League" {
		t.Errorf("TargetLeagueName incorrect: got %s, want 'Test League'", TargetLeagueName)
	}
	if TargetIframeDomain != "test.io" {
		t.Errorf("TargetIframeDomain incorrect: got %s, want 'test.io'", TargetIframeDomain)
	}
}

func TestConfigLoad_Errors(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
	}{
		{
			name: "Missing TELEGRAM_API_ID",
			env: map[string]string{
				"TELEGRAM_API_HASH":        "abc",
				"TELEGRAM_TARGET_USERNAME": "@bot",
				"BET_AMOUNT":              "10",
			},
		},
		{
			name: "Invalid TELEGRAM_API_ID",
			env: map[string]string{
				"TELEGRAM_API_ID":          "not-an-int",
				"TELEGRAM_API_HASH":        "abc",
				"TELEGRAM_TARGET_USERNAME": "@bot",
				"BET_AMOUNT":              "10",
			},
		},
		{
			name: "Missing TELEGRAM_API_HASH",
			env: map[string]string{
				"TELEGRAM_API_ID":          "123",
				"TELEGRAM_TARGET_USERNAME": "@bot",
				"BET_AMOUNT":              "10",
			},
		},
		{
			name: "Missing TELEGRAM_TARGET_USERNAME",
			env: map[string]string{
				"TELEGRAM_API_ID":   "123",
				"TELEGRAM_API_HASH": "abc",
				"BET_AMOUNT":        "10",
			},
		},
		{
			name: "Missing BET_AMOUNT",
			env: map[string]string{
				"TELEGRAM_API_ID":          "123",
				"TELEGRAM_API_HASH":        "abc",
				"TELEGRAM_TARGET_USERNAME": "@bot",
			},
		},
		{
			name: "Invalid BET_AMOUNT",
			env: map[string]string{
				"TELEGRAM_API_ID":          "123",
				"TELEGRAM_API_HASH":        "abc",
				"TELEGRAM_TARGET_USERNAME": "@bot",
				"BET_AMOUNT":              "abc",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			for k, v := range tt.env {
				os.Setenv(k, v)
			}
			err := Load()
			if err == nil {
				t.Errorf("%s: Expected error but got nil", tt.name)
			}
		})
	}
}
