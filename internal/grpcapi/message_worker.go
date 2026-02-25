package grpcapi

import (
	"context"
	"log/slog"
	"time"

	"github.com/tomsobpl/badili/api/gelfapi/v1"
	"google.golang.org/grpc"
)

type MessageWorkerImplementation struct {
	client        gelfapi.MessageServiceClient
	batchSize     int
	flushInterval time.Duration
	messages      chan *gelfapi.Message
	workerId      int
}

func (w *MessageWorkerImplementation) Start(ctx context.Context) {
	l := slog.With("worker_id", w.workerId)
	stream, err := w.client.StreamMessageBatch(ctx, grpc.WaitForReady(true))
	if err != nil {
		l.ErrorContext(ctx, "failed to open stream: %v", err)
		return
	}

	ticker := time.NewTicker(w.flushInterval)
	defer ticker.Stop()

	var batch []*gelfapi.Message

	for {
		select {
		case <-ctx.Done():
			l.InfoContext(ctx, "message_worker context done")
			if len(batch) > 0 {
				w.sendMessageBatch(stream, batch)
			}

			stream.CloseAndRecv()
			return
		case msg := <-w.messages:
			batch = append(batch, msg)
			if len(batch) >= w.batchSize {
				l.InfoContext(ctx, "message_worker batch full, sending")
				w.sendMessageBatch(stream, batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				l.InfoContext(ctx, "message_worker sending batch from ticker")
				w.sendMessageBatch(stream, batch)
				batch = batch[:0]
			}
		}
	}
}

func (w *MessageWorkerImplementation) sendMessageBatch(stream gelfapi.MessageService_StreamMessageBatchClient, batch []*gelfapi.Message) {
	if err := stream.Send(&gelfapi.MessageBatch{Messages: batch}); err != nil {
		slog.Error("error sending message batch", "err", err, "worker_id", w.workerId)
	}
}

func (w *MessageWorkerImplementation) Submit(msg *gelfapi.Message) {
	w.messages <- msg
}

func NewMessageWorker(conn *grpc.ClientConn, bufferSize int, workerId int) *MessageWorkerImplementation {
	return &MessageWorkerImplementation{
		client:        gelfapi.NewMessageServiceClient(conn),
		batchSize:     100,
		flushInterval: 2000 * time.Millisecond,
		messages:      make(chan *gelfapi.Message, bufferSize),
		workerId:      workerId,
	}
}
