// Copyright 2026 Tomasz Sobczak <ictslabs@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package config

const (
	ExporterEnabledDefault bool   = true
	ExporterPortDefault    int    = 50051
	ExporterTypeDefault    string = "otlpgrpc"

	ListenerEnabledDefault bool   = true
	ListenerPortDefault    int    = 12201
	ListenerTypeDefault    string = "udp"

	ProcessorEnabledDefault bool = true
)
