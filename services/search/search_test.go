package search

import (
	"net/http"
	"slices"
	"testing"

	"github.com/pathroute/suite/libs/go/authz"
	"github.com/pathroute/suite/libs/go/platform/platformtest"
	issues "github.com/pathroute/suite/services/issues/api"
	mail "github.com/pathroute/suite/services/mail-tracker/api"
	"github.com/pathroute/suite/services/search/api"
	todo "github.com/pathroute/suite/services/todo/api"
)

func TestTokenize(t *testing.T) {
	if got := Tokenize("Fix the Login-page, a 404!"); !slices.Equal(got, []string{"fix", "the", "login", "page", "404"}) {
		t.Fatalf("got %v", got)
	}
}

func TestSearchRanksAndScopes(t *testing.T) {
	ix := NewIndex()
	ix.AddTodo(todo.Todo{ID: "t1", Org: "acme", Title: "fix login"})
	ix.AddIssue(issues.Issue{ID: "i1", Org: "acme", Title: "login broken", Body: "login fails after fix"})
	ix.AddMessage(mail.Message{ID: "m1", Org: "globex", Subject: "login fix"})

	hits := ix.Search("acme", "login fix")
	if len(hits) != 2 || hits[0].ID != "i1" || hits[0].Score != 3 {
		t.Fatalf("hits %+v", hits)
	}
	// Re-indexing replaces the old terms.
	ix.AddTodo(todo.Todo{ID: "t1", Org: "acme", Title: "buy milk"})
	if hits := ix.Search("acme", "fix"); len(hits) != 1 || hits[0].Kind != "issue" {
		t.Fatalf("after reindex %+v", hits)
	}
	if hits := ix.Search("acme", "  "); hits != nil {
		t.Fatalf("empty query %+v", hits)
	}
}

func TestIngestChecksOrg(t *testing.T) {
	h := NewServer(authz.TestSigner).Handler()
	member := authz.TestToken("u1", "acme", authz.RoleMember)
	platformtest.Expect(t, platformtest.Do(t, h, http.MethodPost, "/index/todos", member, []todo.Todo{{ID: "t1", Org: "acme", Title: "quarterly report"}}), http.StatusOK)
	platformtest.Expect(t, platformtest.Do(t, h, http.MethodPost, "/index/mail", member, []mail.Message{{ID: "m1", Org: "globex"}}), http.StatusForbidden)

	hits := platformtest.Decode[[]api.Hit](t, platformtest.Do(t, h, http.MethodGet, "/search?q=report", member, nil))
	if len(hits) != 1 || hits[0].Kind != "todo" {
		t.Fatalf("hits %+v", hits)
	}
}
