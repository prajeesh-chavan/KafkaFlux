<div align="center">
  <h1>KafkaFlux</h1>
  <p><strong>Enterprise Event Stream Simulator for Kafka</strong></p>
  <p>
    <a href="#quick-start">Quick Start</a> •
    <a href="#architecture">Architecture</a> •
    <a href="#data-model">Data Model</a> •
    <a href="#features">Features</a> •
    <a href="#running">Running</a> •
    <a href="#profiles">Profiles</a> •
    <a href="#metrics">Metrics</a>
  </p>
  <p>
    <img src="https://img.shields.io/badge/Go-1.25-00ADD8?logo=go" alt="Go 1.25">
    <img src="https://img.shields.io/badge/Kafka-7.5.0-231F20?logo=apachekafka" alt="Kafka 7.5.0">
    <img src="https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker" alt="Docker Compose">
    <img src="https://img.shields.io/badge/Prometheus-Metrics-E6522C?logo=prometheus" alt="Prometheus">
  </p>
  <p>Built by <a href="https://linkedin.com/in/prajeeshchavan"><strong>Prajeesh Chavan</strong></a></p>
</div>

---

KafkaFlux generates **realistic, high-volume event streams** to Kafka topics (or JSON/CSV files). It simulates a complete ecommerce data platform — orders, customers, payments, clickstream, inventory, shipments, and 24+ other entities — with cross-referenced IDs, weighted distributions, conditional logic, chaos injection, and dynamic traffic scaling.

Built for **load testing**, **demo environments**, **data pipeline development**, and **portfolio demonstrations**.

---

## Quick Start

```sh
git clone <your-repo-url>
cd KafkaFlux
docker compose up
```

That's it. Zookeeper + Kafka + the simulator start together. You'll immediately see:

```
======================================================================
     KAFKAFLUX ENTERPRISE EVENT STREAM SIMULATOR
======================================================================
 System Uptime: 12s | Profiles: 8 | Transport: KAFKA
 Buffer Channel Load: [████................] 21% (21212 / 100000)
----------------------------------------------------------------------
ENTITY           TOPIC                          CURR_EPS     TOTAL_EVENTS
----------------------------------------------------------------------
customers        telemetry.ecommerce.customers   10           112
orders           telemetry.ecommerce.orders      50           524
payments         telemetry.ecommerce.payments    45           478
products         telemetry.ecommerce.products    20           202
shipments        telemetry.ecommerce.shipments   20           201
...
```

Open another terminal:

```sh
curl localhost:9099/              # JSON status dashboard
curl localhost:9099/metrics       # Prometheus metrics
```

**Without Kafka (JSON output):**

```sh
SIMULATOR_MODE=json go run ./cmd/producer/main.go
```

---

## Architecture

