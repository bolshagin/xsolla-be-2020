package model

import (
	"time"
)

type Session struct {
	SessionID    uint      `json:"-"`
	SessionToken string    `json:"session_token,omitempty"`
	Amount       float64   `json:"amount"`
	Purpose      string    `json:"purpose"`
	CreatedAt    time.Time `json:"created_at"`
	ClosedAt     time.Time `json:"closed_at,omitempty"`
}
