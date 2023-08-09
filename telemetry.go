package main

import (
	"context"
	"log"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	gcppropagator "github.com/GoogleCloudPlatform/opentelemetry-operations-go/propagator"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

func tpExporter(projectID, serviceName string) *sdktrace.TracerProvider {
	exporter, err := texporter.New(texporter.WithProjectID(projectID))
	if err != nil {
		log.Fatalf("texporter.New: %v", err)
		return nil
	}

	ctx := context.Background()
	res, err := resource.New(ctx,
		resource.WithDetectors(gcp.NewDetector()),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		log.Fatalf("resource.New: %v", err)
		return nil
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	func() {
		otel.SetTextMapPropagator(
			propagation.NewCompositeTextMapPropagator(
				gcppropagator.CloudTraceOneWayPropagator{},
				propagation.TraceContext{},
				propagation.Baggage{},
			))
	}()
	return tp
}
