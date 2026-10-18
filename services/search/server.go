package search

import (
	"net/http"

	"github.com/pathroute/suite/libs/go/authz"
	"github.com/pathroute/suite/libs/go/platform"
	issues "github.com/pathroute/suite/services/issues/api"
	mail "github.com/pathroute/suite/services/mail-tracker/api"
	todo "github.com/pathroute/suite/services/todo/api"
)

// Server serves queries and accepts documents from the services it indexes.
type Server struct {
	signer authz.Signer
	index  *Index
}

// NewServer returns a server with an empty index.
func NewServer(signer authz.Signer) *Server {
	return &Server{signer: signer, index: NewIndex()}
}

// Handler routes the service's endpoints.
func (s *Server) Handler() http.Handler {
	mux := platform.NewMux()
	mux.Handle("GET /search", s.signer.Require(authz.RoleViewer, http.HandlerFunc(s.search)))
	mux.Handle("POST /index/todos", s.signer.Require(authz.RoleMember, ingest(s.index.AddTodo)))
	mux.Handle("POST /index/issues", s.signer.Require(authz.RoleMember, ingest(s.index.AddIssue)))
	mux.Handle("POST /index/mail", s.signer.Require(authz.RoleMember, ingest(s.index.AddMessage)))
	return mux
}

// orgOf lets ingest check that documents belong to the caller's org.
type orgOf interface {
	todo.Todo | issues.Issue | mail.Message
}

func docOrg[T orgOf](v T) string {
	switch d := any(v).(type) {
	case todo.Todo:
		return d.Org
	case issues.Issue:
		return d.Org
	case mail.Message:
		return d.Org
	}
	return ""
}

func ingest[T orgOf](add func(T)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, _ := authz.FromContext(r.Context())
		var docs []T
		if err := platform.ReadJSON(r, &docs); err != nil {
			platform.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		for _, d := range docs {
			if docOrg(d) != c.Org {
				platform.WriteError(w, http.StatusForbidden, "document from another org")
				return
			}
		}
		for _, d := range docs {
			add(d)
		}
		platform.WriteJSON(w, http.StatusOK, map[string]int{"indexed": len(docs)})
	})
}

func (s *Server) search(w http.ResponseWriter, r *http.Request) {
	c, _ := authz.FromContext(r.Context())
	platform.WriteJSON(w, http.StatusOK, s.index.Search(c.Org, r.URL.Query().Get("q")))
}