```
                         ┌─────────────────────────────────┐
                         │         config.yaml             │
                         │  workers, profiles_dir, broker  │
                         └────────────┬────────────────────┘
                                      │ LoadRuntime()
                                      ▼
┌────────────────────────────────────────────────────────────────────┐
│                        app.Run()                                   │
│                                                                    │
│  ┌──────────────┐     ┌──────────────────┐   ┌──────────────────┐  │
│  │ config.Load  │     │ field.InitData   │   │ telemetry.New    │  │
│  │ Profiles()   │ ──> │ Loader()         │   │ Metrics()        │  │
│  │ (30 YAMLs)   │     │ (10 JSON files)  │   │                  │  │
│  └──────┬───────┘     └──────────────────┘   └────────┬─────────┘  │
│         │                                             │            │
│         ▼                                             ▼            │
│  ┌────────────────────────────────────────────────────────────┐    │
│  │                                                            │    │  
│  │  ┌──────────┐  ┌───────────┐  ┌──────────┐  ┌──────────┐   │    │
│  │  │ Worker 1 │  │ Worker 2  │  │ Worker 3 │  │ Worker N │   │    │
│  │  │(orders)  │  │(customers)│  │(payments)│  │(clickst..│   │    │
│  │  │ 50 EPS   │  │ 10 EPS    │  │ 45 EPS   │  │ 200 EPS  │   │    │
│  │  └────┬─────┘  └────┬──────┘  └─────┬────┘  └─────┬────┘   │    │
│  │       │             │               │             │        │    │
│  │       └─────────────┴───────────────┴─────────────┘        │    │
│  │                          │                                 │    │
│  │                    ┌─────▼──────┐                          │    │
│  │                    │   Buffer   │  100,000 events          │    │
│  │                    │  Channel   │  (buffered channel)      │    │
│  │                    └─────┬──────┘                          │    │
│  └──────────────────────────┼─────────────────────────────────┘    │
│                             │                                      │
│  ┌──────────────────────────▼──────────────────────────────────┐   │
│  │              transport.DataPublisher                        │   │
│  │  ┌──────────────────┐    ┌──────────────────┐               │   │
│  │  │ KafkaPublisher   │    │ FilePublisher    │               │   │
│  │  │ (N workers)      │    │ (JSON / CSV)     │               │   │
│  │  └────────┬─────────┘    └────────┬─────────┘               │   │
│  └───────────┼──────────────────────┼──────────────────────────┘   │
│              │                      │                              │
│              ▼                      ▼                              │
│         Kafka Topic           data_output/*.{json,csv}             │
│                                                                    │
│  ┌───────────────────────────────────────────────────────────┐     │
│  │  telemetry                                                │     │
│  │  ┌────────────────┐  ┌────────────────┐                   │     │
│  │  │  Prometheus    │  │  JSON Status   │                   │     │
│  │  │  /metrics      │  │                │                   │     │
│  │  └────────────────┘  └────────────────┘                   │     │
│  │                                                           │     │
│  │  pool.BufferPool (sync.Pool for byte slice reuse)         │     │
│  │  engine.StateRegistry (cross-profile ID reference pools)  │     │
│  │  engine.Dashboard (real-time terminal UI, 500ms tick)     │     │
│  └───────────────────────────────────────────────────────────┘     │
└────────────────────────────────────────────────────────────────────┘
```

### Package Map

| Package | Purpose |
|---------|---------|
| `cmd/producer` | Entrypoint — 7 lines, calls `app.Run()` |
| `cmd/generator` | CLI tool for generating/validating profile YAMLs |
| `internal/app` | Wiring: config, profiles, publisher, simulator, metrics, lifecycle |
| `internal/config` | YAML config loading, env override, profile compilation |
| `internal/engine` | Simulator, workers, state pools, sine-wave scaler, dashboard |
| `internal/field` | 50+ field generators + `CompileField` dispatcher |
| `internal/pool` | `BufferPool` interface + `SyncPool` for byte slice reuse |
| `internal/telemetry` | Structured logging (`slog`) + Prometheus/JSON metrics HTTP server |
| `internal/transport` | `DataPublisher` interface + Kafka / File implementations |

---

## Data Model

KafkaFlux models a **complete retail ecommerce data platform** with 30+ entities organized into reference and transactional layers.

```
REFERENCE ENTITIES                    TRANSACTIONAL ENTITIES
┌───────────────┐                     ┌───────────────────┐
│ brands        │                     │ orders            │───┐
│ categories    │─────┐               │ order_items       │   │
│ vendors       │     │               │ payments          │───┤
│ warehouses    │     ├──────┐        │ shipments         │   │
│ carriers      │─────┘      │        │ returns           │   │
│ products      │────────────┘        │ refunds           │◄──┘
│ customers     │                     │ product_reviews   │
│ addresses     │                     │ customer_events   │
│ sales_channels│                     │ carts / cart_items│
└───────────────┘                     └───────────────────┘
```

### Entity-to-Topic Mapping

| Profile | Kafka Topic | EPS | Scaling |
|---------|-------------|-----|---------|
| `customers` | `telemetry.ecommerce.customers` | 10 | Static |
| `orders` | `telemetry.ecommerce.orders` | 50 | **Dynamic** |
| `payments` | `telemetry.ecommerce.payments` | 45 | Static |
| `products` | `telemetry.ecommerce.products` | 20 | Static |
| `shipments` | `telemetry.ecommerce.shipments` | 20 | Static |
| `customer_events` | `telemetry.ecommerce.customer_events` | 200 | Static |
| `inventory` | `telemetry.ecommerce.inventory` | 20 | Static |
| `product_reviews` | `telemetry.ecommerce.product_reviews` | 15 | Static |
| *(plus 22 more reference entities)* | | | |

