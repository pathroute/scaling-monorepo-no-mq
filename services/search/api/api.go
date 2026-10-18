// Package api defines the search service's wire types.
package api

// Hit is one search result.
type Hit struct {
	Kind  string `json:"kind"` // todo, issue or mail
	ID    string `json:"id"`
	Title string `json:"title"`
	Score int    `json:"score"`
}
