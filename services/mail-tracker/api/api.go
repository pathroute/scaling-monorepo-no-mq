// Package api defines the mail-tracker service's wire types; other services import it.
package api

import "time"

// Message is a tracked outgoing email.
type Message struct {
	ID       string    `json:"id"`
	Org      string    `json:"org"`
	To       string    `json:"to"`
	Subject  string    `json:"subject"`
	Body     string    `json:"body"`
	Opens    int       `json:"opens"`
	Category string    `json:"category,omitempty"`
	SentAt   time.Time `json:"sent_at"`
}

// Track is the body of POST /messages.
type Track struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// Categories a message can be classified into.
var Categories = []string{"sales", "support", "billing", "other"}
