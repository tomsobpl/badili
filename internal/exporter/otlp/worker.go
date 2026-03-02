// Copyright 2026 Tomasz Sobczak <ictslabs@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package otlp

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/tomsobpl/badili/api/gelfapi/v1"
	"github.com/tomsobpl/badili/internal/platform/telemetry"
	otellog "go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
	"google.golang.org/protobuf/types/known/structpb"
)

func StartWorkerSupervisor(ctx context.Context, id int, messages <-chan *gelfapi.Message, loggerProvider *sdklog.LoggerProvider, wg *sync.WaitGroup) {
	defer wg.Done()

	logger := loggerProvider.Logger("otlp-exporter-worker")

	for {
		if err := runWorker(ctx, id, messages, logger); err != nil {
			slog.WarnContext(ctx, "Worker failed. Restarting in 5s ...", "id", id, "err", err)
			time.Sleep(5 * time.Second)
			continue
		}

		break
	}
}

func runWorker(ctx context.Context, id int, messages <-chan *gelfapi.Message, logger otellog.Logger) (err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.ErrorContext(ctx, "Recovered worker from panic", "worker_id", id, "recover", r)
			err = fmt.Errorf("some error")
		}
	}()

	l := slog.With("worker_id", id)
	l.InfoContext(ctx, "OtlpExporterWorker started")

	for m := range messages {
		_, span := telemetry.Tracer().Start(ctx, "OtlpExporterWorker")

		seconds := int64(m.Timestamp)
		nanoseconds := int64((m.Timestamp - float64(seconds)) * 1e9)

		record := otellog.Record{}
		record.SetBody(otellog.StringValue(m.ShortMessage))
		record.SetSeverity(telemetry.OtelSeverityFromGelfLevel(m.Level))
		record.SetSeverityText(record.Severity().String())
		record.SetTimestamp(time.Unix(seconds, nanoseconds))

		attrs := []otellog.KeyValue{
			otellog.String("gelf.version", m.Version),
			otellog.String(string(semconv.HostNameKey), m.Host),
		}

		if m.FullMessage != "" {
			attrs = append(attrs, otellog.String(string(semconv.ExceptionMessageKey), m.FullMessage))
		}

		if m.Extras != nil {
			for key, value := range m.Extras.Fields {
				attrs = append(attrs, otellog.KeyValue{
					Key:   fmt.Sprintf("app.extra.%s", strings.TrimPrefix(key, "_")),
					Value: protoValueToOtelValue(value),
				})
			}
		}

		record.AddAttributes(attrs...)
		logger.Emit(ctx, record)
		span.End()
	}

	l.InfoContext(ctx, "UdpListenerWorker shutdown complete")
	return nil
}

func protoValueToOtelValue(v *structpb.Value) otellog.Value {
	switch x := v.Kind.(type) {
	case *structpb.Value_StringValue:
		return otellog.StringValue(x.StringValue)
	case *structpb.Value_NumberValue:
		return otellog.Float64Value(x.NumberValue)
	case *structpb.Value_BoolValue:
		return otellog.BoolValue(x.BoolValue)
	case *structpb.Value_ListValue:
		var vals []otellog.Value
		for _, item := range x.ListValue.Values {
			vals = append(vals, protoValueToOtelValue(item))
		}
		return otellog.SliceValue(vals...)
	case *structpb.Value_StructValue:
		var kvs []otellog.KeyValue
		for k, v := range x.StructValue.Fields {
			kvs = append(kvs, otellog.KeyValue{Key: k, Value: protoValueToOtelValue(v)})
		}
		return otellog.MapValue(kvs...)
	default:
		return otellog.Empty("_").Value
	}
}
