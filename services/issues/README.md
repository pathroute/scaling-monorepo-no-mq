# issues

Go service: issue tracker.

- Test: `bazel test //services/issues/...`
- Binary: `bazel run //services/issues/cmd/issues` (listens on `PORT`, default 8080).
- Depends on: `libs/go/platform`, `libs/go/authz`, `libs/go/db`, `libs/go/aiclient`.
- Its `api` package is used by: search.
