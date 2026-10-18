package authz

import "os"

// SignerFromEnv reads the HMAC key from AUTH_KEY, falling back to TestSigner when it's unset.
func SignerFromEnv() Signer {
	if k := os.Getenv("AUTH_KEY"); k != "" {
		return Signer{Key: []byte(k)}
	}
	return TestSigner
}
