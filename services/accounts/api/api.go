// Package api defines the accounts service's wire types.
package api

// User is a member of one org.
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Org   string `json:"org"`
	Role  string `json:"role"`
}

// Signup is the body of POST /signup: it creates an org and its first admin.
type Signup struct {
	Email string `json:"email"`
	Org   string `json:"org"`
}

// AddMember is the body of POST /members.
type AddMember struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// Session is returned with a fresh token.
type Session struct {
	User  User   `json:"user"`
	Token string `json:"token"`
}
