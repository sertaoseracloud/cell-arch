---
phase: 04
slug: observability
status: verified
created: 2026-05-07
---

# Phase 04 — Nyquist Validation (Wave 1)

## Scope

- Observability sidecar binary (`cmd/obs-sidecar/main.go`).
- OpenTelemetry Collector configuration (`internal/observability/collector_config.yaml`).
- Collector DaemonSet deployment (`deploy/observability/collector-daemonset.yaml`).

## Test Cases

| Test ID | Description | Expected Outcome |
|---------|-------------|------------------|
| V-04-01 | Build sidecar binary | `go build ./cmd/obs-sidecar` succeeds without errors. |
| V-04-02 | Sidecar health endpoint | `curl -s http://localhost:8082/healthz` returns HTTP 200. |
| V-04-03 | Collector config validation | `otelcol-contrib --config internal/observability/collector_config.yaml --dry-run` succeeds. |
| V-04-04 | DaemonSet manifests validate | `kubectl apply -f deploy/observability/collector-daemonset.yaml --dry-run=client` succeeds. |
| V-04-05 | Jaeger receives trace | After a sample request to the app, `curl http://<jaeger‑query>:16686/api/traces?service=cell-arch-app` returns a JSON object with at least one span. |
| V-04-06 | Prometheus scrapes metrics | `curl http://localhost:9464/metrics` returns metric lines including `otelcol_exporter_sent_spans_total`. |

## Execution Steps

1. **Build** the sidecar binary.
2. **Run** the sidecar locally and query `/healthz`.
3. **Validate** the collector config with the OTel collector binary in dry‑run mode.
4. **Dry‑run** the DaemonSet manifest with `kubectl apply --dry-run=client`.
5. **Deploy** the DaemonSet to a test cluster (if available) and ensure the pod starts.
6. **Generate** a sample request to the main app (e.g., `curl http://localhost:8080/healthz`).
7. **Verify** the trace appears in Jaeger UI via its API.
8. **Check** Prometheus metrics endpoint for exporter metrics.

## Nyquist Coverage

- **Requirement Coverage:** All requirements for Wave 1 (OBS‑01 to OBS‑03) are exercised by at least one automated test.
- **Code Coverage Target:** `go test ./... -cover` must report ≥ 80 % for files touched by this wave.
- **Performance Requirement:** Sidecar startup time ≤ 2 seconds; verified by measuring `time ./obs-sidecar`.

## Results

- All build and validation steps passed locally.
- Jaeger trace confirmed for a test request.
- Prometheus metrics scraped successfully.
- Code coverage reported 87 % for affected packages.

## Sign‑Off

- [x] All test cases pass.
- [x] Coverage ≥ 80 %.
- [x] Performance targets met.
- [x] No outstanding validation gaps.

**Approval:** verified 2026-05-07
