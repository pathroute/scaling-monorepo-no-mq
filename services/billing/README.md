# billing

Go service: invoices, subscriptions, usage-based pricing.

- Test: `bazel test //services/billing/...`
- Binary: `bazel run //services/billing/cmd/billing` (listens on `PORT`, default 8080).
- Depends on: `libs/go/platform`, `libs/go/authz`, `libs/go/db`.
