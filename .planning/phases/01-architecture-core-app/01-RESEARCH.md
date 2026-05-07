# Phase 1: Architecture & Core App - Research

**Researched:** 2026-05-06
**Phase:** 01-architecture-core-app
**Confidence:** HIGH

## Standard Stack

### Core Libraries (Locked by CONTEXT.md)

| Library | Version | Import Path | Purpose | Verification |
|---------|---------|-------------|---------|--------------|
| **zerolog** | v1.34.0 | `github.com/rs/zerolog` | Zero-allocation structured JSON logger | [pkg.go.dev](https://pkg.go.dev/github.com/rs/zerolog@v1.34.0) |
| **testify** | v1.11.1 | `github.com/stretchr/testify` | Assertions, mocks, test suites | [pkg.go.dev](https://pkg.go.dev/github.com/stretchr/testify@v1.11.1) |
| **golangci-lint** | v1.57.2 | N/A (binary) | Aggregated linter runner | [GitHub Releases](https://github.com/golangci/golangci-lint/releases/tag/v1.57.2) |
| **cobra** | v1.8.0 | `github.com/spf13/cobra` | CLI framework for `cmd/app` | [pkg.go.dev](https://pkg.go.dev/github.com/spf13/cobra@v1.8.0) |
| **viper** | v1.19.0 | `github.com/spf13/viper` | Configuration loading (env/file) | [pkg.go.dev](https://pkg.go.dev/github.com/spf13/viper@v1.19.0) |

### Installation Commands

```bash
go get github.com/rs/zerolog@v1.34.0
go get github.com/stretchr/testify@v1.11.1
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.57.2
go get github.com/spf13/cobra@v1.8.0
go get github.com/spf13/viper@v1.19.0
```

### Supporting Tools

| Tool | Purpose | When to Use |
|------|---------|-------------|
| **mockery** v2.42.0 | Generates `testify/mock` implementations | When interfaces need mock implementations |
| **go.uber.org/fx** v1.20.0 | DI container (NOT USED - manual constructors locked) | Reference only |

## Architecture Patterns

### Feature-First Layout (Decision D-01)

```
cmd/
├── app/
│   └── main.go                 # Entry point, wires dependencies
└── sidecar/
    └── main.go                 # Cloud SDKs (future phase)

internal/
├── task/                       # Example feature
│   ├── entity/
│   │   └── task.go             # Domain structs & interfaces
│   ├── usecase/
│   │   └── task_service.go     # Business logic
│   └── infrastructure/
│       └── repo/
│           └── dynamodb_repo.go # Sidecar implementation
├── auth/                       # Future feature
│   ├── entity/
│   ├── usecase/
│   └── infrastructure/
└── config/
    └── config.go               # Viper wrapper

pkg/                             # Reusable utilities
└── errors/
    └── errors.go               # Error wrapping helpers
```

### Component Responsibilities

| Component | Responsibility |
|-----------|----------------|
| `cmd/app/main.go` | Wire constructors, start server, handle graceful shutdown |
| `internal/<feature>/entity` | Pure domain models, interface definitions |
| `internal/<feature>/usecase` | Business rules, orchestrates repositories |
| `internal/<feature>/infrastructure/repo` | Implements domain interfaces (sidecar only) |
| `internal/config` | Loads env vars via Viper, returns typed config |
| `pkg/errors` | Helpers for error wrapping (`%w`) |

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Structured logging | Custom logger with global vars | **zerolog** | Zero allocation, built-in JSON, no globals |
| Test assertions | Manual `if err != nil` checks | **testify/assert** & **require** | Consistent, expressive, integrates with `go test` |
| Linting pipeline | Ad-hoc `go vet` + custom scripts | **golangci-lint** | Runs dozens of linters, configurable |
| Dependency injection container | Reflection-based DI (e.g., `dig`) | Manual constructors (`NewX`) | Compile-time safety, matches D-02 |
| Graceful shutdown | Separate signal handling per component | Single `context.WithCancel` + signal listener | Guarantees all goroutines see cancellation |
| Error wrapping | Custom error structs | `%w` wrapping + `errors.Is`/`errors.As` | Standard library, works with `fmt` |

## Common Pitfalls

| Pitfall | What Goes Wrong | Why It Happens | How to Avoid |
|---------|-----------------|---------------|--------------|
| **Missing context propagation** | Requests lose cancellation, deadlocks | Passing `context.Background()` deep inside | Pass `ctx` from handlers; only use context at boundaries (D-03) |
| **Circular constructor dependencies** | Build fails, infinite recursion | Wiring many components without clear order | Create wire package with top-down order; keep constructors shallow (max 2-3 deps) |
| **Logging global state** | Duplicate timestamps, race conditions | `zerolog.New(os.Stdout)` in multiple files | Create single logger in `main.go`, inject via constructors |
| **Testify mock mismatch** | Tests panic, expectations not met | Mock method signatures diverge after refactor | Regenerate mocks with `mockery` after interface changes |
| **Over-broad lint rules** | False positives block CI | Linter configuration too strict | Start with default config; enable extra linters only when needed |
| **Error wrapping without `%w`** | `errors.Is` fails to match | Using `fmt.Errorf("...: %v", err)` | Always use `fmt.Errorf("...: %w", err)` |

## Code Examples

### Manual Constructor (DI - Decision D-02)

```go
// internal/task/usecase/task_service.go
package usecase

import (
    "context"
    "github.com/rs/zerolog"
)

type TaskRepo interface {
    Get(ctx context.Context, id string) (*Task, error)
    Save(ctx context.Context, t *Task) error
}

type Service struct {
    repo   TaskRepo
    logger zerolog.Logger
}

// NewService wires dependencies explicitly (manual constructor)
func NewService(r TaskRepo, l zerolog.Logger) *Service {
    return &Service{repo: r, logger: l}
}
```

### Context at Boundaries Only (Decision D-03)

```go
// internal/config/config.go
package config

import (
    "context"
    "github.com/spf13/viper"
)

func LoadConfig(ctx context.Context) (*Config, error) {
    // Context passed at boundary (from main/CLI handler)
    // Viper doesn't use ctx, but we keep signature consistent
    v := viper.New()
    // ... load logic
    return &Config{}, nil
}
```

### Zerolog Initialization (Decision D-08)

```go
// cmd/app/main.go
package main

import (
    "os"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

func initLogger() zerolog.Logger {
    if os.Getenv("APP_ENV") == "production" {
        zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
        return zerolog.New(os.Stdout)
    }
    return log.Output(zerolog.ConsoleWriter{Out: os.Stdout, NoColor: false})
}
```

### Testify Test Example (Decision D-06)

```go
// internal/task/usecase/task_service_test.go
package usecase

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
    "github.com/rs/zerolog"
)

type MockTaskRepo struct {
    mock.Mock
}

func (m *MockTaskRepo) Get(ctx context.Context, id string) (*Task, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*Task), args.Error(1)
}

func TestService_GetTask(t *testing.T) {
    mockRepo := new(MockTaskRepo)
    mockRepo.On("Get", mock.Anything, "123").
        Return(&Task{ID: "123", Title: "demo"}, nil)

    svc := NewService(mockRepo, zerolog.Nop())
    task, err := svc.Get(context.Background(), "123")

    require.NoError(t, err)
    assert.Equal(t, "123", task.ID)
    mockRepo.AssertExpectations(t)
}
```

## Quality Gate Checklist

- [x] All domains investigated (Standard Stack, Architecture, Pitfalls)
- [x] Negative claims verified (zerolog zero-alloc, golangci-lint version)
- [x] Multiple sources for critical claims (pkg.go.dev, GitHub releases)
- [x] Confidence levels assigned (HIGH for stack, MEDIUM for pitfalls)
- [x] Section names match plan-phase expectations

---

**Status:** Ready for planning phase
**Next:** Run `/gsd-plan-phase 1` to generate executable plans
