// Package mailtracker tracks sent emails (open pixels) and classifies them.
package mailtracker

import (
	"net/http"
	"net/mail"
	"slices"
	"strings"
	"time"

	"github.com/pathroute/suite/libs/go/aiclient"
	"github.com/pathroute/suite/libs/go/authz"
	"github.com/pathroute/suite/libs/go/db"
	"github.com/pathroute/suite/libs/go/platform"
	"github.com/pathroute/suite/services/mail-tracker/api"
)

// pixel is a 1x1 transparent GIF.
var pixel = []byte("GIF89a\x01\x00\x01\x00\x80\x00\x00\x00\x00\x00\xff\xff\xff!\xf9\x04\x01\x00\x00\x00\x00,\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02D\x01\x00;")

// Server holds the mail-tracker service's state.
type Server struct {
	signer   authz.Signer
	ai       aiclient.Provider
	messages *db.Table[api.Message]
	now      func() time.Time
}

// NewServer returns a server with an empty store.
func NewServer(signer authz.Signer, ai aiclient.Provider) *Server {
	return &Server{signer: signer, ai: ai, messages: db.NewTable[api.Message](), now: time.Now}
}

// Handler routes the service's endpoints. The open pixel is public by design.
func (s *Server) Handler() http.Handler {
	mux := platform.NewMux()
	mux.Handle("GET /messages", s.signer.Require(authz.RoleViewer, http.HandlerFunc(s.list)))
	mux.Handle("POST /messages", s.signer.Require(authz.RoleMember, http.HandlerFunc(s.track)))
	mux.Handle("POST /messages/{id}/classify", s.signer.Require(authz.RoleMember, http.HandlerFunc(s.classify)))
	mux.HandleFunc("GET /t/{id}", s.open)
	return mux
}

func org(r *http.Request) string {
	c, _ := authz.FromContext(r.Context())
	return c.Org
}

func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	o := org(r)
	platform.WriteJSON(w, http.StatusOK, s.messages.List(func(m api.Message) bool { return m.Org == o }))
}

func (s *Server) track(w http.ResponseWriter, r *http.Request) {
	var req api.Track
	if err := platform.ReadJSON(r, &req); err != nil {
		platform.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if _, err := mail.ParseAddress(req.To); err != nil {
		platform.WriteError(w, http.StatusBadRequest, "invalid recipient")
		return
	}
	m := api.Message{ID: db.NewID(), Org: org(r), To: req.To, Subject: req.Subject, Body: req.Body, SentAt: s.now()}
	s.messages.Insert(m.ID, m)
	platform.WriteJSON(w, http.StatusCreated, m)
}

// open records a pixel hit. Unknown ids still get the pixel, so the tracker can't be probed.
func (s *Server) open(w http.ResponseWriter, r *http.Request) {
	s.messages.Tx(func(tx *db.Tx[api.Message]) error {
		m, err := tx.Get(r.PathValue("id"))
		if err != nil {
			return err
		}
		m.Opens++
		return tx.Update(m.ID, m)
	})
	w.Header().Set("Content-Type", "image/gif")
	w.Write(pixel)
}

// Category maps a model's answer onto a known category.
func Category(answer string) string {
	a := strings.ToLower(answer)
	for _, c := range api.Categories {
		if strings.Contains(a, c) {
			return c
		}
	}
	return "other"
}

func (s *Server) classify(w http.ResponseWriter, r *http.Request) {
	m, err := s.messages.Get(r.PathValue("id"))
	if err != nil || m.Org != org(r) {
		platform.WriteError(w, http.StatusNotFound, "no such message")
		return
	}
	resp, err := s.ai.Complete(r.Context(), aiclient.Request{
		Model:     "small",
		Prompt:    "Classify as one of " + strings.Join(slices.Clone(api.Categories), ", ") + ": " + m.Subject,
		MaxTokens: 5,
	})
	if err != nil {
		platform.WriteError(w, http.StatusBadGateway, err.Error())
		return
	}
	m.Category = Category(resp.Text)
	s.messages.Tx(func(tx *db.Tx[api.Message]) error { return tx.Update(m.ID, m) })
	platform.WriteJSON(w, http.StatusOK, m)
}
