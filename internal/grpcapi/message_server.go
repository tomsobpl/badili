// Copyright 2026 Tomasz Sobczak <ictslabs@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package grpcapi

import (
	"context"
	"io"
	"log/slog"

	"github.com/tomsobpl/badili/api/gelfapi/v1"
)

type MessageServerImplementation struct {
	MessagesChan chan<- *gelfapi.Message
	gelfapi.UnimplementedMessageServiceServer
}

func (s *MessageServerImplementation) StreamMessageBatch(stream gelfapi.MessageService_StreamMessageBatchServer) error {
	var totalMessagesReceived int64

	for {
		messageBatch, err := stream.Recv()

		slog.Info("message received", "messageBatch", messageBatch)

		if err == io.EOF {
			slog.Info("worker stream closed", "totalMessagesReceived", totalMessagesReceived)
			return stream.SendAndClose(&gelfapi.MessageBatchSummary{TotalMessagesReceived: totalMessagesReceived})
		}

		if err != nil {
			slog.Error("error receiving from stream", "err", err)
			slog.Error("context", "err", context.Cause(context.Background()))
			return err
		}

		for _, m := range messageBatch.Messages {
			slog.Info("received message", "message", m)
			s.MessagesChan <- m
		}

		totalMessagesReceived += int64(len(messageBatch.Messages))
	}

	return nil
}
