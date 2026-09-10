# Security Policy

## Supported Versions

Security fixes are provided for the latest tagged minor release. Older releases may receive fixes when the impact is high and the patch can be applied safely.

## Reporting a Vulnerability

Please report suspected vulnerabilities privately to the repository maintainers. Do not open a public issue for security-sensitive reports.

Include:

- affected package, version, and commit if known
- minimal reproduction steps
- expected and observed impact
- affected runtime environment
- any known workaround or mitigation

## Security Expectations

- Do not commit secrets, tokens, private keys, credentials, or production endpoints that contain credentials.
- Do not log bearer tokens, API keys, passwords, private keys, session IDs, or sensitive user data.
- Prefer environment variables, cloud-native identity, or a secret manager for credentials.
- Validate configuration, request payloads, notification payloads, worker inputs, and storage connection settings.
- Use TLS for external service calls and gRPC connections where supported.
- Review dependency upgrades before merge, but prioritize security upgrades ahead of routine feature work.

## Required Release Checks

```bash
go test ./...
go vet ./...
staticcheck ./...
golangci-lint run ./...
govulncheck ./...
go test -race ./cloud ./cloud/providers ./cloud/service/aws ./core/logger ./telemetry ./coordination/network/discovery ./coordination/network/watcher ./middleware ./server/grpc/server ./server/http ./storage/aerospike ./storage/memcached ./storage/elasticsearch ./storage/mongo ./storage/mysql ./storage/redis ./utils/rest ./notifier/notifyrclient ./workers
```
