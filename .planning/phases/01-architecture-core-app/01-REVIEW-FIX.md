---
phase: 01-architecture-core-app
fixed_at: 2026-05-07T08:13:00Z
review_path: .planning/phases/01-architecture-core-app/01-REVIEW.md
iteration: 1
findings_in_scope: 8
fixed: 0
skipped: 8
status: none_fixed
---

# Phase 01: Code Review Fix Report

**Fixed at:** 2026-05-07T08:13:00Z
**Source review:** .planning/phases/01-architecture-core-app/01-REVIEW.md
**Iteration:** 1

**Summary:**
- Findings in scope: 8 (CR-01, CR-02, WR-01 through WR-06)
- Fixed: 0 (this iteration)
- Skipped: 8

**Note:** All findings except WR-06 were already remediated in prior commits before this fixer
iteration ran. The code was inspected and confirmed correct against the REVIEW.md expectations.
WR-06 was attempted but rolled back due to a dependency constraint (see details below).

## Fixed Issues

None — all findings were either pre-existing fixes or skipped.

## Skipped Issues

### CR-01: Goroutine Leak in `shutdown.signalName`

**File:** `pkg/shutdown/shutdown.go:41-52`
**Reason:** Code context differs from review — already fixed in prior commit `699212a`. The `signalName()` function has been removed and the implementation already uses a dedicated `sigCh` set up before `signal.NotifyContext`, matching the suggested fix exactly.
**Original issue:** `signalName()` always returned "unknown" and registered a second signal listener after the context was already cancelled.

---

### CR-02: Sentinel Error Created with `fmt.Errorf` Breaks `errors.Is` Chains

**File:** `internal/task/infrastructure/repo/sidecar_repo.go:67`
**Reason:** Code context differs from review — already fixed in prior commit `59b2139`. The sentinel is already declared as `var errNotImplemented = errors.New("not implemented")` using the standard `errors` package.
**Original issue:** `var errNotImplemented = fmt.Errorf("not implemented")` used a non-canonical pattern for sentinel errors.

---

### WR-01: `logger.Fatal().Err(err).Msg(...)` Followed by Redundant `os.Exit(1)`

**File:** `cmd/app/main.go:28-29`, `cmd/sidecar/main.go:34-35`
**Reason:** Code context differs from review — already fixed in prior commit `4a878ab`. Both files already have the explanatory comment and no `os.Exit(1)` after `logger.Fatal()`.
**Original issue:** Dead `os.Exit(1)` after zerolog's Fatal which already exits.

---

### WR-02: Hardcoded Cloud Discriminator String `"aws"` in Application Wiring

**File:** `cmd/app/app.go:34`
**Reason:** Code context differs from review — already fixed in prior commit `b2e9702`. `config.go` defines `type Cloud string` with `CloudAWS` and `CloudAzure` constants, `AppConfig` has a `Cloud Cloud` field loaded from `CLOUD_PROVIDER` env var, and `app.go` passes `string(cfg.Cloud)` to `NewSidecarRepo`.
**Original issue:** Hardcoded `"aws"` string literal with no typed constant or configuration.

---

### WR-03: Unsentineled "Item Not Found" Error in DynamoDB Client

**File:** `internal/sidecar/aws/dynamodb_client.go:46-47`
**Reason:** Code context differs from review — already fixed in prior commit `0748166`. `dynamodb_client.go` declares `var ErrItemNotFound = errors.New("item not found")` and returns it when `result.Item == nil`. `mapper.go` checks `errors.Is(err, dynamodbclient.ErrItemNotFound)` before the switch.
**Original issue:** `fmt.Errorf("item not found")` produced an unsentineled error that the mapper could not match, causing `codes.Internal` instead of `codes.NotFound`.

---

### WR-04: Fragile String-Based Error Classification in Error Mapper

**File:** `internal/sidecar/errors/mapper.go:10-62`
**Reason:** Code context differs from review — already fixed in earlier work. `mapper.go` uses `errors.As` with AWS SDK typed error types (`*dynamodbtypes.ResourceNotFoundException`, etc.) and `strings.Contains` from stdlib. The custom `contains()`/`findSubstr()` functions no longer exist.
**Original issue:** Substring matching on error message strings and a reimplementation of `strings.Contains`.

---

### WR-05: Missing Input Validation in `Service.Create` and `Service.Update`

**File:** `internal/task/usecase/task_service.go:41-55`, `58-75`
**Reason:** Code context differs from review — already fixed in prior commit `8879797`. `Create` guards with `strings.TrimSpace(title) == ""` and `Update` guards with `id == ""`, both returning descriptive sentinel errors.
**Original issue:** Blank titles and empty IDs were silently passed through to the repository.

---

### WR-06: `go.mod` Declares `go 1.25.0` — Unreleased Version

**File:** `go.mod:3`
**Reason:** Fix attempted and rolled back. The directive was changed to `go 1.23.0` and committed (`88cc339`), but `go mod tidy` reverted it to `1.25.0` because the pinned transitive dependencies (`google.golang.org/grpc v1.81.0`, `golang.org/x/net v0.51.0`, `golang.org/x/sys v0.42.0`, `go.opentelemetry.io/otel*`, `google.golang.org/genproto/googleapis/rpc`) all declare `GoVersion: 1.25`. The Go toolchain enforces that a module's `go` directive must be at least as high as any dependency's declared minimum. Downgrading to 1.23 is not achievable with the current pinned dependency versions without also downgrading those dependencies. The commit was reverted (`1594a2a`). To resolve this finding, the dependency versions must be downgraded to releases that declare `go 1.23` or the finding must be accepted as a toolchain constraint of the chosen library versions.
**Original issue:** Specifying an unreleased Go version risks CI toolchain incompatibility.

---

_Fixed: 2026-05-07T08:13:00Z_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_
