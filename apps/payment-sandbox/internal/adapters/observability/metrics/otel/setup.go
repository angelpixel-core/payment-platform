package otelmetrics

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	stdoutmetric "go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/sdk/metric"
)

func Setup(ctx context.Context) (func(context.Context) error, error) {
	exporter, err := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
	if err != nil {
		return nil, err
	}
	provider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter, metric.WithInterval(30*time.Second))),
	)
	otel.SetMeterProvider(provider)
	return provider.Shutdown, nil
}
