// Copyright 2026 Tomasz Sobczak <ictslabs@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"context"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func InitTracer() {
	otel.SetTracerProvider(sdktrace.NewTracerProvider())
}

func Tracer() trace.Tracer {
	return otel.Tracer("badili")
}

func CreateRawOtlpProvider(ctx context.Context) (*sdklog.LoggerProvider, error) {
	//exporter, err := stdoutlog.New()
	exporter, err := otlploggrpc.New(
		ctx,
		otlploggrpc.WithEndpoint("uptrace:4317"),
		otlploggrpc.WithInsecure(),
		otlploggrpc.WithHeaders(map[string]string{
			"uptrace-dsn": "http://badili_secret@localhost:14318?grpc=14317",
		}),
	)

	if err != nil {
		slog.ErrorContext(ctx, "failed to create exporter", "err", err)
		return nil, err
	}

	processor := sdklog.NewBatchProcessor(
		exporter,
		sdklog.WithMaxQueueSize(2048),
		sdklog.WithExportMaxBatchSize(512),
		sdklog.WithExportTimeout(2*time.Second),
	)

	return sdklog.NewLoggerProvider(
		sdklog.WithProcessor(processor),
		sdklog.WithResource(resource.Empty()),
	), nil
}

//// Simplified initialization for the example
//func InitTracerProvider() *sdktrace.TracerProvider {
//	exporter, _ := stdouttrace.New(stdouttrace.WithPrettyPrint())
//	tp := sdktrace.NewTracerProvider(
//		sdktrace.WithBatcher(exporter),
//		sdktrace.WithResource(resource.NewWithAttributes(
//			semconv.SchemaURL,
//			semconv.ServiceNameKey.String("udp-worker-service"),
//		)),
//	)
//	otlp.SetTracerProvider(tp)
//	return tp
//}
