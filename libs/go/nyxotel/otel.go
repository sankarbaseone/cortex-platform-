// Package nyxotel wires OpenTelemetry tracing+metrics and the platform's
// structured JSON logger (RFC-014 G.0): "Tracing: OTel; W3C propagation,
// spans on every RPC, consumer batch, DB call" and "Logs: structured JSON
// {ts,level,svc,tenant,trace_id,msg,fields}; payload contents NEVER logged".
// MustInit is the second call in every service's main.go (ADR-0001 wiring
// order: config -> otel -> stores -> ...).
package nyxotel

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Config is embedded (as OTel) in each service's own config struct and
// populated by libs/go/config from the environment.
type Config struct {
	ServiceName    string  `envconfig:"OTEL_SERVICE_NAME" required:"true"`
	ServiceVersion string  `envconfig:"OTEL_SERVICE_VERSION" default:"dev"`
	Endpoint       string  `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT" default:"localhost:4317"`
	Insecure       bool    `envconfig:"OTEL_EXPORTER_OTLP_INSECURE" default:"true"`
	SampleRatio    float64 `envconfig:"OTEL_TRACES_SAMPLER_RATIO" default:"1.0"`
}

// Shutdown flushes and stops the tracer/meter providers. Called via defer
// immediately after MustInit in main.go.
type Shutdown func(context.Context) error

// MustInit installs a global TracerProvider + MeterProvider exporting via
// OTLP/gRPC, sets the W3C trace-context propagator (RFC-014 G.0), and
// returns a Shutdown func. Panics on exporter/setup failure: a service
// cannot safely run unobserved.
func MustInit(cfg Config) Shutdown {
	ctx := context.Background()

	res, err := resource.Merge(resource.Default(), resource.NewSchemaless(
		semconv.ServiceName(cfg.ServiceName),
		semconv.ServiceVersion(cfg.ServiceVersion),
	))
	if err != nil {
		panic(fmt.Errorf("nyxotel: resource: %w", err))
	}

	traceOpts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(cfg.Endpoint)}
	metricOpts := []otlpmetricgrpc.Option{otlpmetricgrpc.WithEndpoint(cfg.Endpoint)}
	if cfg.Insecure {
		traceOpts = append(traceOpts, otlptracegrpc.WithInsecure())
		metricOpts = append(metricOpts, otlpmetricgrpc.WithInsecure())
	}

	traceExp, err := otlptracegrpc.New(ctx, traceOpts...)
	if err != nil {
		panic(fmt.Errorf("nyxotel: trace exporter: %w", err))
	}
	metricExp, err := otlpmetricgrpc.New(ctx, metricOpts...)
	if err != nil {
		panic(fmt.Errorf("nyxotel: metric exporter: %w", err))
	}

	ratio := cfg.SampleRatio
	if ratio <= 0 {
		ratio = 1.0
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio))),
	)
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp)),
		sdkmetric.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)
	otel.SetTextMapPropagator(propagation.TraceContext{}) // W3C traceparent (RFC-005 B.2 envelope.trace_id)

	return func(shutdownCtx context.Context) error {
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 5*time.Second)
		defer cancel()
		var errs []error
		if err := tp.Shutdown(shutdownCtx); err != nil {
			errs = append(errs, fmt.Errorf("tracer provider: %w", err))
		}
		if err := mp.Shutdown(shutdownCtx); err != nil {
			errs = append(errs, fmt.Errorf("meter provider: %w", err))
		}
		if len(errs) > 0 {
			return fmt.Errorf("nyxotel: shutdown: %v", errs)
		}
		return nil
	}
}
