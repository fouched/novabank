# Novabank Workspace System Guide

## Architecture Profile
- **Pattern:** Isolated Microservices Monorepo via Go Workspaces (`go.work`).
- **Core Modules:**
    - `shared/` -> Local foundational utilities (Consul discovery, messaging wrappers).
    - `user-service/` -> Primary API handling user/account state boundaries.
    - `transaction-service/` -> Ledger ledger API state boundaries.
    - `notification-service/` -> Pure worker consumer loop listening on Kafka events.

## Strict Operational Rules
1. **Never violate database isolation:** Services must keep completely independent `migrations/` directories and explicit DB schemas (`novabank_users` vs `novabank_transactions`). Never share tables or mix connections.
2. **Shared Platform Helpers:** Code regarding custom Kafka driver tools or Consul registration layers must be written *only* within `/shared`. Import via `novabank/shared` natively inside microservice modules.
3. **Task Entry Points:**
    - User & Transaction services execute HTTP routing loops out of `cmd/api/main.go`.
    - Notification service runs a blocking worker sequence out of `cmd/worker/main.go`. Do NOT give it an HTTP router or listener ports.

## Critical Build & Execution Commands
- **Initialize Workspace Mapping:** `go work init`
- **Register Local Modules:** `go work use ./shared ./user-service ./transaction-service ./notification-service`
- **Tidy Active Module Dependencies:** `cd <module-dir> && go mod tidy`
- **Build Core App Tasks:** `go build -o bin/app ./cmd/...` inside specific module roots.
- **Run Local Test Suites:** `go test ./...`

## Contextual Guardrails (Token Mitigation)
- Always prioritize a **Plan Phase** mapping file structural shifts across Go workspace scopes before writing logic changes.
- Never write redundant inline setup blocks or duplicate boilerplate configurations already covered within `/shared`.
