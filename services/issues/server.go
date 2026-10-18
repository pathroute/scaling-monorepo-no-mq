// Package issues serves the issue tracker.
package issues

import (
	"net/http"
	"strings"
	"time"

	"github.com/pathroute/suite/libs/go/aiclient"
	"github.com/pathroute/suite/libs/go/authz"
	"github.com/pathroute/suite/libs/go/db"
	"github.com/pathroute/suite/libs/go/platform"
	"github.com/pathroute/suite/services/issues/api"
)

// allowed lists the legal status transitions.
var allowed = map[api.Status][]api.Status{
	api.StatusOpen:       {api.StatusInProgress, api.StatusClosed},
	api.StatusInProgress: {api.StatusOpen, api.StatusClosed},
	api.StatusClosed:     {api.StatusOpen},
}

// CanTransition reports whether an issue may move from one status to another.
func CanTransition(from, to api.Status) bool {
	for _, s := range allowed[from] {
		if s == to {
			return true
		}
	}
	return false
}

// Server holds the issues service's state.
type Server struct {
	signer authz.Signer
	ai     aiclient.Provider
	issues *db.Table[api.Issue]
	now    func() time.Time
}

// NewServer returns a server with an empty store.
func NewServer(signer authz.Signer, ai aiclient.Provider) *Server {
	return &Server{signer: signer, ai: ai, issues: db.NewTable[api.Issue](), now: time.Now}
}

// Handler routes the service's endpoints.
func (s *Server) Handler() http.Handler {
	mux := platform.NewMux()
	mux.Handle("GET /issues", s.signer.Require(authz.RoleViewer, http.HandlerFunc(s.list)))
	mux.Handle("POST /issues", s.signer.Require(authz.RoleMember, http.HandlerFunc(s.create)))
	mux.Handle("POST /issues/{id}/status", s.signer.Require(authz.RoleMember, http.HandlerFunc(s.transition)))
	mux.Handle("POST /issues/{id}/summarize", s.signer.Require(authz.RoleViewer, http.HandlerFunc(s.summarize)))
	return mux
}

func org(r *http.Request) string {
	c, _ := authz.FromContext(r.Context())
	return c.Org
}

func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	o, status := org(r), api.Status(r.URL.Query().Get("status"))
	platform.WriteJSON(w, http.StatusOK, s.issues.List(func(i api.Issue) bool {
		return i.Org == o && (status == "" || i.Status == status)
	}))
}

func (s *Server) create(w http.ResponseWriter, r *http.Request) {
	var req api.CreateIssue
	if err := platform.ReadJSON(r, &req); err != nil || strings.TrimSpace(req.Title) == "" {
		platform.WriteError(w, http.StatusBadRequest, "title is required")
		return
	}
	i := api.Issue{ID: db.NewID(), Org: org(r), Title: strings.TrimSpace(req.Title), Body: req.Body, Status: api.StatusOpen, CreatedAt: s.now()}
	s.issues.Insert(i.ID, i)
	platform.WriteJSON(w, http.StatusCreated, i)
}

func (s *Server) get(w http.ResponseWriter, r *http.Request) (api.Issue, bool) {
	i, err := s.issues.Get(r.PathValue("id"))
	if err != nil || i.Org != org(r) {
		platform.WriteError(w, http.StatusNotFound, "no such issue")
		return api.Issue{}, false
	}
	return i, true
}

func (s *Server) transition(w http.ResponseWriter, r *http.Request) {
	var req api.Transition
	if err := platform.ReadJSON(r, &req); err != nil {
		platform.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	i, ok := s.get(w, r)
	if !ok {
		return
	}
	if !CanTransition(i.Status, req.Status) {
		platform.WriteError(w, http.StatusConflict, "cannot move from "+string(i.Status)+" to "+string(req.Status))
		return
	}
	i.Status = req.Status
	s.issues.Tx(func(tx *db.Tx[api.Issue]) error { return tx.Update(i.ID, i) })
	platform.WriteJSON(w, http.StatusOK, i)
}

func (s *Server) summarize(w http.ResponseWriter, r *http.Request) {
	i, ok := s.get(w, r)
	if !ok {
		return
	}
	resp, err := s.ai.Complete(r.Context(), aiclient.Request{Model: "small", Prompt: "Summarize: " + i.Title + "\n" + i.Body, MaxTokens: 50})
	if err != nil {
		platform.WriteError(w, http.StatusBadGateway, err.Error())
		return
	}
	platform.WriteJSON(w, http.StatusOK, api.Summary{Text: resp.Text})
}
