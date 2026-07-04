# KafkaFlux — Agent Guide

## Build & run

```sh
# local build (Kafka mode requires CGO + librdkafka)
go build -o kafka-producer ./cmd/producer/main.go

# Docker build (CGO enabled, musl tags)
CGO_ENABLED=1 go build -tags musl -ldflags="-w -s" -o kafka-producer ./cmd/producer/main.go

# interactive profile generator (creates profiles/<entity>.yaml)
go run ./cmd/generator/main.go
```

## Run modes

Controlled by `SIMULATOR_MODE` env var:

| Mode    | Output | Env required                          |
|---------|--------|---------------------------------------|
| `kafka` | Kafka  | `KAFKA_BROKERS` (default `kafka:29092`) |
| `json`  | Files  | `OUTPUT_FILE_PATH` (default `./data_output`) |
| `csv`   | Files  | same as json                          |

Default is `kafka`. Run locally without Kafka:

```sh
SIMULATOR_MODE=json OUTPUT_FILE_PATH=./data_output go run ./cmd/producer/main.go
```

Run locally against Docker Kafka (kafka on localhost:9092):

```sh
SIMULATOR_MODE=kafka KAFKA_BROKERS=localhost:9092 go run ./cmd/producer/main.go
```

## Docker

```sh
docker compose up
```

Starts Zookeeper, Kafka (port 9092), and the producer. `profiles/`, `config.yaml`, and `data_output/` are volume-mounted — no rebuild needed to change profile YAMLs.

## Project structure

- `cmd/producer/main.go` — main entrypoint; loads config, profiles, selects transport, starts workers + dashboard
- `cmd/generator/main.go` — interactive YAML profile generator
- `internal/config/` — config loading + profile compiler (field generators, distributions, conditionals)
- `internal/engine/` — simulator workers (`runWorker`), live terminal dashboard, `StateRegistry` for cross-profile reference pools
- `internal/transport/` — `DataPublisher` interface + Kafka / File implementations
- `profiles/*.yaml` — one YAML per entity (e.g. `orders.yaml`, `customers.yaml`)
- `config.yaml` — global settings: `workers` (8), `profiles_dir`, `kafka_servers`

## Profile YAML — high-signal facts

**Field generator keywords** (single `value` entry):
`uuid`, `int`, `float`, `timestamp`, `first_name`, `last_name`, `name`, `email`, `phone`

`state_action: publish` + `state_pool: <name>` publishes generated values to a named pool; other profiles read them via `pool(<name>)`.

**Distributions & ranges:**
- `range(min, max)` — uniform int
- `normal_distribution(mean=...,stddev=...,min=...)` — clamped Gaussian
- `poisson_distribution(lambda=...)` — Poisson count

**Weighted enums:** multiple `value`/`weight` pairs; weights auto-normalized.

**Conditional fields** must execute after base fields — the compiler places them last automatically.
Syntax: `conditional(field_name = VALUE -> generator; default -> fallback)`

**Chaos injection:**
- `drop_percentage` — chance to skip producing an event
- `corrupt_fields.<name>.rate` — chance to replace field with `"NULL"` or `"CHAOS_CORRUPTION_ERR"`

**Topic naming convention:** `telemetry.ecommerce.<entity>`

## Constraints

- `confluent-kafka-go` requires CGO + `librdkafka`. File mode avoids this dependency.
- Channel buffer: 100 000 events. Terminal dashboard shows fill level.
- Dynamic scaling (when `dynamic_scaling: true`): sine wave over 600s, range 0.1x–1.6x of `target_eps`.
- Graceful shutdown on SIGINT/SIGTERM flushes pending events.

## No tests, no linting, no CI

Zero `_test.go` files, no `.golangci.yml`, no GitHub Actions workflows, no Makefile.
Build verification: `go build ./...`
