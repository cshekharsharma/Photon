# Photon

[![Release](https://img.shields.io/github/v/release/cshekharsharma/photon?display_name=tag&sort=semver)](https://github.com/cshekharsharma/photon/releases)
[![CI](https://github.com/cshekharsharma/photon/actions/workflows/ci.yml/badge.svg)](https://github.com/cshekharsharma/photon/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/cshekharsharma/photon/branch/main/graph/badge.svg)](https://codecov.io/gh/cshekharsharma/photon)
[![Coverage](https://github.com/cshekharsharma/photon/actions/workflows/coverage-gate.yml/badge.svg)](https://github.com/cshekharsharma/photon/actions/workflows/coverage-gate.yml)
[![Lint](https://github.com/cshekharsharma/photon/actions/workflows/lint.yml/badge.svg)](https://github.com/cshekharsharma/photon/actions/workflows/lint.yml)
[![Security](https://github.com/cshekharsharma/photon/actions/workflows/security.yml/badge.svg)](https://github.com/cshekharsharma/photon/actions/workflows/security.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/cshekharsharma/photon)](https://github.com/cshekharsharma/photon/blob/main/go.mod)
[![License](https://img.shields.io/github/license/cshekharsharma/photon)](./LICENSE)

Photon is a lightweight Go service toolkit for building production backends without rebuilding the same infrastructure plumbing in every service.

It gives teams a common foundation for HTTP and gRPC services, configuration, logging, telemetry, storage clients, cloud integrations, middleware, background workers, caching, and resilience utilities. The goal is not to hide Go behind a heavy framework. The goal is to keep normal Go code simple, observable, and consistent across services.

![Photon Logo](./docs/assets/photon.png)

## What Photon Is Good For

- Standardizing how Go services start, validate config, expose APIs, shut down, and emit telemetry.
- Reusing proven wrappers for storage, cache, cloud, notification, auth, and session concerns.
- Building services that need request middleware, retries, rate limits, background workers, and operational guardrails.
- Keeping business logic free from repeated setup code while still allowing direct use of normal Go libraries.

## Capability Map

| Area | What is included |
| --- | --- |
| HTTP | Chi routing, configurable server startup, middleware, CORS, compression, timeout handling, recovery, request IDs, heartbeat, graceful shutdown hooks |
| gRPC | Server and client wrappers, unary interceptors, logging, recovery, retries, health checks, TLS support |
| Config | Koanf-backed loading, file/raw-byte sources, validation, default delimiter handling, optional watcher integration, non-fatal `LoadE()` |
| Observability | Zerolog-backed logging, OpenTelemetry traces, metrics, logs, request context propagation |
| Storage | MySQL, PostgreSQL, MongoDB, Elasticsearch, Aerospike, Redis, Memcached, and in-process FlashDB |
| Caching | Provider-agnostic cache wrappers over Aerospike, Redis, Memcached, and FlashDB |
| Cloud | Provider interfaces with AWS implementations for S3, SQS, SNS, AppConfig, SES, and Rekognition |
| Coordination | Redis-backed locks, Redis-backed rate limiting, Consul service discovery, AppConfig watchers, retry with exponential backoff |
| Auth and Session | JWT, OIDC, secure session management, Redis and memory-backed session stores |
| Workers | Worker contracts, overseer lifecycle, restart policy, hooks, status snapshots |
| Notifications | Slack, Teams, and SQS-backed notification publishing with validation, FIFO dedupe support, and audit hooks |
| Utilities | REST helpers, CLI formatting, file system helpers, i18n data, encoding helpers, common data structures, type utilities |

## Install

Photon targets modern Go and currently declares:

```text
go 1.26
```

Install it with:

```bash
go get github.com/cshekharsharma/photon
```

If the repository is private, configure `GOPRIVATE` for your environment before running `go get`.

```bash
go env -w GOPRIVATE=github.com/cshekharsharma/*
```

## Quick Start

```go
package main

import (
	"net/http"
	"time"

	"github.com/cshekharsharma/photon/core/router"
	server "github.com/cshekharsharma/photon/server/http"
)

func main() {
	server.StartHttpServer(&server.ServerConfig{
		ServerPort:    8080,
		ReadTimeout:   5 * time.Second,
		WriteTimeout:  10 * time.Second,
		IdleTimeout:   120 * time.Second,
		RouteProvider: router.RouterCHI,
		HttpRoutes: []*server.HttpRoute{
			{
				UrlRoute:      "/health",
				RequestMethod: http.MethodGet,
				HttpHandler: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				},
			},
		},
	})
}
```

For more examples, see [docs/examples](./docs/examples).

## License

Photon is released under the [Apache License 2.0](./LICENSE).

## Production Defaults

Photon is designed around explicit validation and safe defaults:

- Config, HTTP server, Mongo, Postgres, notifier, and worker options validate early.
- Default HTTP timeouts are applied when unset.
- CORS rejects wildcard origins when credentials are enabled.
- Mongo and Postgres wrappers avoid mutating caller-owned config while normalizing defaults.
- Worker overseer supports context shutdown, bounded restart policy, lifecycle hooks, and status snapshots.
- Notification publishing supports validation, audit hooks, and FIFO dedupe when the queue client supports it.

See [Production Readiness](./docs/production-readiness.md) for the release checklist.

## Development

```bash
make tools
make configure
make test
make testcoverage
```

Useful checks before a release:

```bash
go test ./...
go vet ./...
staticcheck ./...
golangci-lint run ./...
govulncheck ./...
go test -race ./cloud ./cloud/providers ./cloud/service/aws ./core/logger ./telemetry ./coordination/network/discovery ./coordination/network/watcher ./middleware ./server/grpc/server ./server/http ./storage/aerospike ./storage/memcached ./storage/elasticsearch ./storage/mongo ./storage/mysql ./storage/redis ./utils/rest ./notifier/notifyrclient ./workers
```

## Contributing

Contributions should be small, tested, and easy to review.

- Branch from `main`.
- Keep public APIs backward-compatible unless the change is intentionally breaking.
- Add focused tests for new behavior and failure paths.
- Run formatting, linting, tests, race tests where relevant, and vulnerability checks.
- Update docs when behavior, configuration, or operational expectations change.

## Security

Please do not open public issues for suspected vulnerabilities. Follow the process in [SECURITY.md](./SECURITY.md).
