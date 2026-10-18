# search

Go service: search API over todo, issues and mail.

- Test: `bazel test //services/search/...`
- Binary: `bazel run //services/search/cmd/search` (listens on `PORT`, default 8080).
- Depends on: `libs/go/platform`, `libs/go/authz`, `services/todo/api`, `services/issues/api`, `services/mail-tracker/api`.
