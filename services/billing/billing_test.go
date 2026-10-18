package billing

import (
	"net/http"
	"testing"

	"github.com/pathroute/suite/libs/go/authz"
	"github.com/pathroute/suite/libs/go/platform/platformtest"
	"github.com/pathroute/suite/services/billing/api"
)

func TestPrice(t *testing.T) {
	cases := []struct {
		tokens, cents int64
		lines         int
	}{
		{0, 0, 0},
		{50_000, 0, 1},
		{1_100_000, 200, 2},    // 1M standard tokens
		{12_100_000, 2_295, 3}, // 9.9M standard (1980) + 2.1M volume (315)
		{100_001, 1, 2},        // one standard token rounds up to a cent
	}
	for _, c := range cases {
		inv := Price("acme", c.tokens)
		if inv.TotalCents != c.cents || len(inv.Lines) != c.lines {
			t.Errorf("Price(%d) = %d cents in %d lines, want %d in %d", c.tokens, inv.TotalCents, len(inv.Lines), c.cents, c.lines)
		}
	}
}

func TestUsageAccumulatesPerOrg(t *testing.T) {
	h := NewServer(authz.TestSigner).Handler()
	member := authz.TestToken("u1", "acme", authz.RoleMember)
	admin := authz.TestToken("u2", "acme", authz.RoleAdmin)
	for range 3 {
		platformtest.Expect(t, platformtest.Do(t, h, http.MethodPost, "/usage", member, api.Usage{Tokens: 400_000}), http.StatusNoContent)
	}
	platformtest.Expect(t, platformtest.Do(t, h, http.MethodPost, "/usage", member, api.Usage{Tokens: -1}), http.StatusBadRequest)
	platformtest.Expect(t, platformtest.Do(t, h, http.MethodGet, "/invoice", member, nil), http.StatusForbidden)

	inv := platformtest.Decode[api.Invoice](t, platformtest.Do(t, h, http.MethodGet, "/invoice", admin, nil))
	if inv.Tokens != 1_200_000 || inv.TotalCents != 220 {
		t.Fatalf("invoice %+v", inv)
	}
	other := platformtest.Decode[api.Invoice](t, platformtest.Do(t, h, http.MethodGet, "/invoice", authz.TestToken("u3", "globex", authz.RoleAdmin), nil))
	if other.Tokens != 0 {
		t.Fatalf("other org billed: %+v", other)
	}
}
