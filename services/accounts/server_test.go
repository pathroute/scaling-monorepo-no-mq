package accounts

import (
	"net/http"
	"testing"

	"github.com/pathroute/suite/libs/go/authz"
	"github.com/pathroute/suite/libs/go/platform/platformtest"
	"github.com/pathroute/suite/services/accounts/api"
)

func TestSignupMeAndMembers(t *testing.T) {
	h := NewServer(authz.TestSigner).Handler()
	rec := platformtest.Do(t, h, http.MethodPost, "/signup", "", api.Signup{Email: "Ann@Acme.io", Org: "acme"})
	platformtest.Expect(t, rec, http.StatusCreated)
	admin := platformtest.Decode[api.Session](t, rec)

	me := platformtest.Decode[api.User](t, platformtest.Do(t, h, http.MethodGet, "/me", admin.Token, nil))
	if me.Email != "ann@acme.io" || me.Role != authz.RoleAdmin {
		t.Fatalf("me %+v", me)
	}

	rec = platformtest.Do(t, h, http.MethodPost, "/members", admin.Token, api.AddMember{Email: "bob@acme.io", Role: authz.RoleMember})
	platformtest.Expect(t, rec, http.StatusCreated)
	bob := platformtest.Decode[api.Session](t, rec)
	if bob.User.Org != "acme" {
		t.Fatalf("bob %+v", bob.User)
	}
	platformtest.Expect(t, platformtest.Do(t, h, http.MethodPost, "/members", bob.Token, api.AddMember{Email: "c@acme.io", Role: authz.RoleMember}), http.StatusForbidden)
}

func TestSignupConflictsAndValidation(t *testing.T) {
	h := NewServer(authz.TestSigner).Handler()
	platformtest.Expect(t, platformtest.Do(t, h, http.MethodPost, "/signup", "", api.Signup{Email: "a@x.io", Org: "acme"}), http.StatusCreated)
	platformtest.Expect(t, platformtest.Do(t, h, http.MethodPost, "/signup", "", api.Signup{Email: "b@x.io", Org: "acme"}), http.StatusConflict)
	platformtest.Expect(t, platformtest.Do(t, h, http.MethodPost, "/signup", "", api.Signup{Email: "a@x.io", Org: "other"}), http.StatusConflict)
	platformtest.Expect(t, platformtest.Do(t, h, http.MethodPost, "/signup", "", api.Signup{Email: "a@x.io", Org: "Bad Org"}), http.StatusBadRequest)
	// The failed email signup must not leave its org behind.
	platformtest.Expect(t, platformtest.Do(t, h, http.MethodPost, "/signup", "", api.Signup{Email: "c@x.io", Org: "other"}), http.StatusCreated)
}
