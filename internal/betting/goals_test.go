package betting

import (
	"bolsadeaposta-bot/internal/models"
	"testing"
)

func TestParseOddString(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		wantLine float64
		wantOdd  float64
		wantOk   bool
	}{
		{
			name:     "Multi-line with arrows",
			text:     "Mais de 4.5\n▲1.71",
			wantLine: 4.5,
			wantOdd:  1.71,
			wantOk:   true,
		},
		{
			name:     "Combo format",
			text:     "4.5 1.80",
			wantLine: 4.5,
			wantOdd:  1.80,
			wantOk:   true,
		},
		{
			name:     "Combo format with spaces",
			text:     "1.5   1.95",
			wantLine: 1.5,
			wantOdd:  1.95,
			wantOk:   true,
		},
		{
			name:     "Single line variety",
			text:     "Over 2.5 @ 1.95",
			wantLine: 2.5,
			wantOdd:  1.95,
			wantOk:   true,
		},
		{
			name:     "No numbers",
			text:     "No numbers here",
			wantLine: 0,
			wantOdd:  0,
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLine, gotOdd, gotOk := ParseOddString(tt.text)
			if gotLine != tt.wantLine || gotOdd != tt.wantOdd || gotOk != tt.wantOk {
				t.Errorf("%s: got (%v, %v, %v), want (%v, %v, %v)", tt.name, gotLine, gotOdd, gotOk, tt.wantLine, tt.wantOdd, tt.wantOk)
			}
		})
	}
}

func TestValidateBet(t *testing.T) {
	tipOver := &models.Tip{Market: "Mais de 2.5", Line: "2.5", TargetOdd: 1.70}
	tipUnder := &models.Tip{Market: "Menos de 2.5", Line: "2.5", TargetOdd: 1.70}

	tests := []struct {
		name     string
		tip      *models.Tip
		slipLine float64
		slipOdd  float64
		wantOk   bool
	}{
		{
			name:     "Over: Better line (lower), Better odd",
			tip:      tipOver,
			slipLine: 2.0,
			slipOdd:  1.80,
			wantOk:   true,
		},
		{
			name:     "Over: Same line, same odd",
			tip:      tipOver,
			slipLine: 2.5,
			slipOdd:  1.70,
			wantOk:   true,
		},
		{
			name:     "Over: Worse line (higher), Better odd",
			tip:      tipOver,
			slipLine: 3.5,
			slipOdd:  1.80,
			wantOk:   false,
		},
		{
			name:     "Over: Same line, Worse odd",
			tip:      tipOver,
			slipLine: 2.5,
			slipOdd:  1.60,
			wantOk:   false,
		},
		{
			name:     "Under: Better line (higher), Better odd",
			tip:      tipUnder,
			slipLine: 3.0,
			slipOdd:  1.80,
			wantOk:   true,
		},
		{
			name:     "Under: Same line, same odd",
			tip:      tipUnder,
			slipLine: 2.5,
			slipOdd:  1.70,
			wantOk:   true,
		},
		{
			name:     "Under: Worse line (lower), Better odd",
			tip:      tipUnder,
			slipLine: 1.5,
			slipOdd:  1.80,
			wantOk:   false,
		},
		{
			name:     "Unknown Market",
			tip:      &models.Tip{Market: "Winning Team", Line: "1", TargetOdd: 1.70},
			slipLine: 1,
			slipOdd:  1.8,
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOk, _ := ValidateBet(tt.tip, tt.slipLine, tt.slipOdd)
			if gotOk != tt.wantOk {
				t.Errorf("%s: got %v, want %v", tt.name, gotOk, tt.wantOk)
			}
		})
	}
}
