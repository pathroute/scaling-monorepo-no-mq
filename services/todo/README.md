# todo

Go service: "AI-enabled" todo lists.

- Test: `bazel test //services/todo/...`
- Binary: `bazel run //services/todo/cmd/todo` (listens on `PORT`, default 8080).
- Depends on: `libs/go/platform`, `libs/go/authz`, `libs/go/db`, `libs/go/aiclient`.
- Its `api` package is used by: search.
