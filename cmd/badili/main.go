// Copyright 2026 Tomasz Sobczak <ictslabs@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	exporters "github.com/tomsobpl/badili/internal/exporters/otlp"
	listeners "github.com/tomsobpl/badili/internal/listeners/gelf"
	"github.com/tomsobpl/badili/internal/logging"
	"github.com/tomsobpl/badili/internal/platform/telemetry"
)

func init() {
	logging.SetupLogger()
	telemetry.InitTracer()
}

func main() {
	// setup signal handling
	ctx, cancel := context.WithCancel(context.Background())

	// main telemetry span
	//ctx, span := telemetry.Tracer().Start(ctx, "Main")
	//defer span.End()

	osSignalChan := make(chan os.Signal, 1)
	signal.Notify(osSignalChan, syscall.SIGINT, syscall.SIGTERM)

	var listenersWaitGroup sync.WaitGroup
	var exportersWaitGroup sync.WaitGroup

	// start listener components
	for i := 0; i < 1; i++ {
		listenersWaitGroup.Add(1)
		go listeners.StartUdpListenerSupervisor(ctx, 12201, &listenersWaitGroup)
	}

	// start exporter components
	for i := 0; i < 1; i++ {
		exportersWaitGroup.Add(1)
		go exporters.StartExporterSupervisor(ctx, 50051, &exportersWaitGroup)
	}

	// wait for OS interrupts
	<-osSignalChan
	slog.InfoContext(ctx, "shutdown signal received, exiting")

	// trigger shutdown sequence
	cancel()                  // start shutdown procedure
	listenersWaitGroup.Wait() // wait for listeners shutdown
	exportersWaitGroup.Wait() // wait for exporters shutdown

	slog.InfoContext(ctx, "shutdown complete")
}
