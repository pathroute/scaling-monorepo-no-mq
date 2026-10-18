package todo

import (
	"net/http"
	"testing"

	"github.com/pathroute/suite/libs/go/aiclient"
	"github.com/pathroute/suite/libs/go/authz"
	"github.com/pathroute/suite/libs/go/platform/platformtest"
	"github.com/pathroute/suite/services/todo/api"
)

var (
	member = authz.TestToken("u1", "acme", authz.RoleMember)
	viewer = authz.TestToken("u2", "acme", authz.RoleViewer)
	other  = authz.TestToken("u3", "globex", authz.RoleMember)
)

func TestCreateListDone(t *testing.T) {
	h := NewServer(authz.TestSigner, aiclient.Fake{}).Handler()
	rec := platformtest.Do(t, h, http.MethodPost, "/todos", member, api.CreateTodo{Title: " ship it "})
	platformtest.Expect(t, rec, http.StatusCreated)
	created := platformtest.Decode[api.Todo](t, rec)
	if created.Title != "ship it" || created.Org != "acme" {
		t.Fatalf("created %+v", created)
	}

	rec = platformtest.Do(t, h, http.MethodPost, "/todos/"+created.ID+"/done", member, nil)
	platformtest.Expect(t, rec, http.StatusOK)

	list := platformtest.Decode[[]api.Todo](t, platformtest.Do(t, h, http.MethodGet, "/todos", viewer, nil))
	if len(list) != 1 || !list[0].Done {
		t.Fatalf("list %+v", list)
	}
}

func TestCreateValidatesAndAuthorizes(t *testing.T) {
	h := NewServer(authz.TestSigner, aiclient.Fake{}).Handler()
	platformtest.Expect(t, platformtest.Do(t, h, http.MethodPost, "/todos", member, api.CreateTodo{}), http.StatusBadRequest)
	platformtest.Expect(t, platformtest.Do(t, h, http.MethodPost, "/todos", viewer, api.CreateTodo{Title: "x"}), http.StatusForbidden)
	platformtest.Expect(t, platformtest.Do(t, h, http.MethodGet, "/todos", "", nil), http.StatusUnauthorized)
}

func TestOrgsAreIsolated(t *testing.T) {
	h := NewServer(authz.TestSigner, aiclient.Fake{}).Handler()
	created := platformtest.Decode[api.Todo](t, platformtest.Do(t, h, http.MethodPost, "/todos", member, api.CreateTodo{Title: "secret"}))
	platformtest.Expect(t, platformtest.Do(t, h, http.MethodPost, "/todos/"+created.ID+"/done", other, nil), http.StatusNotFound)
	if list := platformtest.Decode[[]api.Todo](t, platformtest.Do(t, h, http.MethodGet, "/todos", other, nil)); len(list) != 0 {
		t.Fatalf("other org sees %+v", list)
	}
}

func TestSuggest(t *testing.T) {
	h := NewServer(authz.TestSigner, aiclient.Fake{}).Handler()
	created := platformtest.Decode[api.Todo](t, platformtest.Do(t, h, http.MethodPost, "/todos", member, api.CreateTodo{Title: "plan offsite"}))
	rec := platformtest.Do(t, h, http.MethodPost, "/todos/"+created.ID+"/suggest", member, nil)
	platformtest.Expect(t, rec, http.StatusOK)
	if s := platformtest.Decode[api.Suggestion](t, rec); len(s.Subtasks) != 1 {
		t.Fatalf("suggestion %+v", s)
	}
}
