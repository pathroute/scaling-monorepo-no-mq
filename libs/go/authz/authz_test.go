package authz

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var (
	t0     = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	signer = Signer{Key: []byte("test-key"), Now: func() time.Time { return t0 }}
)

func issue(t *testing.T, roles ...string) string {
	t.Helper()
	tok, err := signer.Issue(Claims{Subject: "u1", Org: "o1", Roles: roles, Expires: t0.Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func TestIssueVerifyRoundTrip(t *testing.T) {
	c, err := signer.Verify(issue(t, RoleMember))
	if err != nil || c.Subject != "u1" || !c.Has(RoleMember) || c.Has(RoleAdmin) {
		t.Fatalf("got %+v, %v", c, err)
	}
}

func TestVerifyErrors(t *testing.T) {
	tok := issue(t, RoleMember)
	other := Signer{Key: []byte("other"), Now: signer.Now}
	late := Signer{Key: signer.Key, Now: func() time.Time { return t0.Add(2 * time.Hour) }}
	cases := map[string]struct {
		s    Signer
		tok  string
		want error
	}{
		"malformed": {signer, "nodot", ErrMalformed},
		"signature": {other, tok, ErrSignature},
		"expired":   {late, tok, ErrExpired},
	}
	for name, c := range cases {
		if _, err := c.s.Verify(c.tok); !errors.Is(err, c.want) {
			t.Errorf("%s: got %v, want %v", name, err, c.want)
		}
	}
}

func TestAdminHasEveryRole(t *testing.T) {
	if !(Claims{Roles: []string{RoleAdmin}}).Has(RoleMember) {
		t.Fatal("admin should have member")
	}
}

func TestRequire(t *testing.T) {
	h := signer.Require(RoleMember, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, _ := FromContext(r.Context())
		w.Write([]byte(c.Subject))
	}))
	for _, c := range []struct {
		auth string
		want int
	}{
		{"", http.StatusUnauthorized},
		{"Bearer " + issue(t, RoleViewer), http.StatusForbidden},
		{"Bearer " + issue(t, RoleMember), http.StatusOK},
	} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		if c.auth != "" {
			req.Header.Set("Authorization", c.auth)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != c.want {
			t.Errorf("auth %q: got %d, want %d", c.auth, rec.Code, c.want)
		}
	}
}

func TestRolesAreOrdered(t *testing.T) {
	member := Claims{Roles: []string{RoleMember}}
	if !member.Has(RoleViewer) || member.Has(RoleAdmin) || (Claims{Roles: []string{"bogus"}}).Has(RoleViewer) {
		t.Fatal("role order wrong")
	}
}
