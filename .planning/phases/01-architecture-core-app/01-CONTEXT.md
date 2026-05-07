# Phase 1: Architecture & Core App - Context

**Gathered:** 2026-05-06
**Status:** Ready for planning

<domain>
## Phase Boundary

This phase delivers the foundational Go application structure using Clean Architecture principles. It establishes the project layout, dependency injection strategy, context propagation rules, graceful shutdown handling, error handling style, testing framework, linting configuration, and logging library. The main application must compile and run with zero cloud SDK imports.
</domain>

<decisions>
## Implementation Decisions

### Project Layout

- **D-01:** Feature-first layout (group by feature/domain, e.g., `task/`, `auth/`, `sidecar/` with subfolders `entity`, `usecase`, `repo`). This provides feature isolation.

### Dependency Injection

- **D-02:** Manual constructors (explicit `NewX(dep1, dep2)` functions). Aligns with PROJECT.md decision: "Dependency Injection: All components wired via constructors, no `init()` or globals."

### Context Propagation

- **D-03:** Context passed only at boundaries (exported methods, API handlers, goroutine entry points). Not required for every internal/private function.

### Graceful Shutdown

- **D-04:** OS signal handling with `context.WithCancel`. Listen for SIGINT/SIGTERM, cancel root context to trigger cleanup.

### Error Handling Style

- **D-05:** Wrapped errors using `%w` (standard Go error wrapping). Supports `errors.Is`/`errors.As`.

### Testing Framework

- **D-06:** `testify` library. Adds assertions, mocks, and test suites on top of standard `testing` package.

### Code Formatting & Linting

- **D-07:** `golangci-lint` with a standard set of linters (includes `go vet`, `staticcheck`, etc.). Enforces code quality beyond `go fmt`.

### Logging Library

- **D-08:** `zerolog` for structured, zero-allocation JSON logging.

### Claude's Discretion

- None – all decisions were user-specified.

</decisions>

<canonical_refs>

## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Architecture

- `.planning/ROADMAP.md` — Phase 1 goals, success criteria, and plan breakdown
- `.planning/PROJECT.md` — Project overview, architecture principles, key design decisions
- `.planning/REQUIREMENTS.md` — ARCH-01 through ARCH-05, TEST-01, TEST-03 requirements

### Technical Standards

- `.claude/specs/infrastructure/app-go-clean-architecture.md` — Clean Architecture layout rules
- `.claude/specs/technical/golang-implementation-standards.md` — Go coding standards
- `.claude/specs/infrastructure/tdd-lifecycle-go.md` — TDD Red-Green-Refactor flow

### Compliance & Hardness

- `.claude/hardness/test-coverage-thresholds.md` — Required test coverage thresholds (>80%)
- `.claude/hardness/performance-budgets.md` — Performance requirements
- `.claude/hardness/security-rules.md` — Security constraints

</canonical_refs>

<code_context>

## Existing Code Insights

### Reusable Assets

- None yet — greenfield project (Phase 1).

### Established Patterns

- None yet — greenfield project (Phase 1).

### Integration Points

- `cmd/app` — main application entry point to be created
- `cmd/sidecar` — sidecar proxy (future phase, but layout should accommodate)
</code_context>

<specifics>
## Specific Ideas

No specific UI/UX ideas — this is a backend Go service.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within Phase 1 scope.

</deferred>

---
*Phase: 01-architecture-core-app*
*Context gathered: 2026-05-06*
