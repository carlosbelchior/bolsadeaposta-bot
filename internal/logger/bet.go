package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var (
	// LogFolder defines where the log files will be stored.
	LogFolder = "logs"
)

// LogBet salva o registro de uma aposta realizada no arquivo de log do dia.
// Formato: data;jogador1 x jogador2;linha;odd;valor
func LogBet(match, line, odd, amount string) error {
	folder := LogFolder
	if err := os.MkdirAll(folder, 0755); err != nil {
		return fmt.Errorf("erro ao criar pasta de logs: %w", err)
	}

	today := time.Now().Format("2006-01-02")
	fileName := filepath.Join(folder, fmt.Sprintf("%s.log", today))
	
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo de log: %w", err)
	}
	defer f.Close()

	dateStr := time.Now().Format("2006-01-02 15:04:05")
	logLine := fmt.Sprintf("%s;%s;%s;%s;%s\n", dateStr, match, line, odd, amount)

	if _, err := f.WriteString(logLine); err != nil {
		return fmt.Errorf("erro ao escrever no arquivo de log: %w", err)
	}

	return nil
}
