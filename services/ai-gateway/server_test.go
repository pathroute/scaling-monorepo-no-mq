package aigateway

import (
	"context"
	"net/http"
	"testing"

	"github.com/pathroute/suite/libs/go/aiclient"
	"github.com/pathroute/suite/libs/go/authz"
	"github.com/pathroute/suite/libs/go/platform/platformtest"
	"github.com/pathroute/suite/services/ai-gateway/api"
)

type down struct{}

func (down) Complete(context.Context, aiclient.Request) (aiclient.Response, error) {
	return aiclient.Response{}, aiclient.ErrRetryable
}

func newHandler(limit int) http.Handler {
	routes := map[string]Route{
		"small": {Provider: "fake", Client: aiclient.Fake{}},
		"large": {Provider: "down", Client: down{}},
	}
	return NewServer(authz.TestSigner, routes, NewQuota(limit)).Handler()
}

var member = authz.TestToken("u1", "acme", authz.RoleMember)

func TestCompleteChargesQuota(t *testing.T) {
	h := newHandler(10)
	rec := platformtest.Do(t, h, http.MethodPost, "/v1/complete", member, api.Complete{Model: "small", Prompt: "one two three"})
	platformtest.Expect(t, rec, http.StatusOK)
	got := platformtest.Decode[api.Completion](t, rec)
	if got.Provider != "fake" || got.QuotaRemaining != 4 {
		t.Fatalf("got %+v", got)
	}
	platformtest.Do(t, h, http.MethodPost, "/v1/complete", member, api.Complete{Model: "small", Prompt: "one two three"})
	platformtest.Expect(t, platformtest.Do(t, h, http.MethodPost, "/v1/complete", member, api.Complete{Model: "small", Prompt: "x"}), http.StatusTooManyRequests)
	// Quotas are per org.
	platformtest.Expect(t, platformtest.Do(t, h, http.MethodPost, "/v1/complete", authz.TestToken("u2", "globex", authz.RoleMember), api.Complete{Model: "small", Prompt: "x"}), http.StatusOK)
}

func TestCompleteErrors(t *testing.T) {
	h := newHandler(100)
	platformtest.Expect(t, platformtest.Do(t, h, http.MethodPost, "/v1/complete", member, api.Complete{Model: "nope"}), http.StatusBadRequest)
	platformtest.Expect(t, platformtest.Do(t, h, http.MethodPost, "/v1/complete", member, api.Complete{Model: "large"}), http.StatusServiceUnavailable)
}
