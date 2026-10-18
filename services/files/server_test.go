package files

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pathroute/suite/libs/go/authz"
	"github.com/pathroute/suite/libs/go/platform/platformtest"
	"github.com/pathroute/suite/services/files/api"
)

var member = authz.TestToken("u1", "acme", authz.RoleMember)

func put(h http.Handler, name, body, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPut, "/files/"+name, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestPutGetList(t *testing.T) {
	h := NewServer(authz.TestSigner).Handler()
	rec := put(h, "a.txt", "hello", member)
	platformtest.Expect(t, rec, http.StatusOK)
	meta := platformtest.Decode[api.File](t, rec)
	if meta.Size != 5 || meta.SHA256 != "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" {
		t.Fatalf("meta %+v", meta)
	}
	put(h, "a.txt", "hello again", member) // overwrite
	rec = platformtest.Do(t, h, http.MethodGet, "/files/a.txt", member, nil)
	platformtest.Expect(t, rec, http.StatusOK)
	if rec.Body.String() != "hello again" {
		t.Fatalf("body %q", rec.Body)
	}
	if list := platformtest.Decode[[]api.File](t, platformtest.Do(t, h, http.MethodGet, "/files", member, nil)); len(list) != 1 {
		t.Fatalf("list %+v", list)
	}
	platformtest.Expect(t, platformtest.Do(t, h, http.MethodGet, "/files/a.txt", authz.TestToken("u2", "globex", authz.RoleViewer), nil), http.StatusNotFound)
}

func TestPutLimits(t *testing.T) {
	h := NewServer(authz.TestSigner).Handler()
	platformtest.Expect(t, put(h, "big.bin", strings.Repeat("x", MaxSize+1), member), http.StatusRequestEntityTooLarge)
}

func TestValidName(t *testing.T) {
	for n, want := range map[string]bool{"a.txt": true, "": false, "..": false, `a\b`: false, strings.Repeat("n", 201): false} {
		if ValidName(n) != want {
			t.Errorf("ValidName(%q) != %v", n, want)
		}
	}
}
