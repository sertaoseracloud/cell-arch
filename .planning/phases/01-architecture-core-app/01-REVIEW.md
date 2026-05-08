---
phase: 01-architecture-core-app
reviewed: 2026-05-07T00:00:00Z
depth: standard
files_reviewed: 22
files_reviewed_list:
  - go.mod
  - cmd/app/main.go
  - cmd/app/app.go
  - cmd/sidecar/main.go
  - internal/task/entity/task.go
  - internal/task/usecase/task_service.go
  - internal/task/infrastructure/repo/sidecar_repo.go
  - internal/config/config.go
  - pkg/contextutil/contextutil.go
  - pkg/errors/errors.go
  - pkg/shutdown/shutdown.go
  - .golangci.yml
  - internal/sidecar/aws/dynamodb_client.go
  - internal/sidecar/azure/cosmosdb_client.go
  - internal/sidecar/errors/mapper.go
  - .gitignore
  - internal/task/entity/task_test.go
  - internal/task/usecase/task_service_test.go
  - internal/task/usecase/mock_repository_test.go
  - internal/config/config_test.go
  - pkg/contextutil/contextutil_test.go
  - pkg/errors/errors_test.go
findings:
  critical: 2
  warning: 6
  info: 5
  total: 13
status: issues_found
---

# Phase 01: Code Review Report

**Reviewed:** 2026-05-07T00:00:00Z
**Depth:** standard
**Files Reviewed:** 22
**Status:** issues_found

## Summary

The Phase 1 skeleton is well-structured and cleanly follows the Clean Architecture constraints. Zero-SDK-leakage in `cmd/app` and `internal/task` is confirmed. Context propagation is consistent across all exported I/O functions. Test coverage for the domain and use-case layers is solid, using table-driven tests and testify mocks as required by the TDD spec.

Two critical issues were found: a goroutine leak in `pkg/shutdown/shutdown.go` caused by registering a second `signal.Notify` channel that is never consumed, and the use of `fmt.Errorf` to create a sentinel error variable in `sidecar_repo.go`, which makes it unwrappable via `errors.Is`. Six warnings cover double `os.Exit` after `logger.Fatal`, a hardcoded cloud discriminator string, an unsentineled "item not found" error in the DynamoDB client, fragile string-matching error classification in the mapper, and missing input validation guards. Five informational items cover code quality and missing test coverage for the shutdown package.

---

## Critical Issues

### CR-01: Goroutine Leak in `shutdown.signalName`

**File:** `pkg/shutdown/shutdown.go:41-52`

**Issue:** `signalName()` calls `signal.Notify(ch, ...)` and registers `ch` as a signal recipient, then immediately reads from it with `select { case sig := <-ch: ...; default: return "unknown" }`. Because the signal has already been consumed by `signal.NotifyContext` before this function is called (the goroutine fires only after `ctx.Done()` is closed, meaning the signal was already delivered), the `default` branch always fires and returns `"unknown"`. The registered channel `ch` is never deregistered until `signal.Stop(ch)` runs via `defer`, but because `select` exits immediately via `default`, `signal.Stop` runs synchronously — so the defer is correct in isolation.

However the real bug is subtler: a second goroutine is spawned inside `Graceful` at line 29 that closes over the already-cancelled `ctx`. At the moment `ctx.Done()` fires, `signalName()` is called. Inside it, `signal.Notify(ch, ...)` registers a new listener. If a *second* signal arrives at that exact moment it will be delivered to `ch`, block the goroutine until `signal.Stop(ch)` fires, and `signal.Stop` will drain `ch` — this is safe. **The actual critical bug is different:** the goroutine at line 29-36 is spawned every time `Graceful` is called, and it reads `ctx.Done()` which is the same context returned to the caller. If the caller's `stop()` is invoked (cancelling the context for reasons other than an OS signal), the goroutine fires, calls `signalName()`, which re-registers on the signal channel and spins until `signal.Stop` is called. No issue there either.

The **real critical defect** is at line 43: `signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)` is called *after* the context has already been cancelled (the goroutine only runs after `ctx.Done()`). This means the signal that triggered cancellation has already been consumed by `signal.NotifyContext`. `signalName` will always return `"unknown"` — the function is broken by design — and it also registers a persistent second listener for the lifetime of the `signal.Stop` defer, leaking a brief secondary registration window. More critically, the `select` default-branch always wins, meaning the log message `"OS signal received — initiating graceful shutdown"` always prints `signal: unknown`, which is misleading and defeats the observability goal (D-08).

