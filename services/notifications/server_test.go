package notifications

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/pathroute/suite/libs/go/platform/platformtest"
	"github.com/pathroute/suite/services/notifications/api"
)

type failing struct{}

func (failing) Send(context.Context, string, string) error { return errors.New("smtp down") }

func TestFanOut(t *testing.T) {
	push := &LogSender{}
	h := NewServer(map[string]Sender{api.ChannelEmail: failing{}, api.ChannelPush: push}).Handler()
	rec := platformtest.Do(t, h, http.MethodPost, "/notify", "", api.Notify{User: "u1", Channels: []string{api.ChannelEmail, api.ChannelPush}, Message: "hi"})
	platformtest.Expect(t, rec, http.StatusOK)
	ds := platformtest.Decode[[]api.Delivery](t, rec)
	if len(ds) != 2 || ds[0].OK || ds[0].Error != "smtp down" || !ds[1].OK {
		t.Fatalf("deliveries %+v", ds)
	}
	if len(push.Sent) != 1 || push.Sent[0] != "u1: hi" {
		t.Fatalf("push %v", push.Sent)
	}
	if got := platformtest.Decode[[]api.Delivery](t, platformtest.Do(t, h, http.MethodGet, "/deliveries?user=u1", "", nil)); len(got) != 2 {
		t.Fatalf("log %+v", got)
	}
}

func TestNotifyValidation(t *testing.T) {
	h := NewServer(map[string]Sender{api.ChannelPush: &LogSender{}}).Handler()
	platformtest.Expect(t, platformtest.Do(t, h, http.MethodPost, "/notify", "", api.Notify{User: "u1"}), http.StatusBadRequest)
	platformtest.Expect(t, platformtest.Do(t, h, http.MethodPost, "/notify", "", api.Notify{User: "u1", Channels: []string{"fax"}}), http.StatusBadRequest)
}
