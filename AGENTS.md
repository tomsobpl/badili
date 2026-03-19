# AI agent & developer instructions (Badili project)

## Project vision

**Badili** (Swahili: "to change/transform") is a high-throughput GELF-to-OTel pipeline.
It ingests GELF log packets over UDP and exports them as OpenTelemetry `LogRecord`s to an
OTLP-compatible collector (e.g. Uptrace). Internally, three microservices communicate via
gRPC + Protocol Buffers:

```
UDP:12201 → Listener → (gRPC) → Processor → (gRPC) → Exporter → OTLP Collector
```

---

## Directory map

```
cmd/badili/          – binary entry point (signal handling + DI only)
api/gelfapi/v1/      – Protobuf + gRPC definitions (generated files, do not edit by hand)
internal/
  config/            – Viper-based configuration loader and compiled defaults
  listener/gelf/     – UDP socket, packet decompress, JSON decode → Protobuf Message
  processor/gelf/    – Chunk aggregation (WIP – stub only)
  exporter/otlp/     – gRPC server, Message → OTel LogRecord, OTLP/gRPC emission
  grpcapi/           – Shared gRPC server (MessageService) and client batcher (MessageWorker)
  telemetry/         – OTel tracer/logger provider setup; GELF→OTel severity mapping
  logging/           – slog + OTel handler initialisation
examples/            – Docker Compose + Uptrace config for local testing
```

---

## Data flow (detailed)

```
UDP socket (port 12201)
  └─ raw packets → buffered channel (cap 1024)
        └─ Listener workers (×N)
              ├─ detect magic bytes → chunked / gzip / zlib / plain
              ├─ decompress
              ├─ unmarshal JSON → gelfapi.Message (Protobuf)
              └─ MessageWorker (gRPC client batcher)
                    ├─ accumulate up to 100 messages OR 2 s flush
                    └─ StreamMessageBatch RPC → Exporter gRPC server :50051
                          └─ messages channel (cap 1024)
                                └─ Exporter workers (×N)
                                      ├─ gelfapi.Message → otellog.Record
                                      └─ Logger.Emit() → OTLP/gRPC batch processor → Uptrace :4317
```

### GELF → OTel LogRecord field mapping

| GELF field       | OTel LogRecord            |
|------------------|---------------------------|
| `short_message`  | Body                      |
| `level` (0–7)    | Severity (see table below)|
| `timestamp`      | Timestamp (Unix epoch)    |
| `version`        | Attribute `gelf.version`  |
| `host`           | Attribute `host.name`     |
| `full_message`   | Attribute `exception.stacktrace` |
| `_key` (extras)  | Attribute `app.extra.key` |

### GELF level → OTel severity

| GELF | Syslog name  | OTel Severity  |
|------|--------------|----------------|
| 0    | Emergency    | SeverityFatal4 |
| 1    | Alert        | SeverityFatal1 |
| 2    | Critical     | SeverityError4 |
| 3    | Error        | SeverityError1 |
| 4    | Warning      | SeverityWarn   |
| 5    | Notice       | SeverityInfo2  |
| 6    | Informational| SeverityInfo1  |
| 7    | Debug        | SeverityDebug  |

---

## gRPC / Protobuf API (`api/gelfapi/v1/`)

### MessageService (`message.proto`)
- `StreamMessageBatch(stream MessageBatch) → MessageBatchSummary`
- `MessageBatch` carries repeated `Message` (version, host, short_message, full_message,
  timestamp, level, extras as `google.protobuf.Struct`).
- **Currently wired** in both listener worker (client) and exporter (server).

### ChunkService (`chunk.proto`)
- `StreamChunkBatch(stream ChunkBatch) → ChunkBatchSummary`
- `Chunk` carries message_id, sequence_num, sequence_count, payload (bytes), timestamp.
- **Not yet wired** — defined for future processor implementation.

**Generated files** (`*.pb.go`, `*_grpc.pb.go`) must never be edited by hand.
Regenerate with `make proto`.

---

## Configuration (`internal/config/`)

