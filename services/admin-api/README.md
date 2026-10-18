# admin-api

Go service: internal admin endpoints.

- Test: `bazel test //services/admin-api/...`
- Binary: `bazel run //services/admin-api/cmd/admin-api` (listens on `PORT`, default 8080).
- Depends on: `libs/go/platform`, `libs/go/authz`, `libs/go/db`.
