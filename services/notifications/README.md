# notifications

Go service: email/push/webhook fan-out.

- Test: `bazel test //services/notifications/...`
- Binary: `bazel run //services/notifications/cmd/notifications` (listens on `PORT`, default 8080).
- Depends on: `libs/go/platform`, `libs/go/db`.
