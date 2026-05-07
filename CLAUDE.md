# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Development Commands

- **Build**: `go build ./cmd/...` – builds all binaries under `cmd/` (e.g., `app` and `sidecar`).
- **Run Application**: `go run ./cmd/app` – starts the main Go application.
- **Run Sidecar**: `go run ./cmd/sidecar` – starts the multicloud sidecar proxy.
- **Run Tests**: `go test ./...` – runs all unit and integration tests.
- **Run a Single Test**: `go test ./path/to/package -run TestName`.
- **Run Test with Race Detector**: `go test -race ./...`.
- **Generate Mocks**: `mockgen -source=internal/domain/repository.go -destination=internal/mocks/repository_mock.go -package=mocks`.
- **Lint**: `golangci-lint run` (assumes `golangci-lint` is installed).
- **Format Code**: `go fmt ./...`.
- **Check Coverage**: `go test -cover ./...` (must meet thresholds defined in `.claude/hardness/test-coverage-thresholds.md`).
- **Terraform Init**: `terraform -chdir=infra init` (infra directory contains Terraform code).
- **Terraform Plan**: `terraform -chdir=infra plan`.
- **Terraform Apply**: `terraform -chdir=infra apply`.
- **Terraform Validate**: `terraform -chdir=infra validate`.

## High‑Level Architecture Overview

The repository follows a **Clean Architecture** layout with strict separation of concerns to keep the core domain logic cloud‑agnostic.

```
repo-root/
├─ cmd/
│   ├─ app/          # Main application entry point
│   └─ sidecar/      # Multicloud sidecar proxy (AWS/Azure SDKs live here)
├─ internal/
│   ├─ domain/       # Entities and repository interfaces (pure Go, no cloud SDKs)
│   ├─ usecase/      # Business logic orchestrating domain interfaces
│   └─ infrastructure/ # Implementations of domain interfaces; talks to sidecar via HTTP/gRPC
├─ pkg/               # Reusable utilities (e.g., logging, OpenTelemetry helpers)
├─ test/              # Integration test helpers (testcontainers, emulators)
├─ infra/             # Terraform IaC modules and environment setup
├─ go.mod, go.sum     # Go module definition
└─ Makefile           # Convenience wrappers for common tasks
```

### Key Design Points

1. **Domain‑First** – All core business entities and interfaces live under `internal/domain`. No external cloud SDKs are imported here, ensuring the core can be compiled and tested without credentials.
2. **Use‑Case Layer** – `internal/usecase` contains application logic that depends only on domain interfaces.
3. **Infrastructure Layer** – Implements the domain interfaces. The actual cloud interaction (AWS SDK, Azure SDK) is isolated in `internal/infrastructure` and communicated to the main app via the sidecar.
4. **Sidecar (`cmd/sidecar`)** – A thin proxy that holds cloud SDK credentials and exposes a local HTTP/gRPC API. The main app talks to it over `localhost`, keeping the binary free of cloud‑specific dependencies.
5. **Testing Strategy** – Unit tests target the domain and use‑case layers with generated mocks. Integration tests spin up local emulators (LocalStack, CosmosDB Emulator) using `testcontainers-go` to validate the sidecar interaction.
6. **Observability** – OpenTelemetry instrumentation lives in `pkg/otel` and is wired into both the app and sidecar. Correlation IDs are propagated across the HTTP/gRPC boundary.
7. **Terraform IaC** – All infrastructure lives under `infra/` and follows the symmetric module standards described in `.claude/specs/infrastructure/terraform-standards.md`. State is locked via cloud‑native backends (S3 + DynamoDB for AWS, Blob + Lease for Azure).

## Important Project‑Specific Rules

- **Zero SDK Leakage** – The main app must never import AWS/Azure SDK packages; all cloud calls go through the sidecar.
- **Dependency Injection** – All components are wired via constructors; avoid global state or `init()` side effects.
- **Context Propagation** – Every I/O function receives a `context.Context` as the first argument.
- **Graceful Shutdown** – Both `cmd/app` and `cmd/sidecar` listen for `SIGINT/SIGTERM` and close resources cleanly.
- **Hardness Compliance** – Code must satisfy the constraints in `.claude/hardness/` (coverage, performance, security, observability). Tests that fail these checks are considered blocking.
- **Terraform Standards** – Follow the module layout and naming conventions from `.claude/specs/infrastructure/terraform-standards.md`.
- **TDD Lifecycle** – Follow the Red‑Green‑Refactor flow described in `.claude/specs/infrastructure/tdd-lifecycle-go.md`.

## Git Flow & Commit Standards

- **Commit Author** – All commits must be authored by **sertaoseracloud <engcfraposo@gmail.com>**.
- **Conventional Commits** – Use the format `<type>(<scope>): <subject>` (e.g., `feat(sidecar): add DynamoDB client`).
- **Commitlint** – The project uses Commitlint with Husky to enforce commit message conventions. See `.commitlintrc.json` and `.husky/commit-msg`.
- **Git Flow** – Follow the Git Flow branching model: `feature/`, `hotfix/`, `release/` prefixes as appropriate.
- **Default Branch** – The default branch is `main`. All work is merged into `main` via pull requests.

## Cursor / Copilot Rules (if present)

*No `.cursor/rules` or `.github/copilot-instructions.md` files were found in this repository.*

---

*Generated automatically for Claude Code guidance.*
