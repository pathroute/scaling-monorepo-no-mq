// Package api defines the issues service's wire types; other services import it.
package api

import "time"

// Status of an issue.
type Status string

const (
	StatusOpen       Status = "open"
	StatusInProgress Status = "in_progress"
	StatusClosed     Status = "closed"
)

// Issue is one tracked issue.
type Issue struct {
	ID        string    `json:"id"`
	Org       string    `json:"org"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateIssue is the body of POST /issues.
type CreateIssue struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// Transition is the body of POST /issues/{id}/status.
type Transition struct {
	Status Status `json:"status"`
}

// Summary is the response of POST /issues/{id}/summarize.
type Summary struct {
	Text string `json:"text"`
}
