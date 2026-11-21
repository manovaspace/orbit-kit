# orbit-kit

[![CI](https://github.com/manovaspace/orbit-kit/actions/workflows/ci.yml/badge.svg)](https://github.com/manovaspace/orbit-kit/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE)

orbit-kit is the Go kit for Orbit services. It holds three packages in one module:

- rate limiting for API edges,
- observability (logs, traces, health checks, Sentry),
- a message bus with several transports.

Part of the [Manova / Orbit](https://github.com/manovaspace) open toolkit. Status: `beta`.

## Install

```bash
go get github.com/manovaspace/orbit-kit@latest
```

## Packages

| Import path | What it does | Main dependencies |
| --- | --- | --- |
| `github.com/manovaspace/orbit-kit/ratelimiting` | Token-bucket rate limiter. Runs in memory or on Redis (Lua). Helpers build bucket keys for organizations and IP addresses. | `github.com/redis/go-redis/v9` |
| `github.com/manovaspace/orbit-kit/observability` | JSON `slog` logging, OpenTelemetry traces, health and readiness handlers, HTTP and gRPC middleware, shutdown on signal, optional Sentry. | `go.opentelemetry.io/otel` and its OTLP exporter, `github.com/getsentry/sentry-go`, `google.golang.org/grpc` |
| `github.com/manovaspace/orbit-kit/messagebus` | `Bus` port with three transports: NATS JetStream (root package), Redis Streams (`redisbus`) and in-memory (`inproc`). Also a transport factory (`transport`), a Postgres transactional outbox (`outbox`) and a shared test suite (`bustest`). | `github.com/nats-io/nats.go`, `github.com/redis/go-redis/v9`, `github.com/jackc/pgx/v5` (outbox only), `github.com/google/uuid`, `go.opentelemetry.io/otel` |

Each package has its own README with usage examples:
[ratelimiting](./ratelimiting/README.md), [observability](./observability/README.md),
[messagebus](./messagebus/README.md).

## Migrating from the old modules

The three old modules are archived. Their last releases were `orbit-rate-limiting` v0.2.2,
`orbit-observability` v0.2.1 and `orbit-message-bus` v0.1.0. Change the import paths in your code:

| Old import path | New import path |
| --- | --- |
| `github.com/manovaspace/orbit-rate-limiting` | `github.com/manovaspace/orbit-kit/ratelimiting` |
| `github.com/manovaspace/orbit-observability` | `github.com/manovaspace/orbit-kit/observability` |
| `github.com/manovaspace/orbit-message-bus` | `github.com/manovaspace/orbit-kit/messagebus` |

For each old path, run this in your repository, with `<old>` and `<new>` set to one row of the table:

```bash
grep -rl <old> --include='*.go' . | xargs sed -i 's#<old>#<new>#g'
```

Then update the module and drop the old requirements:

```bash
go get github.com/manovaspace/orbit-kit@latest && go mod tidy
```

Package names do not change, so the code after the import line needs no edits.

## License

MIT. See [LICENSE](./LICENSE).