Entities share IDs through **state pools** — when `orders` publishes an `order_id`, the `payments` profile can reference it via `type: pool, pool: orders`.

### Sample Profile: `profiles/orders.yaml`

```yaml
entity: orders
topic: "telemetry.ecommerce.orders"
target_eps: 50
dynamic_scaling: true

chaos:
  drop_percentage: 1.0     # 1% chance to drop entire event
  corrupt_fields: {}

fields:
  order_id:
    type: uuid
    publish_to: orders      # ← Publish to state pool for cross-refs

  customer_id:
    type: pool              # ← Fetch from customers pool
    pool: customers

  order_status:
    type: weighted          # ← Probability-weighted enum
    values:
      PLACED: 15
      CONFIRMED: 15
      PROCESSING: 20
      SHIPPED: 20
      DELIVERED: 25
      CANCELLED: 4
      RETURNED: 1

  tax_amount:
    type: normal            # ← Gaussian distribution
    mean: 15.0
    stddev: 5.0
    min: 1.0

  completed_at:
    type: conditional       # ← Only generated when condition met
    rules:
      - when: order_status == DELIVERED
        then:
          type: timestamp
```

---

## Features

### 50+ Field Generators

| Category | Types |
|----------|-------|
| **Primitives** | `uuid`, `int`, `float`, `timestamp`, `boolean` |
| **Identity** | `first_name`, `last_name`, `name`, `full_name`, `email`, `phone`, `username`, `password`, `ssn` |
| **Business** | `company`, `company_email`, `job_title`, `currency`, `language` |
| **Location** | `street`, `city`, `state`, `zip`, `country`, `country_code`, `full_address` |
| **Geo** | `latitude`, `longitude`, `coordinate_pair`, `timezone` |
| **Network** | `ip`, `ipv6`, `mac`, `user_agent`, `http_method`, `http_status`, `mime_type` |
| **Finance** | `credit_card` |
| **Ecommerce** | `sku`, `product_name`, `url` |
| **Text** | `word`, `sentence`, `paragraph`, `lorem ipsum` |
| **Time** | `past_timestamp`, `future_timestamp`, `date`, `birth_date` |
| **Distributions** | `range` (uniform), `normal` (Gaussian), `poisson` |
| **Composite** | `weighted` (CDF-based enum), `pool` (cross-ref fetch), `conditional`, `regex`, literal values |

### Dynamic Traffic Scaling

Traffic follows a **sine wave** over a 10-minute period, ranging from **0.1x to 1.6x** of the target EPS — simulating realistic ecommerce traffic patterns (morning ramp, lunch peak, evening taper):

```go
// internal/engine/scaler.go
func getTrafficScale(startTime time.Time) float64 {
    duration := time.Since(startTime)
    seconds := math.Mod(duration.Seconds(), 600.0)
    radians := (2.0 * math.Pi * seconds / 600.0) - (math.Pi / 2.0)
    scale := 1.0 + (0.6 * math.Sin(radians))
    if scale < 0.1 { return 0.1 }
    return scale  // range: [0.1, 1.6]
}
```

### Chaos Engineering

- **Event drops**: Per-profile `drop_percentage` randomly discards events
- **Field corruption**: Per-field `rate` replaces values with `NULL` or `CHAOS_CORRUPTION_ERR`
- Use case: Test how downstream pipelines handle incomplete/corrupted data

### Conditional Fields

Fields are evaluated **only when a condition is met**, keeping generated data logically consistent:

```yaml
failure_code:
  type: conditional
  rules:
    - when: payment_status == FAILED
      then:
        type: weighted
        values:
          INSUFFICIENT_FUNDS: 35
          CARD_DECLINED: 25
```

### Real-Time Dashboard

```
======================================================================
     KAFKAFLUX ENTERPRISE EVENT STREAM SIMULATOR
======================================================================
 System Uptime: 2m35s | Profiles: 8 | Transport: KAFKA
 Buffer Channel Load: ████████░░░░░░░░░░░░ 42% (42345 / 100000)
----------------------------------------------------------------------
ENTITY           TOPIC                          CURR_EPS     TOTAL_EVENTS
----------------------------------------------------------------------
customers        telemetry.ecommerce.customers   10           15,342
orders           telemetry.ecommerce.orders      78           87,231 [Dynamic]
payments         telemetry.ecommerce.payments    45           77,834
...
```

