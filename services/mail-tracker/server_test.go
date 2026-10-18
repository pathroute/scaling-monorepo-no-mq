package mailtracker

import (
	"net/http"
	"testing"

	"github.com/pathroute/suite/libs/go/aiclient"
	"github.com/pathroute/suite/libs/go/authz"
	"github.com/pathroute/suite/libs/go/platform/platformtest"
	"github.com/pathroute/suite/services/mail-tracker/api"
)

var member = authz.TestToken("u1", "acme", authz.RoleMember)

func TestTrackAndOpen(t *testing.T) {
	h := NewServer(authz.TestSigner, aiclient.Fake{}).Handler()
	platformtest.Expect(t, platformtest.Do(t, h, http.MethodPost, "/messages", member, api.Track{To: "not an address"}), http.StatusBadRequest)
	rec := platformtest.Do(t, h, http.MethodPost, "/messages", member, api.Track{To: "a@example.com", Subject: "hi"})
	platformtest.Expect(t, rec, http.StatusCreated)
	m := platformtest.Decode[api.Message](t, rec)

	for range 2 {
		rec = platformtest.Do(t, h, http.MethodGet, "/t/"+m.ID, "", nil)
		platformtest.Expect(t, rec, http.StatusOK)
	}
	if rec.Header().Get("Content-Type") != "image/gif" {
		t.Fatalf("pixel content type %q", rec.Header().Get("Content-Type"))
	}
	platformtest.Expect(t, platformtest.Do(t, h, http.MethodGet, "/t/unknown", "", nil), http.StatusOK)

	list := platformtest.Decode[[]api.Message](t, platformtest.Do(t, h, http.MethodGet, "/messages", member, nil))
	if len(list) != 1 || list[0].Opens != 2 {
		t.Fatalf("list %+v", list)
	}
}

func TestCategory(t *testing.T) {
	for in, want := range map[string]string{"echo: Billing question": "billing", "SALES": "sales", "dunno": "other"} {
		if got := Category(in); got != want {
			t.Errorf("Category(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestClassify(t *testing.T) {
	h := NewServer(authz.TestSigner, aiclient.Fake{}).Handler()
	m := platformtest.Decode[api.Message](t, platformtest.Do(t, h, http.MethodPost, "/messages", member, api.Track{To: "a@example.com", Subject: "x"}))
	rec := platformtest.Do(t, h, http.MethodPost, "/messages/"+m.ID+"/classify", member, nil)
	platformtest.Expect(t, rec, http.StatusOK)
	if got := platformtest.Decode[api.Message](t, rec); got.Category == "" {
		t.Fatal("no category")
	}
}
