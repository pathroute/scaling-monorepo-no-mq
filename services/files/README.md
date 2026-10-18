# files

Go service: attachments, blob storage.

- Test: `bazel test //services/files/...`
- Binary: `bazel run //services/files/cmd/files` (listens on `PORT`, default 8080).
- Depends on: `libs/go/platform`, `libs/go/authz`, `libs/go/db`.
