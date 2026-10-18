// Package accounts manages users, orgs and the tokens other services verify.
package accounts

import (
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/pathroute/suite/libs/go/authz"
	"github.com/pathroute/suite/libs/go/db"
	"github.com/pathroute/suite/libs/go/platform"
	"github.com/pathroute/suite/services/accounts/api"
)

// TokenTTL is how long issued tokens stay valid.
const TokenTTL = 12 * time.Hour

var orgName = regexp.MustCompile(`^[a-z][a-z0-9-]{1,30}$`)

// Server holds the accounts service's state. Users are keyed by email.
type Server struct {
	signer authz.Signer
	users  *db.Table[api.User]
	orgs   *db.Table[string]
	now    func() time.Time
}

// NewServer returns a server with an empty store.
func NewServer(signer authz.Signer) *Server {
	return &Server{signer: signer, users: db.NewTable[api.User](), orgs: db.NewTable[string](), now: time.Now}
}

// Handler routes the service's endpoints.
func (s *Server) Handler() http.Handler {
	mux := platform.NewMux()
	mux.HandleFunc("POST /signup", s.signup)
	mux.Handle("GET /me", s.signer.Require(authz.RoleViewer, http.HandlerFunc(s.me)))
	mux.Handle("POST /members", s.signer.Require(authz.RoleAdmin, http.HandlerFunc(s.addMember)))
	return mux
}

func (s *Server) session(u api.User) (api.Session, error) {
	tok, err := s.signer.Issue(authz.Claims{Subject: u.ID, Org: u.Org, Roles: []string{u.Role}, Expires: s.now().Add(TokenTTL)})
	return api.Session{User: u, Token: tok}, err
}

func (s *Server) signup(w http.ResponseWriter, r *http.Request) {
	var req api.Signup
	if err := platform.ReadJSON(r, &req); err != nil {
		platform.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	email := strings.ToLower(req.Email)
	if _, err := mail.ParseAddress(email); err != nil || !orgName.MatchString(req.Org) {
		platform.WriteError(w, http.StatusBadRequest, "need a valid email and org name")
		return
	}
	u := api.User{ID: db.NewID(), Email: email, Org: req.Org, Role: authz.RoleAdmin}
	if err := s.orgs.Insert(req.Org, req.Org); err != nil {
		platform.WriteError(w, http.StatusConflict, "org already exists")
		return
	}
	if err := s.users.Insert(email, u); err != nil {
		s.orgs.Delete(req.Org)
		platform.WriteError(w, http.StatusConflict, "email already registered")
		return
	}
	sess, err := s.session(u)
	if err != nil {
		platform.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	platform.WriteJSON(w, http.StatusCreated, sess)
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	c, _ := authz.FromContext(r.Context())
	users := s.users.List(func(u api.User) bool { return u.ID == c.Subject })
	if len(users) != 1 {
		platform.WriteError(w, http.StatusNotFound, "no such user")
		return
	}
	platform.WriteJSON(w, http.StatusOK, users[0])
}

func (s *Server) addMember(w http.ResponseWriter, r *http.Request) {
	c, _ := authz.FromContext(r.Context())
	var req api.AddMember
	if err := platform.ReadJSON(r, &req); err != nil {
		platform.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Role != authz.RoleViewer && req.Role != authz.RoleMember && req.Role != authz.RoleAdmin {
		platform.WriteError(w, http.StatusBadRequest, "unknown role")
		return
	}
	u := api.User{ID: db.NewID(), Email: strings.ToLower(req.Email), Org: c.Org, Role: req.Role}
	if err := s.users.Insert(u.Email, u); err != nil {
		platform.WriteError(w, http.StatusConflict, "email already registered")
		return
	}
	sess, _ := s.session(u)
	platform.WriteJSON(w, http.StatusCreated, sess)
}
