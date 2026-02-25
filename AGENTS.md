# AI agent & developer instructions (Badili project)

## Project vision
**Badili** is a high-throughput GELF-to-OTel pipeline. It is architected as 
three distinct microservices (listener, packager, exporter) connected via gRPC
and serialized with Protocol Buffers.

## Core architectural rules
### Serialization
- All internal data transfer must use `google.golang.org/protobuf`.

### Observability
- Use the **OpenTelemetry Go SDK** for internal service logging, metrics, and spans. 
- Do not use standard `log` or `fmt.Printf` for production logging.

## Logic distribution
- `cmd/badili`: Main entry point. Handle signal interrupts and dependency injection.
- `api/proto`: Protobuf definitions for internal service communication.
- `internal/`: Private application code.
    - `exporters`: Forwards data to OTel collectors.
    - `listeners`: Ingests GELF (UDP/TCP/HTTP).
    - `packagers`: Batches and transforms chunks into messages.

## Coding standards for AI agents
### Concurrency and scaling
- **Dynamic workers:** Do not hardcode worker counts. Use a dynamic worker pattern (e.g., a worker pool that scales based on channel depth or a `context`-managed goroutine per request).
- **Non-blocking:** Network I/O operations should not block the main execution thread; use buffered channels and select statements with context cancellation.

### State management
- **Statelessness:** Services should remain stateless where possible. 
- **Graceful Shutdown:** Use `context.WithCancel` and `sync.WaitGroup` to ensure all in-flight messages are processed or checkpointed before exit.

### Testing

### Error handling
- Use the "Return Early" pattern.
- Wrap errors with context: `fmt.Errorf("packager: failed to decode msgpack: %w", err)`.

## Tooling constraints
- **Go Version:** 1.21+
- **Protoc:** Use `protoc-gen-go` and `protoc-gen-go-grpc`.
- **Linting:** Follow `golangci-lint` standard rules.
