# Headmail

English | [Korean](README.ko.md)

Headmail is an open-source email campaign management backend with a lightweight Vue.js admin frontend. It provides campaign creation, subscriber/list management, templated and personalized deliveries, and event tracking (opens/clicks). The project is written in Go for the backend and Vue 3 + Vite for the frontend.

The name "Headmail" combines "head" and "mailer" — here "head" refers to the API "head" rather than a full user-facing system. In other words, Headmail is an API-first, "head-only" mailer: it provides an API backend for creating and sending campaigns but does not include built-in account or user-management. Authentication, multi-tenancy, or any account features should be handled by an external system when needed. The system is designed for extensibility — storage backends, mailer/receiver drivers, and template systems are pluggable through well-defined interfaces so Headmail can be adapted and extended for different environments and scale requirements.

## Highlights

- Clear separation between API, service logic, and storage—enables implementing different database backends without changing business logic.
- Abstracted mail sending / receiving interfaces so different mail delivery implementations (SMTP, third-party APIs) can be plugged in easily.
- Template rendering with data merging and a small i18n helper for localized messages.
- Simple, modular code layout suitable for incremental extension and testing.

## Quick overview

Core components:
- `cmd/server` — server entrypoint
- `pkg/api` — HTTP handlers (admin and public/tracking endpoints)
- `pkg/service` — business logic (campaigns, lists, deliveries, tracking)
- `pkg/repository` — repository interfaces and transaction utilities
- `internal/db/sqlite` — SQLite repository implementation
- `pkg/template` — template rendering service (uses Go templates + sprig)
- `pkg/mailer`, `pkg/receiver` — mail sending and receiving abstractions
- `frontend` — Vue 3 admin UI (Vite)

## Concepts & data model (brief)

- Campaign: templated message with subject, HTML/text bodies, optional template reference, data map and headers.
- Template: Use MJML format templates and GrapesJS Editor.
- List: a grouping of subscribers.
- Subscriber: email + metadata, membership in multiple lists.
- Delivery: a scheduled send instance created from a campaign and recipient data.
- DeliveryEvent: events like open/click recorded for analytics.

## Requirements

- Backend:
  - Go 1.24+
- Frontend:
  - Node.js (recommended LTS) and pnpm

## Local development

Backend
1. Ensure Go is installed and `GOPATH`/modules configured.
2. Download dependencies:
   - go mod tidy
3. Run the server (development):
   - go run ./cmd/server
   - or build: go build -o headmail ./cmd/server && ./headmail

Configuration is loaded from `configs/` (see [configs/config.example.yaml](configs/config.example.yaml)).

Frontend
1. cd frontend
2. Install packages:
   - pnpm install
3. Start dev server:
   - pnpm dev

Frontend communicates with the backend API (check `frontend/src/api` and `frontend/vite.config.ts` for proxy / base URL settings).

## Tests

- Run Go tests:
  - go test ./...
- Note: tests use external Go modules (for example, `github.com/stretchr/testify`); `go test` will fetch them if not present.

## Extending storage / mail drivers

- Storage: implement the `repository.DB` and related repository interfaces in `pkg/repository/interfaces.go` to add another backend (Postgres, MySQL, etc.). The service layer depends only on the interfaces.
- Mailer/Receiver: implement the interfaces in `pkg/mailer` and `pkg/receiver` to add alternate transport mechanisms or third-party integrations.

## Operational considerations

- Template safety: templates are executed server-side — validate and sanitize data used in templates to avoid exposing secrets accidentally.
- Large lists: some repo/service functions currently load subscribers in memory or stream via channels. For very large lists, implement pagination/streaming and batched delivery creation.
- Rate-limiting / backoff: integrate with a queue or worker pool for high-volume sending and retry logic.

## Project structure

- `cmd/server/main.go` — server startup and wiring
- `pkg/server` — server runtime and worker
- `pkg/service/*` — business logic implementations
- `pkg/repository/interfaces.go` — repository interfaces and filters
- `internal/db/sqlite` — reference SQLite repository implementations
- `pkg/template/template.go` — template rendering service and tests

## Contributing

- Use gofmt/goimports for formatting.
- Include unit tests for new logic.
- Open issues or PRs on the repository.

## License

AGPL-3.0-or-later (see [LICENSE](LICENSE))
