# mail-tracker

Go service: email tracking, classification hooks.

- Test: `bazel test //services/mail-tracker/...`
- Binary: `bazel run //services/mail-tracker/cmd/mail-tracker` (listens on `PORT`, default 8080).
- Depends on: `libs/go/platform`, `libs/go/authz`, `libs/go/db`, `libs/go/aiclient`.
- Its `api` package is used by: search.
