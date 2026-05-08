---
phase: 04
slug: observability
status: verified
threats_open: 0
asvs_level: 1
created: 2026-05-07
---

# Phase 04 — Security (Wave 2)

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| App ↔ Sidecar gRPC | Trace context propagation | Trace IDs, span metadata |
| Sidecar → Collector | Telemetry export | Traces, metrics, logs |
| Collector → Jaeger UI | Trace visualization | Full trace data |
| Collector → Prometheus | Metrics scraping | Aggregated metrics |
| App/Sidecar → Structured logs | Log output | JSON logs with trace IDs |

---

## Threat Register

| Threat ID | Category | Component | Disposition | Mitigation | Status |
|-----------|----------|-----------|-------------|------------|--------|
| T-04-01 | Spoofing | gRPC interceptors | mitigate | Enforce mTLS and validate peer certificates | closed |
| T-04-02 | Tampering | Collector config | mitigate | ConfigMap with `immutable: true`; RBAC restricts edits | closed |
| T-04-03 | Information Disclosure | Jaeger exporter | mitigate | Jaeger with authentication; UI restricted to internal network | closed |
| T-04-04 | Repudiation | Structured logs | accept | Logs are immutable JSON; stored in write-once storage | closed |
| T-04-05 | DoS | Prometheus scrape | mitigate | Rate-limit via `scrape_interval`; collector resource limits | closed |
| T-04-06 | Elevation of Privilege | Collector process | mitigate | Run as non-root user; drop capabilities | closed |
| T-04-09 | Spoofing | gRPC metadata | mitigate | Interceptors enforce `TraceContext` propagation; mTLS blocks MITM | closed |
| T-04-10 | Information Disclosure | Logs | mitigate | Logger helper filters sensitive keys; no PII in trace/log data | closed |
| T-04-11 | Availability | Prometheus scrape | mitigate | Set scrape interval ≥ 15s; collector resource limits prevent overload | closed |

---

## Accepted Risks Log

- **T-04-04**: Immutable JSON logs accepted as sufficient for repudiation in this phase. Future phases may add signed log appenders.

---

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-05-07 | 9 | 9 | 0 | gsd-security-auditor (Wave 1) |
| 2026-05-07 | 3 | 3 | 0 | gsd-security-auditor (Wave 2) |

---

## Sign-Off

- [x] All threats have a disposition
- [x] Accepted risks documented
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set

**Approval:** verified 2026-05-07
