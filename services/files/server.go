// Package files stores attachments (blob storage in memory for now).
package files

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pathroute/suite/libs/go/authz"
	"github.com/pathroute/suite/libs/go/db"
	"github.com/pathroute/suite/libs/go/platform"
	"github.com/pathroute/suite/services/files/api"
)

// MaxSize is the largest upload accepted, in bytes.
const MaxSize = 1 << 20

type blob struct {
	meta api.File
	data []byte
}

// Server holds the stored files, keyed by "<org>/<name>".
type Server struct {
	signer authz.Signer
	blobs  *db.Table[blob]
	now    func() time.Time
}

// NewServer returns a server with no files.
func NewServer(signer authz.Signer) *Server {
	return &Server{signer: signer, blobs: db.NewTable[blob](), now: time.Now}
}

// Handler routes the service's endpoints.
func (s *Server) Handler() http.Handler {
	mux := platform.NewMux()
	mux.Handle("GET /files", s.signer.Require(authz.RoleViewer, http.HandlerFunc(s.list)))
	mux.Handle("PUT /files/{name}", s.signer.Require(authz.RoleMember, http.HandlerFunc(s.put)))
	mux.Handle("GET /files/{name}", s.signer.Require(authz.RoleViewer, http.HandlerFunc(s.get)))
	return mux
}

func key(r *http.Request) (org, k string) {
	c, _ := authz.FromContext(r.Context())
	return c.Org, c.Org + "/" + r.PathValue("name")
}

// ValidName rejects names that could escape a directory in a real blob store.
func ValidName(n string) bool {
	return n != "" && len(n) <= 200 && !strings.ContainsAny(n, `/\`) && n != "." && n != ".."
}

func (s *Server) put(w http.ResponseWriter, r *http.Request) {
	if !ValidName(r.PathValue("name")) {
		platform.WriteError(w, http.StatusBadRequest, "invalid name")
		return
	}
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, MaxSize))
	if err != nil {
		platform.WriteError(w, http.StatusRequestEntityTooLarge, "file too large")
		return
	}
	org, k := key(r)
	sum := sha256.Sum256(data)
	b := blob{meta: api.File{Name: r.PathValue("name"), Org: org, Size: int64(len(data)), SHA256: hex.EncodeToString(sum[:]), UploadedAt: s.now()}, data: data}
	s.blobs.Tx(func(tx *db.Tx[blob]) error {
		if tx.Update(k, b) != nil {
			return tx.Insert(k, b)
		}
		return nil
	})
	platform.WriteJSON(w, http.StatusOK, b.meta)
}

func (s *Server) get(w http.ResponseWriter, r *http.Request) {
	_, k := key(r)
	b, err := s.blobs.Get(k)
	if err != nil {
		platform.WriteError(w, http.StatusNotFound, "no such file")
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("ETag", `"`+b.meta.SHA256+`"`)
	w.Write(b.data)
}

func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	org, _ := key(r)
	var out []api.File
	for _, b := range s.blobs.List(func(b blob) bool { return b.meta.Org == org }) {
		out = append(out, b.meta)
	}
	platform.WriteJSON(w, http.StatusOK, out)
}
