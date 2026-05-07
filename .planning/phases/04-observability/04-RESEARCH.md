# Phase4 – Observability – Research

## Standard Stack
- **OpenTelemetry SDK** (go.opentelemetry.io/otel) – core tracing, metrics, and logging APIs.
- **OpenTelemetry Collector** (otel/opentelemetry-collector-contrib) – runs as a DaemonSet, exports traces to Jaeger via the Jaeger exporter.
- **Jaeger Exporter** (`github.com/open-telemetry/opentelemetry-collector-contrib/exporter/jaegerexporter`) – sends traces to Jaeger backend.
- **Jaeger UI** – primary trace visualization tool; deploy Jaeger all‑in‑one (or Jaeger with Elasticsearch backend) for querying traces.
- **Prometheus Exporter** – optional, for on‑prem metrics scraping if needed.
- **Logging** – `log/slog` with `slog.JSONHandler` for structured logs; inject `trace_id` and `span_id` from the current context.

## Architecture Patterns
1. **Tracer Provider per Process** – initialize a singleton tracer in `cmd/app/main.go` and `cmd/sidecar/main.go` using the same service name (`cell-arch-app` / `cell-arch-sidecar`).
2. **gRPC Interceptors** – use `otelgrpc.UnaryServerInterceptor` and `otelgrpc.StreamServerInterceptor` on both servers to auto‑propagate trace context.
3. **Collector DaemonSet** – one pod per node, configured with a single pipeline that exports traces to Jaeger via the Jaeger exporter, and optionally metrics to Prometheus.
4. **Sidecar‑Specific Collector** – a lightweight collector runs alongside the existing sidecar binary (same pod) to avoid extra network hops.
5. **Performance Benchmark** – use `ghz` to generate gRPC load against sidecar RPCs; instrument both client and server with OTel to capture latency metrics.

## Don't Hand‑Roll
- **Tracing implementation** – never implement custom HTTP headers for propagation; always rely on the OpenTelemetry gRPC interceptor.
- **Metric exporters** – if metrics are needed, use the official Prometheus exporter; do not write bespoke exporters.
- **Log correlation** – do not manually concatenate trace IDs; use `slog` fields populated from the OTel context.

## Common Pitfalls
- **Jaeger exporter misconfiguration** – forgetting to set `endpoint` (e.g., `jaeger-collector:14268`) leads to dropped traces.
- **Sampling mismatch** – applying different samplers on client and server can cause incomplete traces.
- **Collector resource limits** – insufficient CPU/memory causes back‑pressure and lost telemetry.
- **Duplicate spans** – creating new spans inside library code that already creates spans results in noisy traces.

## Code Examples
```go
// Initialize tracer (app and sidecar) – export to Jaeger
func initTracer(service string) (*sdktrace.TracerProvider, error) {
    exporter, err := jaeger.NewExporter(
        jaeger.WithCollectorEndpoint("http://jaeger-collector:14268/api/traces"),
    )
    if err != nil { return nil, err }
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithSampler(sdktrace.TraceIDRatioBased(0.10)), // 10% probabilistic
        sdktrace.WithSyncer(exporter),
    )
    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(propagation.TraceContext{})
    return tp, nil
}

// gRPC server with interceptor
grpcServer := grpc.NewServer(
    grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
    grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
)
```

## Sources
- OpenTelemetry Go SDK documentation (2026‑05‑07)
- Jaeger exporter README (2026‑05‑06)
- Jaeger UI deployment guide (2026‑05‑05)
- `ghz` load‑testing tool docs (2026‑05‑04)

---
*Research complete. This file will be consumed by `/gsd-plan-phase 4`.*