### Prometheus Metrics

Exposed at `http://localhost:9099/metrics`:

```
kafkaflux_events_total{entity="orders"} 87231
kafkaflux_events_dropped_total 1234
kafkaflux_delivery_failures_total 0
kafkaflux_current_eps{entity="orders"} 78.345
kafkaflux_buffer_fill 42345
kafkaflux_buffer_cap 100000
kafkaflux_uptime_seconds 155
```

### JSON Status Endpoint

`GET http://localhost:9099/` returns structured JSON with per-entity metrics.

---

## Running

### Full Stack (Docker — Recommended)

```sh
docker compose up
```

Starts:
- **Zookeeper** (port 2181)
- **Kafka** (port 9092)
- **Producer** with live dashboard (port 9099)

Mounts `./profiles/`, `./data/`, `./config.yaml` from host — edit profiles without rebuilding.

### Without Kafka (JSON / CSV)

```sh
# JSON files (one file per topic)
SIMULATOR_MODE=json go run ./cmd/producer/main.go

# CSV files
SIMULATOR_MODE=csv OUTPUT_FILE_PATH=./data_output go run ./cmd/producer/main.go
```

### Profile Generator CLI

```sh
go run ./cmd/generator                      # Interactive mode
go run ./cmd/generator --help               # Field type reference
go run ./cmd/generator --init orders         # Template profile
go run ./cmd/generator --validate profiles/  # Validate profiles
```

### Configuration

`config.yaml`:
```yaml
simulator:
  workers: 8              # Publisher goroutines
  profiles_dir: "./profiles"
  kafka_servers: "kafka:29092"
  metrics_port: 9099
  log_level: "info"
```

Environment variable overrides:

| Variable | Overrides | Default |
|----------|-----------|---------|
| `SIMULATOR_MODE` | Transport mode | `kafka` |
| `KAFKA_BROKERS` | Kafka servers | `kafka:29092` |
| `OUTPUT_FILE_PATH` | Output directory | `./data_output` |

---

## Profiles

30 entity profiles in `profiles/` covering a complete retail schema:

```
addresses.yaml          inventory.yaml              order_status_history.yaml   returns.yaml
brand.yaml              inventory_movements.yaml     orders.yaml                 sales_channels.yaml
carriers.yaml           order_addresses.yaml         payments.yaml               shipment_items.yaml
cart_items.yaml         order_item_discounts.yaml    product_reviews.yaml        shipments.yaml
carts.yaml              order_items.yaml             product_variants.yaml       vendor_addresses.yaml
categories.yaml                                     products.yaml               vendors.yaml
customer_addresses.yaml                             promotions.yaml             warehouses.yaml
customer_events.yaml    refunds.yaml
customers.yaml          return_items.yaml
```

---

## Development

```sh
go test -count=1 ./...     # 57 tests across 14 test files
go vet ./...               # Static analysis
go build ./...             # Verify compilation
```

**Note**: Kafka mode requires CGO + `librdkafka`. Use `SIMULATOR_MODE=json` for local development without Kafka.

---

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Language | Go 1.25 |
| Kafka Client | `confluent-kafka-go/v2` |
| Config | YAML (via `gopkg.in/yaml.v3`) |
| Orchestration | Docker Compose |
| Metrics | Prometheus text format (hand-rolled) |
| Logging | `log/slog` (structured, leveled) |
| CI/CD | GitHub Actions (build + test + vet) |

---

## Author

**Prajeesh Chavan** — Data Engineering

[![LinkedIn](https://img.shields.io/badge/LinkedIn-prajeeshchavan-0A66C2?logo=linkedin)](https://linkedin.com/in/prajeeshchavan)
[![GitHub](https://img.shields.io/badge/GitHub-prajeeshchavan-181717?logo=github)](https://github.com/prajeeshchavan)

Built for learning, demonstration, and helping fellow data engineers build better pipelines.