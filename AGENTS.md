# Quoriva backend guidance

This repository owns the Go modular monolith, PostgreSQL migrations, backend tests, and API image. Shared requirements, API contracts, full-stack Compose, system tests, and project orchestration belong to the main Quoriva repository.

## Context and authorization

Inspect this repository's working tree before edits. In an umbrella checkout, explicitly read `../AGENTS.md`, `../docs/planning/status.md`, relevant backlog/decision entries, product invariants and architecture before changing behavior. A standalone checkout must locate or request a matching main-repository checkout when material product context is unavailable; do not assume outer instructions are automatically loaded or invent pending defaults.

The current authorization is the repository-organization refactor; QV-001 is implemented. Do not start subsequent features automatically. No commit, push, merge, deployment, live-data mutation or GitHub changes without explicit authorization. Preserve unrelated work and never commit secrets. Later explicit user instructions and the main status determine future scope.

## Implementation and verification

Use idiomatic Go net/http, explicit errors, cancellation/deadlines, parameterized SQL and understandable transaction boundaries. Keep inventory/reservations/checkout in one backend transaction boundary. Database constraints and transactions must enforce seat ownership across replicas; process-local synchronization is insufficient. Publish availability only after commit. Do not change finalized product rules without explicit approval.

Go 1.27.3 is pinned by go.mod and the Docker build. Relevant checks are gofmt, `go vet ./...`, and `go test -race -count=1 ./...` in a supported environment. Database changes require real PostgreSQL checks. Set `TEST_DATABASE_URL` and `REQUIRE_DATABASE_TESTS=1` when database verification is required; otherwise the integration test deliberately skips. It creates and removes a unique schema, so use a disposable test database/account with schema privileges. Race detection complements database tests and does not prove distributed correctness.

See README.md for API startup, migrations, configuration and umbrella commands. Report exact checks and outcomes; static configuration review is not runtime or hosted-CI evidence.

## AI workflow and learning

Use the main development workflow for cross-repository increments; read its checked-in skill sources if installed skills are unavailable. Give shared schema/integration work one owner and delegate only independent tasks. Explain significant Go/database/concurrency choices before implementation, then connect actual code and verification to a hands-on exercise. Respect an explicit teaching pause; otherwise continue authorized work. Never infer learner mastery from generated explanations or agent-run tests.
