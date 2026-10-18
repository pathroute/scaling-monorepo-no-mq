// Package search indexes todos, issues and mail and answers keyword queries.
package search

import (
	"cmp"
	"slices"
	"strings"
	"sync"
	"unicode"

	issues "github.com/pathroute/suite/services/issues/api"
	mail "github.com/pathroute/suite/services/mail-tracker/api"
	"github.com/pathroute/suite/services/search/api"
	todo "github.com/pathroute/suite/services/todo/api"
)

type doc struct {
	org   string
	hit   api.Hit
	terms map[string]int
}

// Index is an in-memory inverted index, scoped by org.
type Index struct {
	mu   sync.RWMutex
	docs map[string]doc            // kind/id -> doc
	post map[string]map[string]int // term -> kind/id -> count
}

// NewIndex returns an empty index.
func NewIndex() *Index {
	return &Index{docs: map[string]doc{}, post: map[string]map[string]int{}}
}

// Tokenize lowercases text and splits it into words of two letters or more.
func Tokenize(text string) []string {
	words := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
	return slices.DeleteFunc(words, func(w string) bool { return len(w) < 2 })
}

func (ix *Index) add(org, kind, id, title, text string) {
	ix.mu.Lock()
	defer ix.mu.Unlock()
	key := kind + "/" + id
	if old, ok := ix.docs[key]; ok {
		for t := range old.terms {
			delete(ix.post[t], key)
		}
	}
	d := doc{org: org, hit: api.Hit{Kind: kind, ID: id, Title: title}, terms: map[string]int{}}
	for _, t := range Tokenize(title + " " + text) {
		d.terms[t]++
	}
	for t, n := range d.terms {
		if ix.post[t] == nil {
			ix.post[t] = map[string]int{}
		}
		ix.post[t][key] = n
	}
	ix.docs[key] = d
}

// AddTodo indexes a todo.
func (ix *Index) AddTodo(t todo.Todo) { ix.add(t.Org, "todo", t.ID, t.Title, "") }

// AddIssue indexes an issue.
func (ix *Index) AddIssue(i issues.Issue) { ix.add(i.Org, "issue", i.ID, i.Title, i.Body) }

// AddMessage indexes a tracked email.
func (ix *Index) AddMessage(m mail.Message) { ix.add(m.Org, "mail", m.ID, m.Subject, m.Body) }

// Search returns org's documents containing every query term, best first.
func (ix *Index) Search(org, query string) []api.Hit {
	terms := Tokenize(query)
	if len(terms) == 0 {
		return nil
	}
	ix.mu.RLock()
	defer ix.mu.RUnlock()
	var hits []api.Hit
	for key, n := range ix.post[terms[0]] {
		d := ix.docs[key]
		if d.org != org {
			continue
		}
		score := n
		for _, t := range terms[1:] {
			c := ix.post[t][key]
			if c == 0 {
				score = 0
				break
			}
			score += c
		}
		if score > 0 {
			h := d.hit
			h.Score = score
			hits = append(hits, h)
		}
	}
	slices.SortFunc(hits, func(a, b api.Hit) int {
		return cmp.Or(cmp.Compare(b.Score, a.Score), cmp.Compare(a.Kind+a.ID, b.Kind+b.ID))
	})
	return hits
}
