# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Badili** (Swahili: "to change/transform") is a high-throughput GELF-to-OpenTelemetry pipeline. It ingests GELF log packets and exports them as OTel LogRecords to OTLP-compatible collectors (e.g. Uptrace).

## Commands

```bash
# Generate Protobuf Go code from .proto files
make proto

# Build Docker image
make docker-build

# Remove Docker image
make docker-clean

# Remove generated .pb.go files
make clean

# Run tests
go test ./...

# Run a single test
go test ./internal/listener/gelf/... -run TestName

# Lint
golangci-lint run
```

## Architecture

Three internal microservices communicate via gRPC + Protocol Buffers:

```
UDP:12201 → Listener → (gRPC) → Processor → (gRPC) → Exporter → OTLP Collector
```

1. **`internal/listener/gelf/`** — Receives UDP packets, decompresses (gzip/zlib), decodes JSON → `gelfapi.Message` Protobuf, sends via gRPC to the exporter.
2. **`internal/processor/gelf/`** — Aggregates chunked GELF messages (currently WIP).
3. **`internal/exporter/otlp/`** — gRPC server that receives batched messages and converts them to `otellog.Record`, emitting via the OTLP batch processor.
4. **`internal/grpcapi/`** — gRPC service implementations (`MessageService`) and the client-side message batcher (100 items or 2000ms flush).
5. **`api/gelfapi/v1/`** — Protobuf definitions; generated files (`*.pb.go`, `*_grpc.pb.go`) must not be edited by hand — run `make proto`.

**Entry point:** `cmd/badili/main.go` wires up config, telemetry, listener, and exporter supervisors, then blocks on SIGINT/SIGTERM.

**Configuration:** Viper-based. Sources in priority order: env vars (`BADILI_` prefix) > `config.yaml` in CWD > compiled defaults. Key defaults: listener port `12201`, exporter gRPC port `50051`.

## Architectural Rules (from AGENTS.md)

- All internal data transfer uses `google.golang.org/protobuf` — no other serialization formats internally.
- Use the **OpenTelemetry Go SDK** for internal logging/metrics/spans. Never use `log` or `fmt.Printf` for production logging.
- `cmd/badili` is only for signal handling and dependency injection; business logic belongs in `internal/`.
- **No hardcoded worker counts.** Use dynamic worker pools that scale based on channel depth or context-managed goroutines.
- Network I/O must be non-blocking: use buffered channels and `select` with context cancellation.
- Services must be stateless where possible.
- Graceful shutdown via `context.WithCancel` + `sync.WaitGroup`.
- Error handling: "Return Early" pattern; wrap errors with context using `fmt.Errorf("component: action: %w", err)`.