package crawler

import (
	"testing"
)

func TestParseLiveScore(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "Standard newline",
			text: "1\n2",
			want: "1-2",
		},
		{
			name: "With spaces and newline",
			text: " 3 \n 0 ",
			want: "3-0",
		},
		{
			name: "Hyphen format",
			text: "1 - 1",
			want: "1 - 1",
		},
		{
			name: "Empty string",
			text: "",
			want: "",
		},
		{
			name: "Single value",
			text: "5",
			want: "5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseLiveScore(tt.text); got != tt.want {
				t.Errorf("ParseLiveScore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsMatchTarget(t *testing.T) {
	tests := []struct {
		name           string
		extractedTeam1 string
		extractedTeam2 string
		target1        string
		target2        string
		wantMatched     bool
	}{
		{
			name:           "Exact match same order",
			extractedTeam1: "Habibi",
			extractedTeam2: "Jose",
			target1:        "Habibi",
			target2:        "Jose",
			wantMatched:    true,
		},
		{
			name:           "Exact match swapped order",
			extractedTeam1: "Jose",
			extractedTeam2: "Habibi",
			target1:        "Habibi",
			target2:        "Jose",
			wantMatched:    true,
		},
		{
			name:           "Partial match same order",
			extractedTeam1: "Sporting (Habibi)",
			extractedTeam2: "Galatasaray (Jose)",
			target1:        "Habibi",
			target2:        "Jose",
			wantMatched:    true,
		},
		{
			name:           "Case insensitive match",
			extractedTeam1: "HABIBI",
			extractedTeam2: "jose",
			target1:        "habibi",
			target2:        "JOSE",
			wantMatched:    true,
		},
		{
			name:           "Mismatch",
			extractedTeam1: "Habibi",
			extractedTeam2: "Jose",
			target1:        "Habibi",
			target2:        "Carlos",
			wantMatched:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, _, _ := IsMatchTarget(tt.extractedTeam1, tt.extractedTeam2, tt.target1, tt.target2)
			if matched != tt.wantMatched {
				t.Errorf("%s: IsMatchTarget() matched = %v, want %v", tt.name, matched, tt.wantMatched)
			}
		})
	}
}

func TestIsScoreMatch(t *testing.T) {
	tests := []struct {
		name    string
		current string
		target  string
		want    bool
	}{
		{
			name:    "Exact match",
			current: "1-0",
			target:  "1-0",
			want:    true,
		},
		{
			name:    "Match with spaces",
			current: " 1 - 0 ",
			target:  "1-0",
			want:    true,
		},
		{
			name:    "Mismatch",
			current: "2-0",
			target:  "1-0",
			want:    false,
		},
		{
			name:    "Empty target always matches",
			current: "1-0",
			target:  "",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsScoreMatch(tt.current, tt.target); got != tt.want {
				t.Errorf("%s: IsScoreMatch() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
