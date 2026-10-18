// Package api defines the ai-gateway's wire types.
package api

import "github.com/pathroute/suite/libs/go/aiclient"

// Complete is the body of POST /v1/complete.
type Complete = aiclient.Request

// Completion is the response of POST /v1/complete.
type Completion struct {
	aiclient.Response
	Provider       string `json:"provider"`
	QuotaRemaining int    `json:"quota_remaining"`
}