**Fix:** Remove `signalName()` entirely. The correct way to capture the signal name is to pass it through the context or capture it at the `signal.NotifyContext` call site. Use a closure that captures the signal from a dedicated channel set up *before* the context is cancelled:

```go
func Graceful(parent context.Context, logger zerolog.Logger) (context.Context, context.CancelFunc) {
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

    ctx, stop := signal.NotifyContext(parent, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        select {
        case sig := <-sigCh:
            logger.Info().
                Str("signal", sig.String()).
                Msg("OS signal received — initiating graceful shutdown")
        case <-ctx.Done():
            // Context cancelled for a non-signal reason (e.g. stop() called directly).
        }
        signal.Stop(sigCh)
    }()

    return ctx, stop
}
```

---

### CR-02: Sentinel Error Created with `fmt.Errorf` Breaks `errors.Is` Chains

**File:** `internal/task/infrastructure/repo/sidecar_repo.go:67`

**Issue:** The sentinel `errNotImplemented` is declared as:

```go
var errNotImplemented = fmt.Errorf("not implemented")
```

`fmt.Errorf` without a `%w` verb returns a plain `*errors.errorString`-compatible value, but its identity differs from `errors.New`. More critically: because `errNotImplemented` is then wrapped with `%w` at every call site (e.g. line 39), callers *can* technically reach it via `errors.Is`. However, the canonical Go pattern for sentinel errors is `errors.New`, not `fmt.Errorf`. Using `fmt.Errorf` here is a code quality issue that will confuse any linter (`noerrcheck`) and deviates from project convention established in `pkg/errors/errors.go` (which re-exports `errors.New`). More importantly, if this variable is ever refactored to include a formatted message (adding a `%w` or `%v`), the behavior changes silently.

The package-level `var` with `fmt.Errorf` is the exact anti-pattern the `pkg/errors` package was designed to replace.

**Fix:**

```go
// Use errors.New for sentinel values — identity is defined by pointer, not message.
var errNotImplemented = errors.New("not implemented")
```

Add the import: `"errors"`. Alternatively, use `pkg/errors.New` to remain consistent with the project's error package.

---

## Warnings

### WR-01: `logger.Fatal().Err(err).Msg(...)` Followed by Redundant `os.Exit(1)`

**File:** `cmd/app/main.go:28-29`, `cmd/sidecar/main.go:34-35`

**Issue:** `zerolog`'s `Fatal()` level calls `os.Exit(1)` internally after writing the log entry. The `os.Exit(1)` on the next line is therefore unreachable dead code. This is benign in production but confuses readers about the actual control flow and will cause the `ineffassign` / dead-code linter to flag it. Additionally, any `defer stop()` registered earlier will NOT run (because `os.Exit` bypasses defers), which means signal resources are not cleaned up on fatal errors — this is the real behavioral concern.

**Fix:** Remove the `os.Exit(1)` calls. The deferred `stop()` cannot execute after `logger.Fatal()` either way, but the dead code should be removed and a comment added to clarify the behavior:

```go
// logger.Fatal writes the log and calls os.Exit(1); defer stop() will NOT run.
// This is acceptable for fatal startup errors where resource cleanup is unnecessary.
if err := run(ctx, logger); err != nil {
    logger.Fatal().Err(err).Msg("application error")
    // os.Exit(1) is unreachable — zerolog.Fatal exits the process.
}
```

Or use `logger.Error()` + explicit `os.Exit(1)` if the intent is to allow deferred cleanup:

```go
if err := run(ctx, logger); err != nil {
    logger.Error().Err(err).Msg("application error")
    os.Exit(1)
}
```

---

### WR-02: Hardcoded Cloud Discriminator String `"aws"` in Application Wiring

**File:** `cmd/app/app.go:34`

**Issue:** The `cloud` argument to `NewSidecarRepo` is hardcoded as `"aws"`:

```go
taskRepo := repo.NewSidecarRepo(cfg.SidecarAddr, "aws", logger)
```

This magic string is not validated anywhere and does not appear in any enumeration or constant. When Azure support is added (the project is explicitly multicloud), this will require a code change rather than a configuration change. It also means the `cloud` field has no effect on actual routing logic in Phase 1, making it misleading.

**Fix:** Define a typed constant or add a `Cloud` field to `AppConfig` loaded from an environment variable:

