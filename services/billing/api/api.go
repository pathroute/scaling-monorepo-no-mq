// Package api defines the billing service's wire types.
package api

// Usage is the body of POST /usage: tokens consumed by the caller's org.
type Usage struct {
	Tokens int64 `json:"tokens"`
}

// Line is one priced tier on an invoice.
type Line struct {
	Tier       string `json:"tier"`
	Tokens     int64  `json:"tokens"`
	AmountCent int64  `json:"amount_cents"`
}

// Invoice is an org's bill for the current period.
type Invoice struct {
	Org        string `json:"org"`
	Tokens     int64  `json:"tokens"`
	Lines      []Line `json:"lines"`
	TotalCents int64  `json:"total_cents"`
}
