// Package aigateway routes completion requests to LLM providers and enforces per-org quotas.
package aigateway

import (
	"errors"
	"net/http"
	"sync"

	"github.com/pathroute/suite/libs/go/aiclient"
	"github.com/pathroute/suite/libs/go/authz"
	"github.com/pathroute/suite/libs/go/platform"
	"github.com/pathroute/suite/services/ai-gateway/api"
)

// Route maps a model name to the provider that serves it.
type Route struct {
	Provider string
	Client   aiclient.Provider
}

// Quota tracks tokens left per org in the current window.
type Quota struct {
	mu    sync.Mutex
	limit int
	used  map[string]int
}

// NewQuota returns a quota of limit tokens per org.
func NewQuota(limit int) *Quota { return &Quota{limit: limit, used: map[string]int{}} }

// Remaining returns the tokens org has left.
func (q *Quota) Remaining(org string) int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.limit - q.used[org]
}

// Charge records n tokens against org and returns what's left.
func (q *Quota) Charge(org string, n int) int {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.used[org] += n
	return q.limit - q.used[org]
}

// Server holds the gateway's routes and quotas.
type Server struct {
	signer authz.Signer
	routes map[string]Route
	quota  *Quota
}

// NewServer returns a gateway serving routes under quota.
func NewServer(signer authz.Signer, routes map[string]Route, quota *Quota) *Server {
	return &Server{signer: signer, routes: routes, quota: quota}
}

// Handler routes the service's endpoints.
func (s *Server) Handler() http.Handler {
	mux := platform.NewMux()
	mux.Handle("POST /v1/complete", s.signer.Require(authz.RoleMember, http.HandlerFunc(s.complete)))
	return mux
}

func (s *Server) complete(w http.ResponseWriter, r *http.Request) {
	c, _ := authz.FromContext(r.Context())
	var req api.Complete
	if err := platform.ReadJSON(r, &req); err != nil {
		platform.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	route, ok := s.routes[req.Model]
	if !ok {
		platform.WriteError(w, http.StatusBadRequest, "unknown model "+req.Model)
		return
	}
	if s.quota.Remaining(c.Org) <= 0 {
		platform.WriteError(w, http.StatusTooManyRequests, "quota exhausted")
		return
	}
	resp, err := route.Client.Complete(r.Context(), req)
	if err != nil {
		status := http.StatusBadGateway
		if errors.Is(err, aiclient.ErrRetryable) {
			status = http.StatusServiceUnavailable
		}
		platform.WriteError(w, status, err.Error())
		return
	}
	left := s.quota.Charge(c.Org, resp.InputTokens+resp.OutputTokens)
	platform.WriteJSON(w, http.StatusOK, api.Completion{Response: resp, Provider: route.Provider, QuotaRemaining: left})
}
