# Phase 4 – Observability – Context & Decisions

## Goal
Add OpenTelemetry‑based tracing, metrics, and structured logging to both the main app and the sidecar, with full visibility via **Jaeger UI** (traces) and **Grafana** (metrics/dashboards), following **Well‑Architected**, **FinOps**, and **GreenOps** best practices.

## Locked Decisions (must be respected by downstream agents)

| Decision | Rationale | Result |
|---|---|---|
| **Trace Visualization** | Use **Jaeger UI** as the sole trace visualization tool. Deploy Jaeger all‑in‑one (or with Elasticsearch backend) and expose its UI. | Traces are sent from OTel Collector to Jaeger via `jaegerexporter`. Jaeger UI accessible at `http://jaeger-query:16686`. |
| **Metrics Visualization** | Use **Grafana** (connected to Prometheus) for metrics dashboards, cost analysis, and GreenOps monitoring. | Grafana dashboards show request rates, latency, error ratios, and resource‑usage metrics per cloud (`cloud.provider` label). |
| **Trace Sampling Strategy** | Adopt **probabilistic sampling at 10 %** for production traffic, with **always‑sampled** traces for error paths (status ≥ 5xx). | `sampler: type=probabilistic, param=0.10` + `tail_sampler` for error spans. |
| **gRPC Trace Propagation** | Use the OpenTelemetry **Go gRPC interceptor** (`otelgrpc.UnaryServerInterceptor` / `otelgrpc.StreamServerInterceptor`) on both the app and sidecar servers. | No manual injection required; trace IDs flow through the gRPC boundary. |
| **Collector Configuration** | Deploy a **single OpenTelemetry Collector DaemonSet** per node that exports to **Jaeger** (traces) and **Prometheus** (metrics). The collector config includes a `batch` processor and duplicate pipelines for traces and metrics. | `collector-config.yaml` defines `jaeger` and `prometheus` exporters, with a pipeline that sends traces to Jaeger and metrics to Prometheus. |
| **Performance Benchmark Approach** | Adopt the **`ghz`** load‑testing tool to generate realistic traffic against the sidecar. Success criteria: *p95 latency ≤ 5 ms*, *error rate ≤ 0.1 %*, *throughput ≥ 1 k RPS* for `GetItem` RPC. | Benchmarks are reproducible and stored as CI artefacts. |
| **Observability Sidecar** | Introduce a **dedicated Observability Sidecar** (`cmd/obs-sidecar`) that runs alongside the existing sidecar. It hosts the OpenTelemetry Collector DaemonSet, configuration reload endpoint, and health‑check API. | New binary `obs-sidecar` built with same pipeline; logs and metrics are emitted to the collector locally. |
| **Best‑Practice Guidance** | Align with **Well‑Architected Framework**, **FinOps**, and **GreenOps** principles: <br>• Export trace, metric, and log metadata (`cloud.provider`, `environment`, `project`) for cost‑analysis. <br>• Push aggregated metrics to **Grafana** for dashboards; enable **Jaeger UI** for trace visualisation. <br>• Configure collector to **sample down** before forwarding to cost‑incurring back‑ends. <br>• Add **resource labels** (`cloud.account.id`, `cloud.region`) to allow downstream cost attribution per cloud. | Documentation added to `observability/README.md` with dashboard templates and cost‑monitoring guidelines. |
| **Cross‑Cloud Sync of Observability Data** | Ensure that both AWS and Azure receive the **same set of telemetry** by having the collector duplicate each export. Metadata includes `cloud.account.id` and `cloud.region` to allow downstream cost attribution per cloud. | Downstream dashboards can filter by `cloud.provider` to compare cloud‑specific performance and cost. |
| **Grafana Dashboards** | Provide pre‑built Grafana dashboards for: <br>• **Well‑Architected** – operational excellence, reliability, performance efficiency. <br>• **FinOps** – cost per cloud, cost per request, cost trends. <br>• **GreenOps** – energy consumption estimates, carbon footprint per request, resource utilisation. | Dashboard JSON files stored in `observability/grafana/dashboards/`. |
| **Jaeger UI Configuration** | Deploy Jaeger with: <br>• **All‑in‑one** for dev/test. <br>• **Elasticsearch backend** for production (persistent storage). <br>• **Spark dependencies** for service‑graph visualisation. | Jaeger available at `http://jaeger-query:16686`; service graph at `http://jaeger-query:16686/dependencies`. |

## Open Questions (deferred)

- **Retention policies** for trace data in Jaeger (to be decided in Phase 5).  
- **Alerting thresholds** for latency spikes (to be defined after baseline benchmarks).  

## Next Steps for Downstream Agents

1. **gsd‑phase‑researcher** – Investigate:
   - Configuration examples for OTel Collector with Jaeger + Prometheus exporters.
   - Grafana dashboard JSON templates for Well‑Architected, FinOps, GreenOps.
   - `ghz` benchmark scripts targeting the sidecar RPCs.
2. **gsd‑planner** – Produce concrete plan items:
   - Implement `cmd/obs-sidecar` with collector binary and config.
   - Add OpenTelemetry interceptors to existing gRPC servers.
   - Write collector `ConfigMap` and deployment manifests (DaemonSet).
   - Define CI job that runs `ghz` benchmarks and fails on regressions.
   - Create Grafana dashboard JSON files for the three focus areas.
   - Deploy Jaeger all‑in‑one (or Elasticsearch backend) with service‑graph support.
3. **gsd‑security‑auditor** – Verify that trace data does not leak sensitive payloads (PII, credentials).  

These decisions are now locked; downstream agents should treat them as immutable constraints unless the user explicitly revises them.