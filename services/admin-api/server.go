// Package adminapi serves internal admin endpoints: feature flags and their audit log.
package adminapi

import (
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/pathroute/suite/libs/go/authz"
	"github.com/pathroute/suite/libs/go/db"
	"github.com/pathroute/suite/libs/go/platform"
	"github.com/pathroute/suite/services/admin-api/api"
)

var flagName = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

// Server holds flags and the audit log.
type Server struct {
	signer authz.Signer
	flags  *db.Table[api.Flag]
	mu     sync.Mutex
	audit  []api.AuditEntry
	now    func() time.Time
}

// NewServer returns a server with no flags.
func NewServer(signer authz.Signer) *Server {
	return &Server{signer: signer, flags: db.NewTable[api.Flag](), now: time.Now}
}

// Handler routes the service's endpoints; all of them need an admin token.
func (s *Server) Handler() http.Handler {
	mux := platform.NewMux()
	mux.Handle("GET /admin/flags", s.signer.Require(authz.RoleAdmin, http.HandlerFunc(s.listFlags)))
	mux.Handle("PUT /admin/flags/{name}", s.signer.Require(authz.RoleAdmin, http.HandlerFunc(s.setFlag)))
	mux.Handle("GET /admin/audit", s.signer.Require(authz.RoleAdmin, http.HandlerFunc(s.listAudit)))
	return mux
}

func (s *Server) listFlags(w http.ResponseWriter, r *http.Request) {
	platform.WriteJSON(w, http.StatusOK, s.flags.List(nil))
}

func (s *Server) setFlag(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if !flagName.MatchString(name) {
		platform.WriteError(w, http.StatusBadRequest, "invalid flag name")
		return
	}
	var req api.SetFlag
	if err := platform.ReadJSON(r, &req); err != nil {
		platform.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	f := api.Flag{Name: name, Enabled: req.Enabled}
	s.flags.Tx(func(tx *db.Tx[api.Flag]) error {
		if tx.Update(name, f) != nil {
			return tx.Insert(name, f)
		}
		return nil
	})
	c, _ := authz.FromContext(r.Context())
	s.record(c.Subject, fmt.Sprintf("set %s=%t", name, req.Enabled))
	platform.WriteJSON(w, http.StatusOK, f)
}

func (s *Server) record(actor, action string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.audit = append(s.audit, api.AuditEntry{Seq: len(s.audit) + 1, Actor: actor, Action: action, At: s.now()})
}

func (s *Server) listAudit(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	platform.WriteJSON(w, http.StatusOK, s.audit)
}
