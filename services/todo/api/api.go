// Package api defines the todo service's wire types; other services import it.
package api

import "time"

// Todo is one item on an org's list.
type Todo struct {
	ID        string    `json:"id"`
	Org       string    `json:"org"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateTodo is the body of POST /todos.
type CreateTodo struct {
	Title string `json:"title"`
}

// Suggestion is the response of POST /todos/{id}/suggest.
type Suggestion struct {
	Subtasks []string `json:"subtasks"`
}
