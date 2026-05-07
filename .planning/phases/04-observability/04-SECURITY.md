---
phase: 04
slug: observability
status: verified
threats_open: 0
asvs_level: 1
created: 2026-05-07
---

# Phase 04 — Security

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| App → Collector | Telemetry from app to OTel Collector | Trace & metric data |
| Sidecar → Collector | Telemetry from sidecar to collector | Trace & metric data |
| Collector → Jaeger UI | Exported traces | Trace data |
| Collector → Prometheus | Metrics endpoint | Metric data |

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

---

## Accepted Risks Log

_No accepted risks._

---

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-05-07 | 6 | 6 | 0 | gsd-security-auditor |

---

## Sign-Off

- [x] All threats have a disposition
- [x] Accepted risks documented
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set

**Approval:** verified 2026-05-07
