// Package notifications fans a notification out to email, push and webhooks.
// It's internal: other services call it, so it has no end-user auth.
package notifications

import (
	"context"
	"net/http"
	"time"

	"github.com/pathroute/suite/libs/go/db"
	"github.com/pathroute/suite/libs/go/platform"
	"github.com/pathroute/suite/services/notifications/api"
)

// Sender delivers a message to a user on one channel.
type Sender interface {
	Send(ctx context.Context, user, message string) error
}

// Server holds the senders and the delivery log.
type Server struct {
	senders    map[string]Sender
	deliveries *db.Table[api.Delivery]
	now        func() time.Time
}

// NewServer returns a server that delivers through senders, keyed by channel.
func NewServer(senders map[string]Sender) *Server {
	return &Server{senders: senders, deliveries: db.NewTable[api.Delivery](), now: time.Now}
}

// Handler routes the service's endpoints.
func (s *Server) Handler() http.Handler {
	mux := platform.NewMux()
	mux.HandleFunc("POST /notify", s.notify)
	mux.HandleFunc("GET /deliveries", s.list)
	return mux
}

func (s *Server) notify(w http.ResponseWriter, r *http.Request) {
	var req api.Notify
	if err := platform.ReadJSON(r, &req); err != nil || req.User == "" || len(req.Channels) == 0 {
		platform.WriteError(w, http.StatusBadRequest, "user and channels are required")
		return
	}
	for _, ch := range req.Channels {
		if _, ok := s.senders[ch]; !ok {
			platform.WriteError(w, http.StatusBadRequest, "unknown channel "+ch)
			return
		}
	}
	// One failing channel doesn't stop the others; each outcome is logged.
	out := make([]api.Delivery, 0, len(req.Channels))
	for _, ch := range req.Channels {
		d := api.Delivery{ID: db.NewID(), User: req.User, Channel: ch, Message: req.Message, OK: true, SentAt: s.now()}
		if err := s.senders[ch].Send(r.Context(), req.User, req.Message); err != nil {
			d.OK, d.Error = false, err.Error()
		}
		s.deliveries.Insert(d.ID, d)
		out = append(out, d)
	}
	platform.WriteJSON(w, http.StatusOK, out)
}

func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")
	platform.WriteJSON(w, http.StatusOK, s.deliveries.List(func(d api.Delivery) bool { return user == "" || d.User == user }))
}

// LogSender "delivers" by recording the message; it backs every channel in local runs.
type LogSender struct{ Sent []string }

// Send records the message.
func (l *LogSender) Send(_ context.Context, user, message string) error {
	l.Sent = append(l.Sent, user+": "+message)
	return nil
}
