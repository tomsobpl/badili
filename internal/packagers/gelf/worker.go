package gelf

import (
	"context"
	"log/slog"
	"sync"

	"github.com/tomsobpl/badili/api/gelfapi/v1"
	"github.com/tomsobpl/badili/internal/platform/telemetry"
)

func PackagerWorker(ctx context.Context, id int, chunks <-chan *gelfapi.Chunk, wg *sync.WaitGroup) {
	defer wg.Done()

	l := slog.With("worker_id", id)
	l.InfoContext(ctx, "PackagerWorker started")

	for c := range chunks {
		localCtx, span := telemetry.Tracer().Start(ctx, "PackagerWorker")

		l.InfoContext(localCtx, "Processing GELF chunk",
			"messageId", c.MessageId,
			"sequenceNum", c.SequenceNum,
			"sequenceCount", c.SequenceCount,
			"timestamp", c.Timestamp,
		)

		span.End()
	}

	l.InfoContext(ctx, "PackagerWorker shutdown complete")
}
