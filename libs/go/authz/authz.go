// Package authz issues and verifies signed tokens and checks roles.
package authz

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/pathroute/suite/libs/go/platform"
)

// Roles, from least to most privileged.
const (
	RoleViewer = "viewer"
	RoleMember = "member"
	RoleAdmin  = "admin"
)

// Claims is the payload of a token.
type Claims struct {
	Subject string    `json:"sub"`
	Org     string    `json:"org"`
	Roles   []string  `json:"roles"`
	Expires time.Time `json:"exp"`
}

var rank = map[string]int{RoleViewer: 1, RoleMember: 2, RoleAdmin: 3}

// Has reports whether the claims carry role or a more privileged one.
func (c Claims) Has(role string) bool {
	return slices.ContainsFunc(c.Roles, func(r string) bool { return rank[r] >= rank[role] && rank[r] > 0 })
}

var (
	ErrMalformed = errors.New("authz: malformed token")
	ErrSignature = errors.New("authz: bad signature")
	ErrExpired   = errors.New("authz: token expired")
)

// Signer signs and verifies tokens with a shared HMAC key.
type Signer struct {
	Key []byte
	Now func() time.Time
}

func (s Signer) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

func (s Signer) mac(payload string) string {
	m := hmac.New(sha256.New, s.Key)
	m.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}

// Issue returns "<payload>.<signature>".
func (s Signer) Issue(c Claims) (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	p := base64.RawURLEncoding.EncodeToString(b)
	return p + "." + s.mac(p), nil
}

// Verify checks the signature and expiry and returns the claims.
func (s Signer) Verify(token string) (Claims, error) {
	p, sig, ok := strings.Cut(token, ".")
	if !ok {
		return Claims{}, ErrMalformed
	}
	if !hmac.Equal([]byte(sig), []byte(s.mac(p))) {
		return Claims{}, ErrSignature
	}
	b, err := base64.RawURLEncoding.DecodeString(p)
	if err != nil {
		return Claims{}, ErrMalformed
	}
	var c Claims
	if err := json.Unmarshal(b, &c); err != nil {
		return Claims{}, ErrMalformed
	}
	if !s.now().Before(c.Expires) {
		return Claims{}, ErrExpired
	}
	return c, nil
}

type ctxKey struct{}

// FromContext returns the claims set by Require.
func FromContext(ctx context.Context) (Claims, bool) {
	c, ok := ctx.Value(ctxKey{}).(Claims)
	return c, ok
}

// Require wraps next so it only runs for a valid bearer token carrying role.
func (s Signer) Require(role string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tok, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok {
			platform.WriteError(w, http.StatusUnauthorized, "missing bearer token")
			return
		}
		c, err := s.Verify(tok)
		if err != nil {
			platform.WriteError(w, http.StatusUnauthorized, err.Error())
			return
		}
		if !c.Has(role) {
			platform.WriteError(w, http.StatusForbidden, "missing role "+role)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, c)))
	})
}
