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

	exporter "github.com/tomsobpl/badili/internal/exporter/otlp"
	listener "github.com/tomsobpl/badili/internal/listener/gelf"
	"github.com/tomsobpl/badili/internal/logging"
	"github.com/tomsobpl/badili/internal/telemetry"
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

	var listenerWaitGroup sync.WaitGroup
	var exporterWaitGroup sync.WaitGroup

	// start listener component
	listenerWaitGroup.Add(1)
	go listener.StartUdpListenerSupervisor(ctx, 12201, &listenerWaitGroup)

	// start exporter component
	exporterWaitGroup.Add(1)
	go exporter.StartExporterSupervisor(ctx, 50051, &exporterWaitGroup)

	// wait for OS interrupts
	<-osSignalChan
	slog.InfoContext(ctx, "shutdown signal received, exiting")

	// trigger shutdown sequence
	cancel()                 // start shutdown procedure
	listenerWaitGroup.Wait() // wait for listener shutdown
	exporterWaitGroup.Wait() // wait for exporter shutdown

	slog.InfoContext(ctx, "shutdown complete")
}
