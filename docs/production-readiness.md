# Production Readiness

Photon aims to give every service the same operational contract: validated startup, observable runtime behavior, bounded failure handling, and predictable shutdown.

## Startup Contract

- Validate configuration before opening listeners, storage connections, or worker loops.
- Prefer APIs that return errors, such as `LoadE()` and `SetConnectionConfigE()`, when writing libraries or service startup code.
- Initialize cloud services through owned `cloud.Client` instances so provider/config errors fail startup instead of surfacing inside the first downstream call.
- Shut down telemetry handles with a bounded context before process exit.
- Let `ServerConfig.Normalize()` apply HTTP defaults, then treat validation failures as deployment errors.
- Keep HTTP read, write, and idle timeouts explicit for internet-facing services.
- Avoid wildcard CORS origins when credentials are enabled. Photon rejects this combination for HTTP server config.

## Runtime Contract

- Carry request IDs, deadlines, and trace context across handlers, workers, and outbound calls.
- Emit structured logs through `core/logger`; avoid free-form logs for high-volume operational events.
- Record retry, timeout, rate-limit, and queue-publish outcomes with enough context to debug without exposing secrets.
- Expose heartbeat or readiness routes for load balancers and deployment checks.
- Use status snapshots and lifecycle hooks for supervised workers.

## Resilience Contract

- Put deadlines on storage, cache, cloud, queue, and downstream HTTP calls.
- Use bounded retries with exponential backoff only for transient failures.
- Make mutating operations idempotent when they can be retried or delivered more than once.
- Use queue idempotency keys and FIFO dedupe where the broker supports them.
- Keep shutdown hooks short, context-aware, and safe to run more than once.

## Security Contract

- Do not log bearer tokens, API keys, passwords, private keys, session IDs, or sensitive user data.
- Validate all externally supplied config, request payloads, queue payloads, and worker inputs.
- Use TLS for external service calls and gRPC connections where supported.
- Run dependency and vulnerability checks before every release.
- Treat credentials as runtime configuration from a secret manager or environment, never as committed source.

## Release Checklist

```bash
go test ./...
go vet ./...
staticcheck ./...
golangci-lint run ./...
govulncheck ./...
go test -race ./cloud ./cloud/providers ./cloud/service/aws ./core/logger ./telemetry ./coordination/network/discovery ./coordination/network/watcher ./middleware ./server/grpc/server ./server/http ./storage/aerospike ./storage/memcached ./storage/elasticsearch ./storage/mongo ./storage/mysql ./storage/redis ./utils/rest ./notifier/notifyrclient ./workers
```

Before tagging:

- Update `CHANGELOG.md`.
- Confirm `README.md` and `docs/examples` match the public API.
- Confirm any breaking change is called out clearly.
- Confirm package coverage gates remain at `100.0%`.
- Tag with semantic versioning.
