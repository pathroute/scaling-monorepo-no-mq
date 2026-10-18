// Package api defines the admin API's wire types.
package api

import "time"

// Flag is a feature flag.
type Flag struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// SetFlag is the body of PUT /admin/flags/{name}.
type SetFlag struct {
	Enabled bool `json:"enabled"`
}

// AuditEntry records one admin change.
type AuditEntry struct {
	Seq    int       `json:"seq"`
	Actor  string    `json:"actor"`
	Action string    `json:"action"`
	At     time.Time `json:"at"`
}
