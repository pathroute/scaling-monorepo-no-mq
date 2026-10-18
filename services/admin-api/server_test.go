package adminapi

import (
	"net/http"
	"testing"

	"github.com/pathroute/suite/libs/go/authz"
	"github.com/pathroute/suite/libs/go/platform/platformtest"
	"github.com/pathroute/suite/services/admin-api/api"
)

func TestFlagsAndAudit(t *testing.T) {
	h := NewServer(authz.TestSigner).Handler()
	admin := authz.TestToken("root", "acme", authz.RoleAdmin)
	platformtest.Expect(t, platformtest.Do(t, h, http.MethodPut, "/admin/flags/new_ui", admin, api.SetFlag{Enabled: true}), http.StatusOK)
	platformtest.Expect(t, platformtest.Do(t, h, http.MethodPut, "/admin/flags/new_ui", admin, api.SetFlag{Enabled: false}), http.StatusOK)
	platformtest.Expect(t, platformtest.Do(t, h, http.MethodPut, "/admin/flags/Bad-Name", admin, api.SetFlag{}), http.StatusBadRequest)

	flags := platformtest.Decode[[]api.Flag](t, platformtest.Do(t, h, http.MethodGet, "/admin/flags", admin, nil))
	if len(flags) != 1 || flags[0].Enabled {
		t.Fatalf("flags %+v", flags)
	}
	audit := platformtest.Decode[[]api.AuditEntry](t, platformtest.Do(t, h, http.MethodGet, "/admin/audit", admin, nil))
	if len(audit) != 2 || audit[1].Action != "set new_ui=false" || audit[1].Actor != "root" {
		t.Fatalf("audit %+v", audit)
	}
}

func TestAdminOnly(t *testing.T) {
	h := NewServer(authz.TestSigner).Handler()
	platformtest.Expect(t, platformtest.Do(t, h, http.MethodGet, "/admin/flags", authz.TestToken("u", "acme", authz.RoleMember), nil), http.StatusForbidden)
}
