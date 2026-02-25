# Badili

### The swift GELF-to-OTel converter

> **Badili** (/ba.di.li/): Swahili for _to change, transform, or exchange._

**Badili** is high-performance middleware designed to bridge the gap between
legacy logging and modern observability. It ingests **GELF** (Graylog Extended
Log Format) packets and translates them into **OpenTelemetry** (OTel) logs,
ensuring your historical data flows seamlessly into the future of distributed
tracing and monitoring.

## Why Badili?

In the ever-evolving world of observability, moving from a centralized Graylog
setup to an OpenTelemetry-native stack shouldn't feel like a requirement for
complete refactor. **Badili** does the heavy lifting, so you don't have to
rewrite your application's logging logic.

- **Zero-loss translation:** Preserves GELF's `short_message`, `full_message`,
and all custom `_` fields, mapping them accurately to OTel LogRecord attributes.
- **High velocity:** Written for low-latency environments, handling
high-throughput UDP/TCP streams without breaking a sweat.
- **Context rich:** Automatically injects resource attributes (host,
environment, version) into your OTel signals during the conversion process.
- **Native OTLP export:** Direct support for OTLP/gRPC and OTLP/HTTP exporters.

## Features at a Glance

| **Feature**             | **Description**                                                                            |
|-------------------------|--------------------------------------------------------------------------------------------|
| **GELF listener**       | Supports compressed (GZIP/ZLIB) and uncompressed GELF via UDP/TCP/HTTP.                    |
| **Field mapping**       | Intelligent mapping of GELF levels (0-7) to OTel severity numbers.                         |
| **Metadata enrichment** | Add static or dynamic tags to every converted packet for better filtering in your backend. |
| **Reliability**         | Built-in backpressure handling to ensure your telemetry pipeline stays healthy.            |

## Quickstart

### Configuration (`config.yaml`)

Badili uses a simple YAML-based configuration. You can define your ingestion
sources and your OpenTelemetry destinations here.

YAML

```
# Badili Configuration File
# TODO
```

## GELF to OpenTelemetry mapping

Badili performs an intelligent mapping to ensure that your logs remain
searchable and structured.

| **GELF field**  | **OTel LogRecord field**             | **Notes**                                                                    |
|-----------------|--------------------------------------|------------------------------------------------------------------------------|
| `version`       | `attributes["gelf.version"]`         | Usually "1.1"                                                                |
| `host`          | `resource.attributes["host.name"]`   | Mapped to the standard OTel resource attribute                               |
| `short_message` | `Body`                               | The primary log message                                                      |
| `full_message`  | `attributes["exception.stacktrace"]` | Often used for stack traces in GELF; mapped to OTel attributes               |
| `timestamp`     | `Timestamp`                          | Converted from Unix Epoch (seconds) to OTel Nanoseconds                      |
| `level`         | `SeverityNumber`                     | **Mapped 0-7:** (e.g., GELF 3 [Error] → OTel 17 [Error])                     |
| `_user_id`      | `attributes["app.extra._user_id"]`   | All custom `_` fields are stripped of the underscore and added as attributes |

## Contributing
