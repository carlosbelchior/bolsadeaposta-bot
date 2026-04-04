package telegram

import (
	"testing"
)

func TestParseTipMessage_Success(t *testing.T) {
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

	assertEquals(t, "TargetOdd", tip.TargetOdd, 1.71)
	assertEquals(t, "Market", tip.Market, "Mais de 4.5 - Gols")
	assertEquals(t, "Line", tip.Line, "4.5")
	assertEquals(t, "Score", tip.Score, "0-0")
	assertEquals(t, "HomeTeam", tip.HomeTeam, "Habibi")
	assertEquals(t, "AwayTeam", tip.AwayTeam, "Jose")
}

func TestParseTipMessage_Under(t *testing.T) {
	msg := `📈Menos de 2.5 - Gols @1.90
🏃Team A (L) vs Team B (R)
`
	tip, err := ParseTipMessage(msg)
	if err != nil {
		t.Fatalf("Erro inesperado: %v", err)
	}

	assertEquals(t, "Line", tip.Line, "2.5")
	assertEquals(t, "Market", tip.Market, "Menos de 2.5 - Gols")
}

func TestParseTipMessage_NoEmoji(t *testing.T) {
	msg := `Mais de 4.5 - Gols @1.71`
	_, err := ParseTipMessage(msg)
	if err == nil {
		t.Error("Deveria ter retornado erro por falta de emoji 📈")
	}
}

func TestParseTipMessage_MissingTeams(t *testing.T) {
	msg := `📈Mais de 4.5 - Gols @1.71
Score: 0-0
`
	_, err := ParseTipMessage(msg)
	if err == nil {
		t.Error("Deveria ter retornado erro por falta dos times 🏃")
	}
}

func TestParseTipMessage_InvalidOdd(t *testing.T) {
	msg := `📈Mais de 4.5 - Gols @abc
🏃Habibi (S) vs Jose (G)
`
	_, err := ParseTipMessage(msg)
	if err == nil {
		t.Error("Deveria ter retornado erro por odd inválida")
	}
}

func assertEquals(t *testing.T, field string, got, want interface{}) {
	t.Helper()
	if got != want {
		t.Errorf("%s incorreto: got %v, want %v", field, got, want)
	}
}
