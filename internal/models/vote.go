package models

import "time"

type Vote struct {
    PollID  string    `json:"poll_id"`
    Option  string    `json:"option"`
    VotedAt time.Time `json:"voted_at"`
}
