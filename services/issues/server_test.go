package issues

import (
	"net/http"
	"testing"

	"github.com/pathroute/suite/libs/go/aiclient"
	"github.com/pathroute/suite/libs/go/authz"
	"github.com/pathroute/suite/libs/go/platform/platformtest"
	"github.com/pathroute/suite/services/issues/api"
)

var member = authz.TestToken("u1", "acme", authz.RoleMember)

func TestCanTransition(t *testing.T) {
	if !CanTransition(api.StatusOpen, api.StatusClosed) || CanTransition(api.StatusClosed, api.StatusInProgress) {
		t.Fatal("transition table wrong")
	}
}

func TestLifecycle(t *testing.T) {
	h := NewServer(authz.TestSigner, aiclient.Fake{}).Handler()
	rec := platformtest.Do(t, h, http.MethodPost, "/issues", member, api.CreateIssue{Title: "crash on login", Body: "stack trace"})
	platformtest.Expect(t, rec, http.StatusCreated)
	i := platformtest.Decode[api.Issue](t, rec)

	platformtest.Expect(t, platformtest.Do(t, h, http.MethodPost, "/issues/"+i.ID+"/status", member, api.Transition{Status: api.StatusClosed}), http.StatusOK)
	platformtest.Expect(t, platformtest.Do(t, h, http.MethodPost, "/issues/"+i.ID+"/status", member, api.Transition{Status: api.StatusInProgress}), http.StatusConflict)

	closed := platformtest.Decode[[]api.Issue](t, platformtest.Do(t, h, http.MethodGet, "/issues?status=closed", member, nil))
	open := platformtest.Decode[[]api.Issue](t, platformtest.Do(t, h, http.MethodGet, "/issues?status=open", member, nil))
	if len(closed) != 1 || len(open) != 0 {
		t.Fatalf("closed %v open %v", closed, open)
	}
}

func TestSummarize(t *testing.T) {
	h := NewServer(authz.TestSigner, aiclient.Fake{}).Handler()
	i := platformtest.Decode[api.Issue](t, platformtest.Do(t, h, http.MethodPost, "/issues", member, api.CreateIssue{Title: "slow"}))
	rec := platformtest.Do(t, h, http.MethodPost, "/issues/"+i.ID+"/summarize", member, nil)
	platformtest.Expect(t, rec, http.StatusOK)
	if s := platformtest.Decode[api.Summary](t, rec); s.Text == "" {
		t.Fatal("empty summary")
	}
}
