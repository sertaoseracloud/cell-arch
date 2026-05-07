# Technical Spec: Instrumentação OpenTelemetry em Go

## 1. Tracing

* **Otel-Go SDK**: Configuração do `sdktrace.NewTracerProvider`.
* **Middlewares**: Uso de `otelhttp` ou `otelgrpc` para instrumentar automaticamente todas as requisições de saída do App para o Sidecar.

## 2. Metrics (Prometheus/OTLP)

* **Custom Metrics**: Definição de `counters` e `histograms` utilizando a biblioteca `go.opentelemetry.io/otel/metric`.
* **Runtime Metrics**: Exportação de métricas nativas do Go (GC, memstats, goroutines) para monitorar o health dos containers no AKS/EKS.

## 3. Structured Logging

* **Slog (Go 1.21+)**: Uso da biblioteca nativa `log/slog` para logs estruturados em JSON.
* **Trace-Log Correlation**: Injeção automática de `trace_id` nos atributos de log via `slog.Handler` customizado.
