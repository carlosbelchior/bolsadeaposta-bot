package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLogBet(t *testing.T) {
	// 1. Setup - Create a temporary directory for tests
	tempDir, err := os.MkdirTemp("", "bet_logs_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Mock the LogFolder to use our temp dir
	oldFolder := LogFolder
	LogFolder = tempDir
	defer func() { LogFolder = oldFolder }()

	// 2. Data to log
	match := "Habibi vs Jose"
	line := "Mais de 1.5"
	odd := "1.75"
	amount := "5.00"

	// 3. Execution
	err = LogBet(match, line, odd, amount)
	if err != nil {
		t.Errorf("LogBet returned an error: %v", err)
	}

	// 4. Verification
	today := time.Now().Format("2006-01-02")
	expectedFile := filepath.Join(tempDir, fmt.Sprintf("%s.log", today))

	// Check if file exists
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("Log file was not created: %s", expectedFile)
	}

	// Read file content
	content, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logLine := string(content)
	// Example: 2026-04-03 23:51:00;Habibi vs Jose;Mais de 1.5;1.75;5.00
	if !strings.Contains(logLine, match) ||
		!strings.Contains(logLine, line) ||
		!strings.Contains(logLine, odd) ||
		!strings.Contains(logLine, amount) {
		t.Errorf("Log line content is incorrect: %s", logLine)
	}

	// Verify format (separated by ;)
	parts := strings.Split(strings.TrimSpace(logLine), ";")
	if len(parts) != 5 {
		t.Errorf("Expected 5 parts in log line, got %d", len(parts))
	}
}
