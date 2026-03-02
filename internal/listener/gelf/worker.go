// Copyright 2026 Tomasz Sobczak <ictslabs@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package gelf

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/tomsobpl/badili/internal/grpcapi"
	"github.com/tomsobpl/badili/internal/telemetry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

func StartWorkerSupervisor(ctx context.Context, id int, packetChan <-chan Packet, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		if err := runWorker(ctx, id, packetChan); err != nil {
			slog.WarnContext(ctx, "Worker failed. Restarting in 5s ...", "id", id, "err", err)
			time.Sleep(5 * time.Second)
			continue
		}

		break
	}
}

func runWorker(ctx context.Context, id int, packets <-chan Packet) (err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.ErrorContext(ctx, "Recovered UdpListenerWorker from panic", "worker_id", id, "recover", r)
			err = fmt.Errorf("some error")
		}
	}()

	l := slog.With("worker_id", id)
	l.InfoContext(ctx, "exporter worker started")

	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("127.0.0.1:%d", 50051))
	if err != nil {
		slog.ErrorContext(ctx, "local grpcapi exporter ResolveTcpAddr error", "err", err)
		return err
	}

	var cp = keepalive.ClientParameters{
		Time:                10 * time.Second, // Send a ping every 10s
		Timeout:             time.Second,      // Wait 1s for response
		PermitWithoutStream: true,             // Send pings even without active RPCs
	}

	conn, _err := grpc.NewClient(addr.String(),
		//conn, _err := grpc.NewClient("unix:///tmp/badili.exporter.sock",
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(1024*1024*8), grpc.MaxCallRecvMsgSize(1024*1024*8)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(cp),
	)

	if _err != nil {
		l.ErrorContext(ctx, "grpc client creation failed", "err", _err)
	}

	defer conn.Close()

	//// 2. Start Connection Monitor
	//go func() {
	//	for {
	//		if conn.WaitForStateChange(ctx, conn.GetState()) {
	//			log.Printf("Connection state changed to: %s", conn.GetState())
	//		} else {
	//			return
	//		}
	//	}
	//}()

	// initialize and start worker
	worker := grpcapi.NewMessageWorker(conn, 1000, id)
	go worker.Start(ctx)

	for p := range packets {
		c, span := telemetry.Tracer().Start(ctx, "UdpListenerWorker")

		if p.IsChunk() {
			l.InfoContext(c, "Received chunked GELF message", "from", p.Addr)
			continue
		} else {
			// Simulate processing (e.g., database write or complex parsing)
			l.InfoContext(c, "Processing GELF message",
				"from", p.Addr,
				"len", len(p.Data),
			)
			msg, err := DecodePacketToProtoMessage(p)

			if err != nil {
				l.WarnContext(c, "error decoding GELF message", "from", p.Addr, "err", err)
			}

			//l.InfoContext(c, "decoded GELF message", "from", p.Addr, "msg", msg)

			worker.Submit(msg)
		}

		span.End()
	}

	l.InfoContext(ctx, "exporter worker shutdown complete")
	return nil
}
