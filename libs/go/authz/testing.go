package authz

import "time"

// TestSigner is a fixed-key signer for tests and local runs; never use it in production.
var TestSigner = Signer{Key: []byte("test-only-key")}

// TestToken issues a one-hour TestSigner token for subject in org with roles.
func TestToken(subject, org string, roles ...string) string {
	tok, err := TestSigner.Issue(Claims{Subject: subject, Org: org, Roles: roles, Expires: time.Now().Add(time.Hour)})
	if err != nil {
		panic(err)
	}
	return tok
}
