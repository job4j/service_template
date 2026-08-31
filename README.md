# service_template

`servicegen` creates a Go microservice with Fiber, OpenAPI, Swagger UI, typed configuration, PostgreSQL connectivity, structured logs, OpenTelemetry traces, and Prometheus metrics.

The generator creates technical infrastructure only. Business operations, domain rules, repositories, and migrations remain the responsibility of the service developer.

## Install

Install the latest released version:

```bash
go install github.com/job4j/service_template/cmd/servicegen@latest
```

Install a specific version for reproducible course exercises:

```bash
go install github.com/job4j/service_template/cmd/servicegen@v0.1.0
```

Make sure the Go binary directory is available in `PATH`:

```bash
go env GOBIN
go env GOPATH
```

If `GOBIN` is empty, Go installs binaries into `$(go env GOPATH)/bin`.

## Create a service

```bash
servicegen new \
  --name sharetrip-contract \
  --module github.com/student/sharetrip-contract
```

By default, the project is created in a directory matching the service name. Use `--output` to choose another directory:

```bash
servicegen new \
  --name sharetrip-contract \
  --module github.com/student/sharetrip-contract \
  --output ./services/sharetrip-contract
```

Use `--port 8081` when the default port is already occupied.

The destination must not exist or must be empty. `servicegen` never overwrites a non-empty directory.

## Generated project

```text
sharetrip-contract/
├── api/
│   ├── contract.yaml
│   ├── health.go
│   ├── openapi.codegen.yaml
│   └── route.go
├── cmd/
│   └── main.go
├── internal/
│   ├── config/
│   ├── middleware/
│   ├── observability/
│   ├── service/
│   ├── domain/
│   └── repository/
│       └── entity/
├── migrations/
├── observability/
├── .env.example
├── Dockerfile
├── Makefile
├── README.md
└── go.mod
```

The generated OpenAPI contract contains a minimal `GET /health` operation. The student adds the service-specific API and business behavior afterwards.

## Run a generated service

```bash
cd sharetrip-contract
make tools
make generate
go mod tidy
make run
```

Open:

```text
http://localhost:8080/docs
```

The generated service exposes `/metrics`, propagates `X-Request-ID`, writes JSON logs to stdout and a file, and exports traces through OTLP. The `observability` directory contains integration files and instructions for Prometheus, Grafana, Loki, Alloy, and Jaeger.

## Development

Run tests:

```bash
go test ./...
```

Build the CLI:

```bash
go build ./cmd/servicegen
```
