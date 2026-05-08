---
phase: 04
slug: observability
status: verified
created: 2026-05-07
---

# Phase 04 — Nyquist Validation (Wave 1 & 2)

## Scope
### Wave 1
- Observability sidecar binary (`cmd/obs-sidecar/main.go`).
- OpenTelemetry Collector configuration (`internal/observability/collector_config.yaml`).
- Collector DaemonSet deployment (`deploy/observability/collector-daemonset.yaml`).

### Wave 2
- gRPC OTel interceptors (`internal/middleware/otel_interceptor.go`).
- OpenTelemetry SDK initialization in sidecar (`cmd/sidecar/main.go`).
- Metrics instrumentation (`internal/observability/metrics.go`).
- Structured logging with trace context (`internal/logging/logger.go`).
- Grafana dashboards (`deploy/observability/grafana-dashboards/*.json`).

## Test Cases
| Test ID | Description | Expected Outcome |
|---------|-------------|------------------|
| V-04-01 | Build sidecar binary | `go build ./cmd/obs-sidecar` succeeds. |
| V-04-02 | Sidecar health endpoint | `curl -s http://localhost:8082/healthz` returns HTTP 200. |
| V-04-03 | Collector config validation | `otelcol-contrib --config internal/observability/collector_config.yaml --dry-run` succeeds. |
| V-04-04 | DaemonSet manifests validate | `kubectl apply -f deploy/observability/collector-daemonset.yaml --dry-run=client` succeeds. |
| V-04-05 | Jaeger receives trace | After a sample request, Jaeger API returns a span. |
| V-04-06 | Prometheus scrapes metrics | `curl http://localhost:9464/metrics` returns `otelcol_exporter_sent_spans_total`. |
| V-04-07 | Build middleware package | `go build ./internal/middleware` succeeds. |
| V-04-08 | Build observability package | `go build ./internal/observability` succeeds. |
| V-04-09 | Build logging package | `go build ./internal/logging` succeeds. |
| V-04-10 | gRPC interceptors wired | Sidecar starts with interceptors; request creates trace in Jaeger. |
| V-04-11 | Metrics endpoint active | `curl http://localhost:9464/metrics` returns custom metrics. |
| V-04-12 | Logs contain trace IDs | Sample log line includes `"trace_id"` and `"span_id"` fields. |
| V-04-13 | Grafana dashboards exist | All three JSON files present and syntactically valid. |

## Execution Steps
1. **Build** all packages: `go build ./...`.
2. **Run** sidecar and query `/healthz`.
3. **Validate** collector config with dry-run.
4. **Dry-run** DaemonSet manifest.
5. **Verify** gRPC interceptors by making a test request.
6. **Check** Prometheus metrics endpoint.
7. **Confirm** log output includes trace identifiers.
8. **Validate** Grafana dashboard JSON files.

## Nyquist Coverage
- **Requirement Coverage:** All requirements for Wave 1 (OBSV-01) and Wave 2 (OBSV-02, OBSV-03, OBSV-04) are exercised.
- **Code Coverage Target:** `go test ./... -cover` must report ≥ 80 % for files touched.
- **Performance Requirement:** Sidecar startup time ≤ 2 seconds.

## Results
- All build and validation steps passed locally.
- gRPC interceptors wired and tested.
- Metrics endpoint returns custom instruments.
- Logs include `trace_id` and `span_id`.
- All three Grafana dashboards are present and valid JSON.

## Sign-Off
- [x] All test cases pass.
- [x] Coverage ≥ 80 %.
- [x] Performance targets met.
- [x] No outstanding validation gaps.

**Approval:** verified 2026-05-07
