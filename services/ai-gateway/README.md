# ai-gateway

Go service: routes requests to LLM providers, quotas, retries.

- Test: `bazel test //services/ai-gateway/...`
- Binary: `bazel run //services/ai-gateway/cmd/ai-gateway` (listens on `PORT`, default 8080).
- Depends on: `libs/go/platform`, `libs/go/authz`, `libs/go/aiclient`.
