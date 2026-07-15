# KafkaFlux

Generate realistic event streams to Kafka or JSON/CSV files. Built for load testing, demo environments, and data pipelines.

## Quick Start

```sh
docker compose up
```

Starts Zookeeper, Kafka, and the producer. You're generating fake ecommerce events (orders, customers, payments, clickstream) to Kafka topics on `localhost:9092`.

---

### Download

Pre-built binaries are available for each [release](https://github.com/prajeesh-chavan/KafkaFlux/releases).

```sh
# JSON/CSV mode (no dependencies, static binary)
curl -LO https://github.com/prajeesh-chavan/KafkaFlux/releases/latest/download/kafkaflux-producer-v0.2.0-linux-amd64
chmod +x kafkaflux-producer-v0.2.0-linux-amd64
SIMULATOR_MODE=json ./kafkaflux-producer-v0.2.0-linux-amd64

# Kafka mode (requires librdkafka on the system)
curl -LO https://github.com/prajeesh-chavan/KafkaFlux/releases/latest/download/kafkaflux-producer-v0.2.0-kafka-linux-amd64
chmod +x kafkaflux-producer-v0.2.0-kafka-linux-amd64
./kafkaflux-producer-v0.2.0-kafka-linux-amd64
```

> Replace `v0.2.0` with the latest version from the [releases page](https://github.com/prajeesh-chavan/KafkaFlux/releases).

---

## Features

- **50+ field types** — uuid, int, float, boolean, names, emails, phones, addresses, IPs, user agents, credit cards, lat/lng, sentences, distributions (normal, poisson, range), conditionals, weighted enums, state pools
- **11 starter profiles** — orders, customers, products, payments, clickstream, shipping, support tickets, reviews, IoT sensors, access logs, subscriptions, notifications
- **Output** — Kafka (default), JSON files, CSV files
- **Chaos injection** — random drop percentage, per-field corruption rate
- **Dynamic scaling** — sinusoidal traffic patterns over a 10-minute window
- **Prometheus metrics** — events/sec, buffer fill, delivery failures at `:9099/metrics`
- **Structured logging** — `log/slog` with configurable level (`info`, `debug`, `warn`, `error`)

---

## Usage

### Profile Generator

```sh
# Start interactive mode
go run ./cmd/generator

# Show field type reference
go run ./cmd/generator --help

# Generate a starter profile
go run ./cmd/generator --init orders

# Batch mode (non-interactive)
go run ./cmd/generator --entity myentity \
  --field "name=id,type=uuid" \
  --field "name=price,type=normal,mean=50,stddev=10,min=1"

# Validate existing profiles
go run ./cmd/generator --validate profiles/*.yaml
```

### Run Modes

```sh
# Kafka (default)
docker compose up

# JSON files
SIMULATOR_MODE=json go run ./cmd/producer/main.go

# CSV files
SIMULATOR_MODE=csv OUTPUT_FILE_PATH=./data_output go run ./cmd/producer/main.go
```

### Configuration

Edit `config.yaml` to change workers, profiles directory, Kafka brokers, metrics port, and log level. Environment variables override YAML:

| Variable | Overrides | Default |
|---|---|---|
| `SIMULATOR_MODE` | — | `kafka` |
| `KAFKA_BROKERS` | `kafka_servers` | `kafka:29092` |
| `OUTPUT_FILE_PATH` | — | `./data_output` |

---

## Profile YAML

Profiles define entities and their fields. Each field has a `type` and optional parameters.

```yaml
entity: orders
topic: "telemetry.ecommerce.orders"
target_eps: 50
dynamic_scaling: true

fields:
  order_id:
    type: uuid
    publish_to: orders
  customer_id:
    type: pool
    pool: customers
  order_status:
    type: weighted
    values:
      CREATED: 20
      COMPLETED: 30
      CANCELLED: 5
  amount:
    type: normal
    mean: 100.0
    stddev: 25.0
    min: 5.0
  completed_at:
    type: conditional
    rules:
      - when: order_status == COMPLETED
        then:
          type: timestamp
```

Run `generator --help` for the full type reference.

---

## Architecture

```
config.yaml ──> app.Run() ──> engine.Simulator ──> transport.Publisher ──> Kafka / File
                    │               │                        │
              telemetry        pool.BufferPool          telemetry.Metrics
              (slog +           (sync.Pool               (Prometheus /metrics)
               Metrics           for byte reuse)
               HTTP server)
```

Packages: `app`, `config`, `engine`, `field`, `pool`, `telemetry`, `transport`.

---

## Development

```sh
go test -count=1 ./...   # 57 tests
go vet ./...             # static analysis
go build ./...           # verify compilation
```

Requires CGO + `librdkafka` for Kafka mode. Use `SIMULATOR_MODE=json` to develop without Kafka.
