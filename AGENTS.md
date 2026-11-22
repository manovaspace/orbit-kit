---
see_also:
  - handbook/docs/orbit/architecture/contracts.md
  - handbook/docs/orbit/architecture/observability.md
  - handbook/docs/orbit/architecture/data-patterns.md
  - handbook/docs/orbit/architecture/messaging.md
  - handbook/docs/orbit/decisions/071-observability.md
  - handbook/docs/orbit/decisions/070-public-http-edge.md
---

# orbit-kit

Go kit for Orbit services. One module, `github.com/manovaspace/orbit-kit`, with three packages:
`ratelimiting`, `observability` and `messagebus`. Status: `beta`. MIT.

## Commands

```bash
gofmt -l .
go vet ./...
go test ./... -race -count=1
```

Network and database tests skip unless their environment variable is set. See `messagebus` below.

## Shared rules

- One module, one tag. A release versions all three packages together. See [RELEASING.md](./RELEASING.md).
- Keep the package names `ratelimiting`, `observability` and `messagebus`, and their subpackages.
- Exported names and signatures keep their previous names and shapes. A change to them is a breaking
  change: use `feat!:` or `fix!:` and note it in the PR.
- `.golangci.yml` `depguard` rules enforce the boundaries below. Do not weaken them to land a change.
- The libraries run no HTTP server of their own, own no database and publish no events. Services own
  the server, the database and the bus connection.
- Public docs live in the package READMEs and the root README. Staff handbook docs are private.
- Keep the repository free of client-specific product names and private URLs.

## ratelimiting

Redis (or in-memory) token-bucket limiter for API edges. The gateway consumes it as an adapter.

Layout:

- `ratelimiting/limiter.go` — public facade. `NewMemoryLimiter`, `NewRedisLimiter`, `Policy`,
  `Decision`, `Limiter`, `DefaultAuthPolicies` and `OTPRequestIdentifierPolicy` keep their previous
  names and signatures.
- `ratelimiting/keys.go` — `KeyForOrg` and `KeyForIP` bucket key helpers.
- `ratelimiting/internal/domain` — `Policy`, `Decision`, ADR-018 defaults and the `Limiter` port. It
  imports no Redis client, HTTP server or adapter.
- `ratelimiting/internal/application` — the rules every adapter shares: the non-positive-policy bypass
  and the one-second `Retry-After` floor.
- `ratelimiting/internal/infrastructure/memory` — process-local token bucket.
- `ratelimiting/internal/infrastructure/redis` — Redis token bucket (Lua).

Tests need no service and no environment variable.

## observability

Structured logging, OpenTelemetry, health checks, and Sentry hooks for Go services.

Layout:

- `observability/config.go`, `health.go`, `internal_auth.go`, `middleware.go`, `observability.go`,
  `shutdown.go` — the public facade. Every exported function and type keeps its previous name and
  signature.
- `observability/internal/domain` — `Config`, `Exporter` and the console/OTLP constants. It imports no
  telemetry SDK, gRPC, HTTP server or adapter.
- `observability/internal/application` — the health/readiness state machine and the Sentry redaction
  rule.
- `observability/internal/infrastructure/config` — the `DEPLOYMENT_ENVIRONMENT`,
  `OTEL_EXPORTER_OTLP_ENDPOINT`, `SENTRY_DSN` and `TENANT_ID` contract.
- `observability/internal/infrastructure/logging` — the ADR-009 JSON `slog` handler.
- `observability/internal/infrastructure/telemetry` — the only package that talks to OpenTelemetry and
  Sentry.
- `observability/internal/infrastructure/httpapi` — `/healthz` and `/readyz` handlers plus
  `otelhttp` middleware.
- `observability/internal/infrastructure/grpc` — `ORBIT_INTERNAL_TOKEN` interceptors and gRPC options.
- `observability/internal/infrastructure/shutdown` — SIGINT/SIGTERM drain and telemetry flush.

Rules:

- The `net/http` handlers and middleware stay `net/http`. Changing them would change the exported
  contract. Fiber v3 does not apply to this package.
- Tests need no service.

## messagebus

Message-bus port and transports: subject naming, envelope enforcement, retry (max 3) and DLQ.

Services depend only on `messagebus.Bus`. `transport.FromEnv` picks the transport at start-up from
configuration.

Transports:

| Transport | `BUS_TRANSPORT` | Endpoint |
| --- | --- | --- |
| root package (NATS JetStream) | `nats` | `BUS_NATS_URL`, else `Config.URL` |
| `messagebus/redisbus` (Redis Streams) | `redis` | `BUS_REDIS_URL` |
| `messagebus/inproc` (in-memory) | `inproc` (default) | none |

`transport.New(ctx, name, cfg, transport.Urls{Redis: ...})` selects a transport explicitly.
`transport.FromEnv(ctx, cfg)` reads the three environment variables.

