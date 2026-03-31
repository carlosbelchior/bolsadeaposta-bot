package models

type HandicapLine struct {
	Line string
	Odd  string
}

type Match struct {
	Team1       string
	Team2       string
	Score1      string
	Score2      string
	Time        string
	Odd1        string
	OddX        string
	Odd2        string
	HTHandicap1 []HandicapLine
	HTHandicap2 []HandicapLine
}
