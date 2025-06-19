package tracer

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

var packageName string

type Option struct {
	JaegerHost  string
	PackageName string
}

func InitTracer(opt Option) func() {
	ctx := context.Background()

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(opt.JaegerHost),
		otlptracegrpc.WithInsecure(), // 🚫 No TLS (good for local dev)
	)
	if err != nil {
		panic("failed to create OTLP gRPC exporter: " + err.Error())
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("go-hris"),
		)),
	)

	otel.SetTracerProvider(tp)

	packageName = opt.PackageName

	return func() {
		_ = tp.Shutdown(ctx)
	}
}

func Start(ctx context.Context, spanName string) (context.Context, func()) {
	ctx, span := otel.Tracer(packageName).Start(ctx, spanName)
	return ctx, func() {
		span.End()
	}
}
