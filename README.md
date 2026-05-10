# worker-sandbox

Background worker sandbox for event-driven jobs.

The first implementation subscribes to one NATS / JetStream topic, logs the
received job, and acknowledges the message after the handler returns without an
error.

## features

- NATS / JetStream worker consumer
- typed JSON job payload decoding through `lib-async`
- at-least-once delivery with ack / nack handling
- OpenTelemetry tracing, structured logs, and optional Pyroscope profiling
- trace propagation from publisher messages to worker handlers when tracing is
  enabled

## requirements

- go 1.26 or higher
- NATS for local runtime

Optional:

- [air](https://github.com/air-verse/air) for live reload
- [prettier](https://prettier.io/) for Markdown, YAML, JSON, and JSONC
  formatting

## project structure

```bash
worker-sandbox/
├── cmd/
│   └── server/          # worker entrypoint
├── internal/
│   ├── config/          # config loading and structs
│   └── worker/          # job payload and handler
├── config.example.yml
├── Dockerfile
├── Makefile
├── go.mod
└── README.md
```

## configuration

Copy `config.example.yml` to `config.yml` and adjust local values:

```bash
cp config.example.yml config.yml
```

Important sections:

- `service`: service name and shutdown timeout
- `logging`: slog level and trace id injection
- `tracing`: OTLP exporter settings
- `profiling`: Pyroscope settings
- `nats`: broker URL, JetStream toggle, durable name, and retry settings
- `worker`: topic to consume

## run locally

```bash
make run
```

## message shape

```json
{
    "id": "job-1",
    "type": "debug.print",
    "payload": {
        "message": "hello"
    }
}
```

## delivery behavior

With JetStream enabled, the worker uses at-least-once delivery. A handler error
causes a nack and the broker can redeliver the message. That means slow jobs
such as PDF generation, email, or notification delivery should be idempotent.

Core NATS can be used for fire-and-forget events, but it does not provide the
same durable ack behavior.

## available commands

| Command       | Description                               |
| ------------- | ----------------------------------------- |
| `make air`    | Run the worker with Air live reload       |
| `make tidy`   | Run `go mod tidy`                         |
| `make run`    | Start the worker locally                  |
| `make lint`   | Run Go and Markdown linting               |
| `make fmt`    | Format Go code with `go fmt`              |
| `make pretty` | Format Markdown, YAML, JSON, and JSONC    |
| `make format` | Run Go and document/config formatting     |
| `make test`   | Run tests with the race detector          |
| `make cov`    | Generate and open an HTML coverage report |
| `make fix`    | Apply standard Go source rewrites         |