Sources, highest priority first:
1. Environment variables — prefix `BADILI_` (e.g. `BADILI_LISTENER_PORT=12201`)
2. `config.yaml` in the current working directory
3. Compiled defaults

| Key                  | Default     | Description                      |
|----------------------|-------------|----------------------------------|
| `listener.enabled`   | `true`      |                                  |
| `listener.port`      | `12201`     | UDP port                         |
| `listener.type`      | `"udp"`     |                                  |
| `processor.enabled`  | `true`      |                                  |
| `exporter.enabled`   | `true`      |                                  |
| `exporter.port`      | `50051`     | Internal gRPC server port        |
| `exporter.type`      | `"otlpgrpc"`|                                  |

---

## Core architectural rules

### Serialization
- All internal data transfer must use `google.golang.org/protobuf`. No other serialisation
  formats (JSON, msgpack, etc.) are permitted between internal services.

### Observability
- Use the **OpenTelemetry Go SDK** for all internal logging, metrics, and tracing.
- Never use `log`, `fmt.Printf`, or bare `fmt.Println` for production logging.
- Structured logs via `slog` backed by the OTel handler (`internal/logging/slog.go`).

### Logic placement
- `cmd/badili`: Signal handling and dependency injection only — no business logic.
- `internal/`: All business logic. Never import `internal/` packages from `cmd/` except
  for wiring.

### Concurrency and scaling
- **No hardcoded worker counts.** Worker pools must scale dynamically (e.g. based on
  channel depth or a `context`-managed goroutine per request).
- **Non-blocking I/O.** Use buffered channels and `select` with context cancellation;
  never block the goroutine supervisor on a single operation.

### State management
- **Stateless services.** Services must not retain message state between requests except
  where explicitly required (e.g. chunk reassembly in the processor).
- **Graceful shutdown.** Use `context.WithCancel` + `sync.WaitGroup`. Drain in-flight
  messages before exit; the `MessageWorker` must flush its remaining batch.

### Error handling
- Use the **Return Early** pattern — handle errors at the call site, not deep in call stacks.
- Wrap errors with context: `fmt.Errorf("component: action: %w", err)`.
- Supervisors must recover from panics and restart workers with backoff rather than crashing
  the whole process.

---

## Supervisor pattern

Each component uses the same pattern:

```go
func StartXxxSupervisor(ctx context.Context, wg *sync.WaitGroup, ...) {
    defer wg.Done()
    for {
        if err := runXxx(ctx, ...); err != nil {
            if ctx.Err() != nil { return }   // intentional shutdown
            slog.Error("worker failed, restarting", "err", err)
            time.Sleep(5 * time.Second)
        }
    }
}
```

- A supervisor goroutine spawns worker goroutines and restarts them on failure.
- Workers check `ctx.Done()` for clean termination.

---

## Known gaps / WIP areas

| Area | Status | Notes |
|---|---|---|
| `internal/processor/gelf/` | **Stub** | `ProcessorWorker` logs chunks but does not reassemble them |
| `ChunkService` gRPC | **Defined, not wired** | Needs processor implementation first |
| Metrics | **Missing** | Only traces and logs are emitted; no OTel metrics |
| Tests | **Missing** | No test files exist yet |
| Dynamic worker counts | **Violated** | Both supervisor implementations currently hardcode 5 workers |
| Configurable endpoints | **Partial** | gRPC listener connects to hardcoded `127.0.0.1:50051`; Uptrace endpoint hardcoded to `uptrace:4317` |

---

## Tooling

| Tool | Version / Notes |
|---|---|
| Go | 1.25+ |
| `protoc-gen-go` | Use with `protoc-gen-go-grpc`; run via `make proto` |
| Linter | `golangci-lint` — follow standard rule set |
| Docker | Multi-stage build; see `Dockerfile` and `make docker-build` |

---

## Common commands

```bash
make proto          # regenerate *.pb.go from .proto files
make docker-build   # build badili:latest Docker image
make docker-clean   # remove Docker image
make clean          # remove generated .pb.go files
go test ./...       # run all tests
golangci-lint run   # lint
```