Layout:

- `messagebus` (root) — `Bus`, `Envelope`, `Config`, validation, retry and DLQ; NATS implementation in
  `jetstream.go`. It does not import the transport packages.
- `messagebus/inproc` — in-memory `Bus` (`inproc.New(cfg)`), one dispatcher per subscription, no network.
- `messagebus/redisbus` — Redis Streams `Bus` (`redisbus.New(url, cfg)`); stream `orbit:<subject>`,
  group `orbit-<sanitized subject>`, acks, `XAUTOCLAIM` redelivery, DLQ stream `orbit:<subject>.dlq`.
- `messagebus/transport` — factory for `inproc`, `redis` and `nats`; the only package that imports all
  three.
- `messagebus/outbox` — Postgres transactional outbox for pgx v5. `Insert` writes an event in the
  caller's transaction. `MigrationSQL` is the goose migration each service copies. `Relay` (`RunOnce`,
  `Run`) publishes pending rows to any `messagebus.Bus`. Delivery is at-least-once, batches keep order,
  and rows are locked with `FOR UPDATE SKIP LOCKED`. Handlers must be idempotent on the envelope `id`.
- `messagebus/bustest` — shared conformance suite (`bustest.Conformance`), `bustest.UniqueDomain` and
  `bustest.CheckEventSchema`. `CheckEventSchema` pins an event struct's JSON shape against a committed
  file. `UPDATE_EVENT_SCHEMAS=1 go test ./...` refreshes it.
- `messagebus/testdata/envelope.schema.json` pins the library's own envelope.

Network and database tests:

- The NATS and Redis conformance tests skip unless `TEST_NATS_URL` or `TEST_REDIS_URL` is set. Each run
  uses `bustest.UniqueDomain()` and deletes the streams, consumers, groups and Redis keys it created.
- The outbox tests skip unless `TEST_DATABASE_URL` is set. Each test creates and drops its own random
  schema.
- Local NATS for development: `nats://localhost:10422` (ADR-006).

```bash
TEST_NATS_URL=nats://localhost:10422 \
TEST_REDIS_URL=redis://localhost:6379/15 \
TEST_DATABASE_URL='postgres://postgres:postgres@127.0.0.1:5432/test?sslmode=disable' \
go test ./... -race -count=1 -v
```

## Docs

- Public READMEs: [README.md](./README.md), [ratelimiting](./ratelimiting/README.md),
  [observability](./observability/README.md), [messagebus](./messagebus/README.md)
- Staff handbook (private): ADR-006, ADR-009, ADR-010, ADR-018, architecture contracts, observability,
  data patterns and messaging

<!-- manova:conventions v1 — generated from manovaspace/ci conventions/core.md; edit it there -->
## Shared conventions

These rules apply to every Manova repository. They are copied from `manovaspace/ci` (`conventions/core.md`); change them there, not here.

### Delivery

1. Never push to `main`. Every change goes on a topic branch (`feat/<name>`, `fix/<name>`, `refactor/<name>`, `docs/<name>`, `chore/<name>`) and reaches `main` through a pull request.
2. One pull request does one thing. An unrelated problem you find on the way becomes an issue, not part of this change.
3. Pull request titles and commit messages follow Conventional Commits (`feat:`, `fix:`, `docs:`, `refactor:`, `chore:`, `test:`; `!` for a breaking change). Releases are computed from them.
4. Merge with squash once the checks are green. Do not use `--no-verify`; if you ever must, say so in the pull request.

### Checks

5. CI runs `manovaspace/ci@v1`. Run the same gate locally before you push; a red check is fixed, not skipped.
6. Tool versions (Go, Node, Bun, linters) come from `ci/versions.env` in `manovaspace/ci`. Do not pin a different version in this repository.
7. A behaviour change comes with a test that fails without it. Do not delete or weaken a test to make a change pass.

### Secrets

8. Never commit a secret, token, password, private key or `.env` file. Commit `.env.example` with placeholders instead.
9. Never print a secret in logs, test output, error messages or pull requests. Read secrets from the environment or a file at run time.

### Runtime

10. Containers talk to each other by name on a Docker network. Do not publish a port on the host unless a person or a host process needs it, and then bind `127.0.0.1`, never `0.0.0.0`.
11. Configuration comes from the environment. A service with a missing required setting refuses to start; it never falls back to a built-in secret or a development default in production.
12. Port numbers come from the workspace port registry. Do not invent a port.

### Documentation

13. This repository explains itself: `README.md` says how to set up, run, test and deploy it; `AGENTS.md` holds its binding rules; repository decisions live in `docs/decisions/`.
14. Write plain, short sentences. Keep exact names, paths, commands and versions. Do not document what the code already says.
<!-- /manova:conventions -->