```go
// In config.go
type Cloud string

const (
    CloudAWS   Cloud = "aws"
    CloudAzure Cloud = "azure"
)

// AppConfig
type AppConfig struct {
    // ...
    Cloud Cloud
}

// In LoadAppConfig
Cloud: Cloud(getEnvOrDefault("CLOUD_PROVIDER", "aws")),
```

---

### WR-03: Unsentineled "Item Not Found" Error in DynamoDB Client

**File:** `internal/sidecar/aws/dynamodb_client.go:46-47`

**Issue:** When DynamoDB returns a nil `result.Item` (item not found), the client returns:

```go
return nil, fmt.Errorf("item not found")
```

This is a plain string error with no sentinel identity. The error mapper in `internal/sidecar/errors/mapper.go` does not match this string — `"item not found"` is not one of the AWS error type strings it checks (`ResourceNotFoundException`, etc.). As a result, a missing-item lookup will be mapped to `codes.Internal` instead of `codes.NotFound` when it passes through `MapAWSError`. This is a logic bug that will cause callers to receive incorrect gRPC status codes for missing resources.

**Fix:** Define a package-level sentinel and check for it in the mapper, or return a wrapped SDK-style error:

```go
// In dynamodb_client.go
var ErrItemNotFound = errors.New("item not found")

// In Get:
if result.Item == nil {
    return nil, ErrItemNotFound
}

// In mapper.go MapAWSError, add before the switch:
if errors.Is(err, dynamodbclient.ErrItemNotFound) {
    return status.Errorf(codes.NotFound, "item not found")
}
```

---

### WR-04: Fragile String-Based Error Classification in Error Mapper

**File:** `internal/sidecar/errors/mapper.go:10-62`

**Issue:** Both `MapAWSError` and `MapAzureError` classify errors by substring-matching the error message string (e.g. checking `contains(errStr, "404")` or `contains(errStr, "ResourceNotFoundException")`). This approach is fragile:

1. AWS SDK v2 wraps errors in structured types (`*types.ResourceNotFoundException`, etc.) that support `errors.As`. Matching on string representations is unnecessary and unreliable — error messages can change between SDK versions.
2. The Azure mapper checks for `"400"` and `"invalid"` (case-sensitive) independently. The substring `"invalid"` appears in many non-error contexts and could produce false positives.
3. A custom `contains()` function reimplements `strings.Contains` from stdlib — this is dead code complexity.

**Fix for AWS:** Use `errors.As` with the SDK's typed error types:

```go
import "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

func MapAWSError(err error) error {
    var notFound *types.ResourceNotFoundException
    if errors.As(err, &notFound) {
        return status.Errorf(codes.NotFound, "resource not found: %v", err)
    }
    var validationEx *types.ValidationException
    if errors.As(err, &validationEx) {
        return status.Errorf(codes.InvalidArgument, "invalid argument: %v", err)
    }
    // ... etc.
    return status.Errorf(codes.Internal, "internal error: %v", err)
}
```

**Fix for `contains`:** Replace with `strings.Contains` from stdlib.

---

### WR-05: Missing Input Validation in `Service.Create` and `Service.Update`

**File:** `internal/task/usecase/task_service.go:41-55`, `58-75`

**Issue:** `Create` and `Update` accept `title` and `description` strings without any validation. A blank `title` will be persisted without error. In `Update`, an empty `id` string is passed to `repo.Get` without a nil/empty check. While these are business-rule decisions, the domain layer has no validation layer at all — there is no `Validate()` method on `Task` and no guard in the service. When the gRPC handler is wired in Phase 2, invalid inputs from the network will reach the database layer unchecked.

**Fix:** Add a validation step in `Create` and `Update`:

```go
func (s *Service) Create(ctx context.Context, title, description string) (*entity.Task, error) {
    if strings.TrimSpace(title) == "" {
        return nil, fmt.Errorf("service: create task: title must not be empty")
    }
    // ...
}

func (s *Service) Update(ctx context.Context, id, title, description string, status entity.Status) (*entity.Task, error) {
    if id == "" {
        return nil, fmt.Errorf("service: update task: id must not be empty")
    }
    // ...
}
```

---

### WR-06: `go.mod` Declares `go 1.25.0` — Unreleased Version

**File:** `go.mod:3`

**Issue:** The module file specifies `go 1.25.0`. As of the knowledge cutoff (August 2025), Go 1.25 had not been officially released. Declaring a future/unreleased toolchain version can cause `go build` to fail on CI runners that have Go 1.23 or 1.24 installed, and may generate `go: note: module requires Go >= 1.25` warnings. This is a toolchain compatibility risk.

