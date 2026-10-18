# accounts

Go service: users, orgs, auth tokens.

- Test: `bazel test //services/accounts/...`
- Binary: `bazel run //services/accounts/cmd/accounts` (listens on `PORT`, default 8080).
- Depends on: `libs/go/platform`, `libs/go/authz`, `libs/go/db`.
