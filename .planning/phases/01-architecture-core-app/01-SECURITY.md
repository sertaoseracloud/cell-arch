# SECURITY.md – Phase 1: Architecture & Core App

**Generated:** 2026-05-06
**Phase:** 01-architecture-core-app

## Scope

This document records the security threat analysis and mitigations for Phase 1, which establishes the foundational Go application using Clean Architecture principles. The goal is to ensure the core binary contains **zero cloud SDK imports** and follows secure coding practices.

## Threat Model Summary

| Threat ID | Category (STRIDE) | Component | Description | Mitigation (implemented) |
|-----------|-------------------|-----------|-------------|--------------------------|
| T-01-01   | **Spoofing**      | CLI / Cobra command | Unvalidated command‑line flags could be used to inject malicious values. | All Cobra flags are defined with `ValidArgs` and `MarkFlagRequired` where needed. Input is validated using `cobra.CheckErr` and custom validators where appropriate. |
| T-01-02   | **Tampering**     | Source code imports | Accidental inclusion of AWS/Azure SDK packages would break the zero‑SDK policy. | `go.mod` is audited with `go list -m all` and CI lint rule `no_external_sdk` (custom `golangci‑lint` rule) that flags any import matching `github.com/aws/*` or `github.com/Azure/*`. |
| T-01-03   | **Repudiation**   | Logging | Logs could be manipulated to hide malicious activity. | Structured logging is performed with **zerolog**; every log entry includes a `trace_id` (propagated via `context.Context`). Log output is immutable JSON written to stdout/stderr (no file writes). |
| T-01-04   | **Information Disclosure** | Error messages | Detailed errors may leak internal implementation details. | Errors are wrapped with `%w` and returned upstream without exposing stack traces. User‑facing errors are sanitized in `cmd/app/main.go` before logging. |
| T-01-05   | **Denial of Service** | Graceful shutdown | Failure to handle OS signals could leave goroutines running, consuming resources. | OS signal handling implemented with `context.WithCancel` (Decision D‑04). All long‑running goroutines listen to the root context; tests verify graceful termination within 2 seconds. |
| T-01-06   | **Elevation of Privilege** | Dependency injection | Use of a reflection‑based DI container could bypass compile‑time checks. | Manual constructors only (Decision D‑02); no global state or `init()` functions. |
| T-01-07   | **Tampering / Race Conditions** | Concurrency | Improper context propagation may cause race conditions on shared resources. | All public functions accept `context.Context` as the first argument (Decision D‑03). `go test -race ./...` is part of CI and passes with zero data races. |
| T-01-08   | **Integrity** | Configuration | Untrusted configuration files could alter runtime behavior. | Configuration is loaded via **viper** in a read‑only manner; environment variables are preferred. Load failures abort start‑up. |

## Mitigations Detail

### 1. Zero‑SDK Policy (T‑01‑02)

- `go.mod` does not list any `aws-*` or `azure-*` modules.
- CI includes a custom `golangci‑lint` rule (`no_external_sdk`) that scans source for prohibited import paths.
- The planner explicitly verifies **no cloud SDK imports** in any file before committing.

### 2. Manual Dependency Injection (T‑01‑06)

- All components are instantiated via explicit `NewX` functions in `cmd/app/main.go`.
- No use of reflection‑based containers (e.g., Uber/fx, dig) – enforced by the lint rule `no_reflection_di`.

### 3. Context‑First Boundaries (T‑01‑03, T‑01‑07)

- Public functions in the **usecase** and **infrastructure** layers have signatures `func (s *Service) Method(ctx context.Context, ...)`.
- A helper `LoadConfig(ctx context.Context)` exists even though viper does not use the context; the signature is kept for consistency.
