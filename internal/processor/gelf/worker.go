// Copyright 2026 Tomasz Sobczak <ictslabs@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package gelf

import (
	"context"
	"log/slog"
	"sync"

	"github.com/tomsobpl/badili/api/gelfapi/v1"
	"github.com/tomsobpl/badili/internal/telemetry"
)

func ProcessorWorker(ctx context.Context, id int, chunks <-chan *gelfapi.Chunk, wg *sync.WaitGroup) {
	defer wg.Done()

	l := slog.With("worker_id", id)
	l.InfoContext(ctx, "ProcessorWorker started")

	for c := range chunks {
		localCtx, span := telemetry.Tracer().Start(ctx, "ProcessorWorker")

		l.InfoContext(localCtx, "Processing GELF chunk",
			"messageId", c.MessageId,
			"sequenceNum", c.SequenceNum,
			"sequenceCount", c.SequenceCount,
			"timestamp", c.Timestamp,
		)

		span.End()
	}

	l.InfoContext(ctx, "ProcessorWorker shutdown complete")
}
