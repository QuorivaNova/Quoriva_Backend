# Quoriva backend

Go net/http API and PostgreSQL migrations for Quoriva. QV-001 provides liveness, database/schema readiness, and a frontend status endpoint. Business features are separate future increments.

## Source map

- `cmd/server/main.go`: HTTP executable setup and graceful shutdown.
- `cmd/migrate/main.go`: separate migration executable.
- `internal/config`: environment settings; `internal/database`: pool and readiness query; `internal/handlers`: HTTP routes and responses.
- `migrations`: embedded SQL and migration runner.

`cmd` is the usual Go location for executable entry points; `internal` provides a Go-enforced import boundary. The server entry point was renamed from `cmd/api`, and `internal/httpapi` is now `internal/handlers`.

## Work independently

Use Go 1.27.3 and a development PostgreSQL 18.6 database. The Go toolchain can download the module-selected version if supported by your local Go installation. From this repository root in PowerShell, configure a URL for your own disposable development database:

```powershell
$env:DATABASE_URL = 'postgres://quoriva:local_password@localhost:5432/quoriva?sslmode=disable&connect_timeout=2'
$env:HTTP_ADDR = ':8080'
go run ./cmd/migrate
go run ./cmd/server
```

The URL above is an example, not a provisioned account. Use local credentials and do not commit them. Migrations are embedded in the executable and applied by golang-migrate; rerunning applies pending changes or reports the schema is current. Review a failed/dirty migration instead of deleting its version record.

`GET /health/live` returns 200 while the API is responsive. `GET /health/ready` and `GET /api/status` require the expected foundation record and return 503 when the dependency/schema is unavailable. Database checks are bounded and responses omit internal connection details. The API uses a shared pool and handles graceful shutdown.

## Checks

```powershell
gofmt -l cmd internal migrations
go vet ./...
go test ./...
```

gofmt should print no files. To require real PostgreSQL integration checks, configure a disposable test database:

```powershell
$env:TEST_DATABASE_URL = $env:DATABASE_URL
$env:REQUIRE_DATABASE_TESTS = '1'
go test -race -count=1 -v ./...
```

The integration test creates/drops a unique temporary schema and requires schema privileges. It checks migrations, readiness and migration lock contention. Without `TEST_DATABASE_URL`, ordinary tests skip the database test; setting `REQUIRE_DATABASE_TESTS=1` makes missing configuration fail. Race detection needs a supported C toolchain; the main repository's Linux test container supplies it.

## Work in the main Quoriva checkout

When this folder is `Quoriva_backend` inside the main repository, use its Compose configuration from the parent:

```powershell
docker compose up -d --build
docker compose run --build --rm api-test
```

The main repository owns API contracts under `contracts/`, product requirements/invariants and learning notes under `docs/`, full-stack orchestration, and system browser tests. Read these from a matching main checkout rather than duplicating the specification here. The main database is internal to its Compose network by default.

## Repository state and boundaries

This folder is prepared as an independent local Git repository. Full submodule activation requires initial child commits and main-repository gitlink references; see the main refactor notes for actual activation state. Remotes will be added after GitHub organization setup. Backend changes and main integration references are separate commits, both requiring authorization.

See [AGENTS.md](AGENTS.md). The backend CI configuration supplies a disposable PostgreSQL service and requires the real DB tests; hosted execution is unverified until the repository is published and CI actually runs.
