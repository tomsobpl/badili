// Copyright 2026 Tomasz Sobczak <ictslabs@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package otlp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/tomsobpl/badili/api/gelfapi/v1"
	"github.com/tomsobpl/badili/internal/grpcapi"
	"github.com/tomsobpl/badili/internal/platform/telemetry"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
)

func StartExporterSupervisor(ctx context.Context, port int, wg *sync.WaitGroup) {
	defer wg.Done()

	messagesChan := make(chan *gelfapi.Message, 1024)

	loggerProvider, _ := telemetry.CreateRawOtlpProvider(ctx)

	defer func(ctx context.Context, loggerProvider *sdklog.LoggerProvider) {
		slog.InfoContext(ctx, "shutting down logger provider")
		if err := loggerProvider.Shutdown(context.Background()); err != nil {
			slog.ErrorContext(ctx, "failed to shutdown logger provider", "err", err)
		}
	}(ctx, loggerProvider)

	//// Get a logger. The name/version here are for the "Scope",
	//// but if you want it truly raw, keep these strings empty.
	//logger := lp.Logger("grpcapi-transformer")
	//// Create your record (as we did in the previous step)
	//record := log.Record{}
	//record.SetBody(log.StringValue("Your GELF Short Message"))
	//record.AddAttributes(
	//	log.String("host", "my-server-01"),
	//	log.Int("level", 3),
	//)
	//
	//// Emit!
	//// Because we used resource.Empty(), the final JSON/Export
	//// will ONLY contain the Body and these two Attributes.
	//logger.Emit(record)

	var workersWaitGroup sync.WaitGroup

	// start exporter workers
	for i := 0; i < 5; i++ {
		workersWaitGroup.Add(1)
		go StartWorkerSupervisor(ctx, i, messagesChan, loggerProvider, &workersWaitGroup)
	}

	for {
		slog.InfoContext(ctx, "starting exporter", "port", port)

		err := runExporter(ctx, port, messagesChan)

		if err == nil {
			break
		}

		slog.WarnContext(ctx, "exporter failed, restarting in 5s ...", "port", port, "err", err)
		time.Sleep(5 * time.Second)
	}

	close(messagesChan)     // tell workers no more data is coming
	workersWaitGroup.Wait() // wait for workers to finish the queue
}

func runExporter(ctx context.Context, port int, messages chan<- *gelfapi.Message) (err error) {
	defer func(ctx context.Context, port int, err error) {
		if r := recover(); r != nil {
			slog.ErrorContext(ctx, "recovered otlp exporter from panic", "port", port, "recover", r)
			err = errors.New(r.(string))
		}
	}(ctx, port, err)

	// Resolve local TCP address
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		slog.ErrorContext(ctx, "local grpcapi exporter ResolveTcpAddr error", "err", err)
		return err
	}

	// Create TCP listener
	listener, err := net.Listen("tcp", addr.String())
	//listener, err := net.Listen("unix", "/tmp/badili.exporter.sock")
	if err != nil {
		slog.ErrorContext(ctx, "local grpcapi exporter Listen error", "err", err)
		return err
	}

	//defer os.Remove("/tmp/badili.exporter.sock")

	// Ensure TCP listener is closed on exit
	//defer func(listener net.Listener) {
	//	err := listener.Close()
	//	if err != nil {
	//		slog.ErrorContext(ctx, "local grpcapi exporter Close error", "err", err)
	//	}
	//}(listener)

	slog.InfoContext(ctx, "local grpcapi exporter started", "port", port)

	// Initialize gRPC server
	server := grpc.NewServer(
		grpc.MaxRecvMsgSize(1024*1024*8), // 8MB limit
		grpc.MaxSendMsgSize(1024*1024*8), // 8MB limit
	)

	// Register exporter grpcapi implementation
	gelfapi.RegisterMessageServiceServer(server, &grpcapi.MessageServerImplementation{MessagesChan: messages})

	go func(ctx context.Context, server *grpc.Server, listener net.Listener, err error) {
		slog.InfoContext(ctx, "starting listening for grpcapi streams", "port", port)
		if e := server.Serve(listener); err != nil {
			slog.ErrorContext(ctx, "local grpc exporter Serve error", "err", e)
			err = e
		}
	}(ctx, server, listener, err)

	cancel := <-ctx.Done()
	slog.InfoContext(ctx, "received interrupt, shutting down", "cancel", cancel, "port", port)

	slog.InfoContext(ctx, "shutting down grpcapi exporter", "port", port)
	server.GracefulStop()
	slog.InfoContext(ctx, "local grpcapi exporter stopped", "port", port)

	return nil
}
