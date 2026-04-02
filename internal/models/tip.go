package models

import "time"

type TipStatus string

const (
	StatusPending   TipStatus = "PENDING"
	StatusPlaced    TipStatus = "PLACED"
	StatusCancelled TipStatus = "CANCELLED"
)

type Tip struct {
	ID        string
	Market    string // e.g. "Mais de 4.5 - Gols"
	Line      string // e.g. "4.5"
	TargetOdd float64 // e.g. 1.71
	Score     string // e.g. "0-0"
	Team1     string // e.g. "Habibi"
	Team2     string // e.g. "Jose"
	Status    TipStatus
	CreatedAt time.Time
}
