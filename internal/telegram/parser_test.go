package telegram

import (
	"testing"
)

func TestParseTipMessage(t *testing.T) {
	msg := `📈Mais de 4.5 - Gols @1.71

🔴Ao vivo
🕓Tempo da Tip: 2m16s - T1
🅿️Placar da Tip: 0-0
🅿️Placar final: Habibi 2 - 0 Jose
🏃Habibi (Sporting) vs Jose (Galatasaray)

⚽️Último confronto:
🔵 *Jose* 3 : 3 *Habibi* 🔵

🤖 GT Over FT
`
	tip, err := ParseTipMessage(msg)
	if err != nil {
		t.Fatalf("Erro inesperado: %v", err)
	}

	if tip.TargetOdd != 1.71 {
		t.Errorf("TargetOdd incorreta: got %v, want %v", tip.TargetOdd, 1.71)
	}

	if tip.Market != "Mais de 4.5 - Gols" {
		t.Errorf("Market incorreto: got %v, want %v", tip.Market, "Mais de 4.5 - Gols")
	}

	if tip.Line != "4.5" {
		t.Errorf("Line incorreto: got %v, want %v", tip.Line, "4.5")
	}

	if tip.Score != "0-0" {
		t.Errorf("Score incorreto: got %v, want %v", tip.Score, "0-0")
	}

	if tip.Team1 != "Habibi" {
		t.Errorf("Team1 incorreto: got %v, want %v", tip.Team1, "Habibi")
	}

	if tip.Team2 != "Jose" {
		t.Errorf("Team2 incorreto: got %v, want %v", tip.Team2, "Jose")
	}
}
