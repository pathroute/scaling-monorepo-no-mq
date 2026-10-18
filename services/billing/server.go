// Package billing records usage and prices invoices.
package billing

import (
	"net/http"

	"github.com/pathroute/suite/libs/go/authz"
	"github.com/pathroute/suite/libs/go/db"
	"github.com/pathroute/suite/libs/go/platform"
	"github.com/pathroute/suite/services/billing/api"
)

// Server holds the billing service's state: tokens used per org this period.
type Server struct {
	signer authz.Signer
	usage  *db.Table[int64]
}

// NewServer returns a server with no usage.
func NewServer(signer authz.Signer) *Server {
	return &Server{signer: signer, usage: db.NewTable[int64]()}
}

// Handler routes the service's endpoints.
func (s *Server) Handler() http.Handler {
	mux := platform.NewMux()
	mux.Handle("POST /usage", s.signer.Require(authz.RoleMember, http.HandlerFunc(s.record)))
	mux.Handle("GET /invoice", s.signer.Require(authz.RoleAdmin, http.HandlerFunc(s.invoice)))
	return mux
}

func (s *Server) record(w http.ResponseWriter, r *http.Request) {
	c, _ := authz.FromContext(r.Context())
	var req api.Usage
	if err := platform.ReadJSON(r, &req); err != nil || req.Tokens <= 0 {
		platform.WriteError(w, http.StatusBadRequest, "tokens must be positive")
		return
	}
	s.usage.Tx(func(tx *db.Tx[int64]) error {
		cur, err := tx.Get(c.Org)
		if err != nil {
			return tx.Insert(c.Org, req.Tokens)
		}
		return tx.Update(c.Org, cur+req.Tokens)
	})
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) invoice(w http.ResponseWriter, r *http.Request) {
	c, _ := authz.FromContext(r.Context())
	tokens, _ := s.usage.Get(c.Org)
	platform.WriteJSON(w, http.StatusOK, Price(c.Org, tokens))
}
