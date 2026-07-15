<div align="center">
  <h1>KafkaFlux</h1>
  <p><strong>Streaming Data Simulator — One Command. Any Schema.</strong></p>
  <p>
    <a href="#quick-start">Quick Start</a> •
    <a href="#download">Download</a> •
    <a href="#architecture">Architecture</a> •
    <a href="#features">Features</a> •
    <a href="#running">Running</a> •
    <a href="#example-profiles">Example Profiles</a> •
    <a href="#configuration">Configuration</a> •
    <a href="#metrics">Metrics</a>
  </p>
  <p>
    <img src="https://img.shields.io/badge/Go-1.25-00ADD8?logo=go" alt="Go 1.25">
    <img src="https://img.shields.io/badge/Kafka-7.5.0-231F20?logo=apachekafka" alt="Kafka 7.5.0">
    <img src="https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker" alt="Docker Compose">
    <img src="https://img.shields.io/badge/Prometheus-Metrics-E6522C?logo=prometheus" alt="Prometheus">
    <img src="https://img.shields.io/github/v/release/prajeesh-chavan/KafkaFlux?color=blue" alt="Release">
  </p>
  <p>Built by <a href="https://linkedin.com/in/prajeeshchavan"><strong>Prajeesh Chavan</strong></a></p>
</div>

---

KafkaFlux generates **realistic, high-volume event streams** to Kafka topics (or JSON/CSV files). Define your schema in YAML — orders, IoT telemetry, logs, financial transactions, user activity, or any domain — and get streaming data flowing in one command.

**Includes 30+ example profiles for ecommerce and 3 IoT profiles** to get started immediately. Bring your own schemas for anything else.

Built for **load testing**, **demo environments**, **data pipeline development**, and **CI/CD testing**.

---

## Quick Start

```sh
git clone https://github.com/prajeeshchavan/KafkaFlux
cd KafkaFlux
docker compose up
```

That's it. Zookeeper + Kafka + the simulator start together. Open another terminal:

```sh
curl localhost:9099/              # JSON status dashboard
curl localhost:9099/metrics       # Prometheus metrics
```

**Without Kafka (JSON output):**

```sh
SIMULATOR_MODE=json go run ./cmd/producer/main.go
```

---

## Download

Pre-built binaries for each [release](https://github.com/prajeesh-chavan/KafkaFlux/releases) — no Go or Docker required.

```sh
# JSON/CSV mode (static binary, no dependencies)
curl -LO https://github.com/prajeesh-chavan/KafkaFlux/releases/latest/download/kafkaflux-producer-linux-amd64
chmod +x kafkaflux-producer-linux-amd64
SIMULATOR_MODE=json ./kafkaflux-producer-linux-amd64

# Kafka mode (requires librdkafka on the system)
curl -LO https://github.com/prajeesh-chavan/KafkaFlux/releases/latest/download/kafkaflux-producer-kafka-linux-amd64
chmod +x kafkaflux-producer-kafka-linux-amd64
./kafkaflux-producer-kafka-linux-amd64
```

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

## Features

- **50+ field types** — uuid, int, float, boolean, names, emails, phones, addresses, IPs, user agents, credit cards, lat/lng, sentences, distributions (normal, poisson, range), conditionals, weighted enums, state pools
- **33 starter profiles** — 30 ecommerce entities (orders, customers, payments, inventory, shipments, etc.) + 3 IoT (sensors, device events, GPS tracking)
- **Profile organization** — domain subdirectories (`ecommerce/`, `iot/`), `enabled` flag, runtime filter via glob or entity name
- **Output** — Kafka (default), JSON files, CSV files
- **Deterministic mode** — same seed = same output, via `SIMULATOR_SEED` env or `seed` config
- **Batch mode** — generate N events per entity then exit, via `BATCH_SIZE` env or `batch_size` config
- **Graceful drain** — two-phase shutdown, zero event loss on SIGINT/SIGTERM
- **Chaos injection** — random drop percentage, per-field corruption rate
- **Dynamic scaling** — sinusoidal traffic patterns over a 10-minute window
- **Prometheus metrics** — events/sec, buffer fill, delivery failures at `:9099/metrics`
- **Structured logging** — `log/slog` with configurable level (`info`, `debug`, `warn`, `error`)
- **Pre-built binaries** — Linux amd64 + arm64, static (JSON/CSV) and Kafka-enabled

---

## Running

### Full Stack (Docker — Recommended)

```sh
docker compose up
```

Starts Zookeeper, Kafka, and the producer with live dashboard on port 9099.

### Pre-built Binary

```sh
SIMULATOR_MODE=json ./kafkaflux-producer-linux-amd64
```

### Without Kafka (JSON / CSV)

```sh
SIMULATOR_MODE=json go run ./cmd/producer/main.go
SIMULATOR_MODE=csv OUTPUT_FILE_PATH=./data_output go run ./cmd/producer/main.go
```

### Batch Mode

```sh
BATCH_SIZE=10000 SIMULATOR_MODE=json go run ./cmd/producer/main.go
```

### Deterministic Mode

```sh
SIMULATOR_SEED=42 SIMULATOR_MODE=json go run ./cmd/producer/main.go
```

### Profile Generator

```sh
go run ./cmd/generator                      # Interactive mode
go run ./cmd/generator --help               # Field type reference
go run ./cmd/generator --init orders         # Template profile
go run ./cmd/generator --validate profiles/  # Validate profiles
```

---

## Configuration

`config.yaml`:
```yaml
simulator:
  workers: 8              # Publisher goroutines
  profiles_dir: "./profiles"
  profiles: ["orders", "iot/*"]   # Runtime filter (empty = all enabled)
  kafka_servers: "kafka:29092"
  metrics_port: 9099
  log_level: "info"
  seed: 0                 # 0 = random, >0 = deterministic
  batch_size: 0           # 0 = continuous, >0 = stop after N per entity
```

Environment variable overrides:

| Variable | Overrides | Default |
|---|---|---|
| `SIMULATOR_MODE` | Transport mode | `kafka` |
| `KAFKA_BROKERS` | Kafka servers | `kafka:29092` |
| `OUTPUT_FILE_PATH` | Output directory | `./data_output` |
| `PROFILES` | Profile filter | config.yaml value |
| `SIMULATOR_SEED` | Deterministic seed | YAML value (0) |
| `BATCH_SIZE` | Batch event limit | YAML value (0) |

---

## Example Profiles

```
profiles/
├── ecommerce/           # 30 ecommerce entities
│   ├── orders.yaml
│   ├── customers.yaml
│   ├── payments.yaml
│   ├── inventory.yaml
│   ├── shipments.yaml
│   ├── products.yaml
│   └── ... (24 more)
└── iot/                 # 3 IoT entities
      ├── sensors.yaml
      ├── device_events.yaml
      └── gps_tracking.yaml
```

---

## Profile YAML

Profiles define entities and their fields. Each field has a `type` and optional parameters.

```yaml
entity: orders
topic: "telemetry.ecommerce.orders"
target_eps: 50
dynamic_scaling: true
batch_size: 5000                       # per-profile batch override
enabled: true                          # permanently disable without deleting

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
      PLACED: 20
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

## Development

```sh
go test -count=1 ./...     # 67 tests
go vet ./...               # Static analysis
go build ./...             # Verify compilation
```

Requires CGO + `librdkafka` for Kafka mode. Use `SIMULATOR_MODE=json` to develop without Kafka.
