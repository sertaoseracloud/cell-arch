package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var meter = otel.GetMeterProvider().Meter("cell-arch")

var (
	requestLatency metric.Float64Histogram
	sidecarReadRPS metric.Int64Counter
)

func init() {
	var err error
	requestLatency, err = meter.Float64Histogram(
		"request_latency_seconds",
		metric.WithDescription("Latency of requests in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		panic(err)
	}
	sidecarReadRPS, err = meter.Int64Counter(
		"sidecar_read_rps",
		metric.WithDescription("Read throughput in requests per second"),
	)
	if err != nil {
		panic(err)
	}
}

// InitMetrics initialises the Prometheus exporter and registers custom metrics.
func InitMetrics() {
	// The Prometheus exporter is configured in the collector config.
	// Metrics are exposed via the OTel Prometheus receiver on port 8888.
}

// RecordLatency records a request latency with labels.
func RecordLatency(ctx context.Context, cloud, service, statusCode string, latency float64) {
	requestLatency.Record(ctx, latency,
		metric.WithAttributes(
			attribute.String("cloud_provider", cloud),
			attribute.String("service", service),
			attribute.String("status_code", statusCode),
		),
	)
}

// IncReadRPS increments the read throughput counter.
func IncReadRPS(ctx context.Context, cloud, service string) {
	sidecarReadRPS.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("cloud_provider", cloud),
			attribute.String("service", service),
		),
	)
}
