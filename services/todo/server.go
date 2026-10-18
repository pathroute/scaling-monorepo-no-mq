// Package todo serves "AI-enabled" todo lists.
package todo

import (
	"net/http"
	"strings"
	"time"

	"github.com/pathroute/suite/libs/go/aiclient"
	"github.com/pathroute/suite/libs/go/authz"
	"github.com/pathroute/suite/libs/go/db"
	"github.com/pathroute/suite/libs/go/platform"
	"github.com/pathroute/suite/services/todo/api"
)

// Server holds the todo service's state.
type Server struct {
	signer authz.Signer
	ai     aiclient.Provider
	todos  *db.Table[api.Todo]
	now    func() time.Time
}

// NewServer returns a server with an empty store.
func NewServer(signer authz.Signer, ai aiclient.Provider) *Server {
	return &Server{signer: signer, ai: ai, todos: db.NewTable[api.Todo](), now: time.Now}
}

// Handler routes the service's endpoints.
func (s *Server) Handler() http.Handler {
	mux := platform.NewMux()
	mux.Handle("GET /todos", s.signer.Require(authz.RoleViewer, http.HandlerFunc(s.list)))
	mux.Handle("POST /todos", s.signer.Require(authz.RoleMember, http.HandlerFunc(s.create)))
	mux.Handle("POST /todos/{id}/done", s.signer.Require(authz.RoleMember, http.HandlerFunc(s.done)))
	mux.Handle("POST /todos/{id}/suggest", s.signer.Require(authz.RoleMember, http.HandlerFunc(s.suggest)))
	return mux
}

func org(r *http.Request) string {
	c, _ := authz.FromContext(r.Context())
	return c.Org
}

func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	o := org(r)
	platform.WriteJSON(w, http.StatusOK, s.todos.List(func(t api.Todo) bool { return t.Org == o }))
}

func (s *Server) create(w http.ResponseWriter, r *http.Request) {
	var req api.CreateTodo
	if err := platform.ReadJSON(r, &req); err != nil || strings.TrimSpace(req.Title) == "" {
		platform.WriteError(w, http.StatusBadRequest, "title is required")
		return
	}
	t := api.Todo{ID: db.NewID(), Org: org(r), Title: strings.TrimSpace(req.Title), CreatedAt: s.now()}
	if err := s.todos.Insert(t.ID, t); err != nil {
		platform.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	platform.WriteJSON(w, http.StatusCreated, t)
}

// get returns the caller's todo, or writes 404 (other orgs' todos are invisible).
func (s *Server) get(w http.ResponseWriter, r *http.Request) (api.Todo, bool) {
	t, err := s.todos.Get(r.PathValue("id"))
	if err != nil || t.Org != org(r) {
		platform.WriteError(w, http.StatusNotFound, "no such todo")
		return api.Todo{}, false
	}
	return t, true
}

func (s *Server) done(w http.ResponseWriter, r *http.Request) {
	t, ok := s.get(w, r)
	if !ok {
		return
	}
	t.Done = true
	s.todos.Tx(func(tx *db.Tx[api.Todo]) error { return tx.Update(t.ID, t) })
	platform.WriteJSON(w, http.StatusOK, t)
}

func (s *Server) suggest(w http.ResponseWriter, r *http.Request) {
	t, ok := s.get(w, r)
	if !ok {
		return
	}
	resp, err := s.ai.Complete(r.Context(), aiclient.Request{
		Model:     "small",
		Prompt:    "Split into subtasks, one per line: " + t.Title,
		MaxTokens: 64,
	})
	if err != nil {
		platform.WriteError(w, http.StatusBadGateway, err.Error())
		return
	}
	var out api.Suggestion
	for _, line := range strings.Split(resp.Text, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			out.Subtasks = append(out.Subtasks, line)
		}
	}
	platform.WriteJSON(w, http.StatusOK, out)
}
