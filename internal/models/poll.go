package models

import "time"

type Poll struct {
    ID        string    `json:"id"`
    Question  string    `json:"question"`
    Options   []string  `json:"options"`
    Votes     []int     `json:"votes"`
    ExpiresAt time.Time `json:"expires_at"`
}
