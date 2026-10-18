package billing

import "github.com/pathroute/suite/services/billing/api"

// Tier prices tokens up to UpTo (0 = unlimited) at CentsPerMillion.
type Tier struct {
	Name            string
	UpTo            int64
	CentsPerMillion int64
}

// Tiers is the usage-based price list.
var Tiers = []Tier{
	{Name: "free", UpTo: 100_000, CentsPerMillion: 0},
	{Name: "standard", UpTo: 10_000_000, CentsPerMillion: 200},
	{Name: "volume", UpTo: 0, CentsPerMillion: 150},
}

// Price splits tokens across tiers. Partial cents round up per line.
func Price(org string, tokens int64) api.Invoice {
	inv := api.Invoice{Org: org, Tokens: tokens}
	var prev int64
	for _, t := range Tiers {
		if tokens <= prev {
			break
		}
		n := tokens - prev
		if t.UpTo > 0 {
			n = min(tokens, t.UpTo) - prev
		}
		cents := (n*t.CentsPerMillion + 999_999) / 1_000_000
		inv.Lines = append(inv.Lines, api.Line{Tier: t.Name, Tokens: n, AmountCent: cents})
		inv.TotalCents += cents
		prev = t.UpTo
	}
	return inv
}
