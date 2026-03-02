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

	"github.com/tomsobpl/badili/internal/config"
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

	osSignalChan := make(chan os.Signal, 1)
	signal.Notify(osSignalChan, syscall.SIGINT, syscall.SIGTERM)

	// init configuration
	_c, span := telemetry.Tracer().Start(ctx, "InitConfiguration")
	cfg, err := config.InitConfiguration(_c)

	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	slog.DebugContext(ctx, "configuration loaded", "cfg", cfg)
	span.End()

	var listenerWaitGroup sync.WaitGroup
	var exporterWaitGroup sync.WaitGroup

	// start listener component
	if cfg.Listener.Enabled {
		listenerWaitGroup.Add(1)
		go listener.StartUdpListenerSupervisor(ctx, cfg.Listener, &listenerWaitGroup)
	}

	// start exporter component
	if cfg.Exporter.Enabled {
		exporterWaitGroup.Add(1)
		go exporter.StartExporterSupervisor(ctx, cfg.Exporter.Port, &exporterWaitGroup)
	}

	// wait for OS interrupts
	<-osSignalChan
	slog.InfoContext(ctx, "shutdown signal received, exiting")

	// trigger shutdown sequence
	cancel()                 // start shutdown procedure
	listenerWaitGroup.Wait() // wait for listener shutdown
	exporterWaitGroup.Wait() // wait for exporter shutdown

	slog.InfoContext(ctx, "shutdown complete")
}
