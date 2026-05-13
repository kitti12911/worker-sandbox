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

## ci commands

reusable CI entrypoints live in `scripts/ci/` so GitHub Actions and GitLab CI
can call the same commands with provider-specific orchestration around them.

| command                                            | purpose                                       |
| -------------------------------------------------- | --------------------------------------------- |
| `./scripts/ci/go-lint.sh`                          | run `go vet` and `golangci-lint`              |
| `./scripts/ci/go-test.sh`                          | run tests with coverage                       |
| `./scripts/ci/markdownlint.sh`                     | run Markdown linting                          |
| `./scripts/ci/security-scan.sh`                    | run `govulncheck` and Semgrep                 |
| `./scripts/ci/supply-chain-scan.sh`                | run Trivy and Gitleaks                        |
| `./scripts/ci/semantic-release-plan.sh`            | preview the next semantic release             |
| `./scripts/ci/semantic-release-publish.sh`         | publish the semantic release                  |
| `./scripts/ci/fast-forward-prerelease-branches.sh` | fast-forward `uat` and `develop` after `main` |
| `./scripts/ci/update-helm-image-values.sh`         | update homelab GitOps image values            |

GitHub Actions uses `TOOLCHAIN_REGISTRY` and `TOOLCHAIN_IMAGE_NAMESPACE` to
resolve shared CI toolchain images, and `IMAGE_REGISTRY` plus `IMAGE_NAMESPACE`
to publish the application image. GitLab uses full image references so the
private mirror can point at Harbor without changing these scripts:

| GitLab variable                   | Purpose                                     |
| --------------------------------- | ------------------------------------------- |
| `CI_IMAGE_TOOLCHAIN_IMAGE`        | Image for Go lint/test and builds           |
| `CI_SECURITY_TOOLCHAIN_IMAGE`     | Image for `govulncheck` and Semgrep         |
| `CI_SUPPLY_CHAIN_TOOLCHAIN_IMAGE` | Image for Trivy and Gitleaks                |
| `CI_RELEASE_TOOLCHAIN_IMAGE`      | Image for Markdownlint and semantic-release |
| `CI_DOCKER_CLI_IMAGE`             | Docker CLI image for build/publish jobs     |
| `CI_DOCKER_DIND_IMAGE`            | Docker-in-Docker service image              |
| `CI_TRIVY_RUNNER_IMAGE`           | Optional Trivy runner image override        |
| `IMAGE_REGISTRY`                  | Target application image registry           |
| `IMAGE_NAMESPACE`                 | Target application image namespace          |
| `GITLAB_AMD64_RUNNER_TAG`         | Optional runner tag override                |
| `GL_TOKEN` or `GITLAB_TOKEN`      | GitLab semantic-release API/write token     |

| GitLab secret                        | Purpose                  |
| ------------------------------------ | ------------------------ |
| `IMAGE_REGISTRY_USERNAME`            | Target registry username |
| `IMAGE_REGISTRY_PASSWORD`            | Target registry password |
| `COSIGN_PRIVATE_KEY` or `COSIGN_KEY` | Image signing key        |

The `homelab-devops` values update in `.github/workflows/go-ci.yml` is
GitHub-specific homelab orchestration, not part of the portable script contract.
The prerelease branch fast-forward helper is also GitHub-specific because it
pushes through a GitHub App token.
GitLab deployments can use a different project, folder layout, or deployment
tool by calling the same `scripts/ci` build/release helpers and adding its own
deploy job. `DEPLOY_IMAGE_REGISTRY` and `DEPLOY_IMAGE_NAMESPACE` only affect the
homelab GitOps values update and can be omitted outside that workflow.

`GO_TEST_RACE=true` or `GO_TEST_CGO=true` requires a C compiler in the selected
toolchain image. `worker-sandbox` sets `GO_TEST_RACE=false` in GitHub Actions
while using `image-toolchain` v1.1.0 because that image does not include one.

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