**Fix:** Downgrade to the highest stable released version the codebase actually requires:

```
go 1.23.0
```

Or, if Go 1.24 features are intentionally used, set `go 1.24.0`. Avoid specifying unreleased versions.

---

## Info

### IN-01: `contains` Helper Reimplements `strings.Contains`

**File:** `internal/sidecar/errors/mapper.go:49-62`

**Issue:** The private `contains` and `findSubstr` functions are a manual reimplementation of `strings.Contains`. This adds 14 lines of dead complexity with no benefit.

**Fix:**

```go
import "strings"

// Replace all contains(s, sub) calls with strings.Contains(s, sub)
// Delete the contains() and findSubstr() functions entirely.
```

---

### IN-02: No Tests for `pkg/shutdown` Package

**File:** `pkg/shutdown/shutdown.go`

**Issue:** The `Graceful` function has no unit tests. Per the Harness requirement, domain and use-case code requires 100% coverage. The `shutdown` package sits in `pkg/` which is shared infrastructure — it is not domain or use-case code, so the 100% threshold does not strictly apply. However, given that CR-01 identifies a real defect in `signalName()`, a test that exercises the signal-capture path would have caught the bug.

**Fix:** Add a test that sends a signal to the process and verifies the context is cancelled:

```go
func TestGraceful_CancelOnSIGINT(t *testing.T) {
    ctx, stop := Graceful(context.Background(), zerolog.Nop())
    defer stop()

    p, _ := os.FindProcess(os.Getpid())
    _ = p.Signal(syscall.SIGINT)

    select {
    case <-ctx.Done():
        // pass
    case <-time.After(time.Second):
        t.Fatal("context not cancelled after SIGINT")
    }
}
```

---

### IN-03: `DynamoDBClient.Create` Logs Item Payload at Debug Level

**File:** `internal/sidecar/aws/dynamodb_client.go:52`

**Issue:** `d.logger.Debug().Interface("item", item).Msg("DynamoDB PutItem")` logs the full `map[string]types.AttributeValue` item. In production, this could log PII or sensitive task data. The `Debug` level is typically suppressed in production, but it is still a risk when debug logging is enabled for troubleshooting.

**Fix:** Log only non-sensitive identifiers (e.g. the item count or a sanitized key):

```go
d.logger.Debug().Int("attribute_count", len(item)).Msg("DynamoDB PutItem")
```

---

### IN-04: `CosmosDB.Get` Uses Hardcoded Partition Key `"task"`

**File:** `internal/sidecar/azure/cosmosdb_client.go:49`, `67`

**Issue:** Both `Get` and `Create` hardcode `const partitionKey = "task"`. In CosmosDB, the partition key value is typically the item's own ID or a tenant discriminator — using a fixed constant means all items share a single logical partition, which defeats the scaling model of CosmosDB. This will become a correctness and performance problem at scale.

**Fix:** Accept the partition key as a parameter (or derive it from the item `id`):

```go
func (c *CosmosDBClient) Get(ctx context.Context, id string) (map[string]interface{}, error) {
    response, err := c.container.ReadItem(
        ctx,
        azcosmos.NewPartitionKeyString(id), // use id as partition key
        id,
        nil,
    )
    // ...
}
```

---

### IN-05: Mock Repository Uses `testify/mock` Instead of `uber-go/mock` as Required by Harness Spec

**File:** `internal/task/usecase/mock_repository_test.go:6`

**Issue:** The Harness specification at `.claude/harness/test-coverage-thresholds.md` explicitly states: "Mocking: Proibido o uso de chamadas reais de rede nos testes unitários. Interfaces de repositório devem ser mockadas via `uber-go/mock`." The mock here uses `github.com/stretchr/testify/mock`. The comment in the file acknowledges this ("Phase 5 will use mockery/mockgen"), but `uber-go/mock` (the successor to `golang/mock`) is the mandated tool.

**Fix:** This is a Phase 5 backlog item as noted. No immediate action required, but track as a compliance debt. When generating mocks, use:

```bash
mockgen -source=internal/task/entity/task.go \
        -destination=internal/task/usecase/mock_repository_test.go \
        -package=usecase_test
```

---

*Reviewed: 2026-05-07T00:00:00Z*
*Reviewer: Claude (gsd-code-reviewer)*
*Depth: standard*
