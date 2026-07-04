# KafkaFlux Architecture Refactor Plan

## Goal

Refactor KafkaFlux from a monolithic codebase into a modular, testable, industry-practice architecture. The plan is organized as 6 sequential steps, each verifiable before moving to the next.

---

## Current architecture problems

| Problem | Location | Details |
|---------|----------|---------|
| God file | `internal/config/profile.go` (467 lines) | Mixes YAML models, 12+ field generators, weighted choice compiler, conditional parser, name data |
| Transport ↔ Engine coupling | `transport/` imports `engine/` | `SetSimulator(sim)` exists only to recycle byte buffers |
| Split config | `main.go` + env vars | `config.yaml` vs env vars — no single source of truth |
| Logic in entrypoint | `cmd/producer/main.go` | Contains env parsing, mode selection, transport construction |
| Single-goroutine FilePublisher | `internal/transport/file.go` | All topics serialize through one goroutine |
| String-parsed YAML | All profile YAMLs | Expressions parsed from strings are brittle |
| Duplicated types | `cmd/generator/main.go` | Defines its own model types instead of importing from config |

---

## Target architecture

```
internal/
├── app/                   # NEW: wires everything, env/config unification
│   └── app.go
├── config/
│   ├── config.go          # Config model + env loading
│   ├── loader.go          # YAML loading
│   └── profile.go         # Profile model only (trimmed)
├── field/                 # NEW: one file per generator group
│   ├── field.go           # Shared types: FieldGen, FieldConfig, etc.
│   ├── primitives.go      # uuid, int, float, timestamp
│   ├── person.go          # first_name, last_name, name, email, phone
│   ├── distributions.go   # range, normal, poisson
│   ├── composite.go       # conditional, weighted, pool
│   └── compile.go         # Compile orchestration
├── engine/
│   ├── event.go           # DataEvent type
│   ├── simulator.go       # Core orchestrator
│   ├── worker.go          # runWorker
│   ├── scaler.go          # getTrafficScale
│   └── dashboard.go       # Terminal UI
├── transport/
│   ├── publisher.go       # DataPublisher interface (no engine dependency)
│   ├── kafka.go
│   └── file.go
├── pool/                  # NEW: buffer pool (shared, no circular deps)
│   └── buffer.go
└── registry/              # StateRegistry (already separated from engine)
    └── state.go
```

---

## Step 1: Extract `internal/field/`

**What**: Move all generator logic out of `config/profile.go` into a new `field` package.

**New files**:

| File | Contents |
|------|----------|
| `internal/field/field.go` | Shared types: `FieldGen`, `PoolFetcher`, `FieldConfig` (new YAML model) |
| `internal/field/primitives.go` | `uuid`, `int`, `float`, `timestamp` generators |
| `internal/field/person.go` | `first_name`, `last_name`, `name`, `email`, `phone` + name data |
| `internal/field/distributions.go` | `range`, `normal_distribution`, `poisson_distribution` |
| `internal/field/composite.go` | `conditional`, `weighted`, `pool` |
| `internal/field/compile.go` | `Compile` — orchestrates field compilation |

**Files modified**:

| File | Change |
|------|--------|
| `internal/config/profile.go` | Remove all generator functions, name data, compilation logic. Keep only `EntityProfile`, `ChaosConfig`, `LoadProfiles` wrapper |
| `cmd/generator/main.go` | Import config types instead of duplicating |

**After**: `config/profile.go` drops from ~467 to ~80 lines.

---

## Step 2: Change profile YAML format

Transition from string-parsed `value` entries to structured `type`-based config.

### Old format (per field)

```yaml
field_name:
  - value: "expression"
    weight: 100
    state_action: publish
    state_pool: pool_name
```

### New format (per field)

```yaml
# Primitives
order_id:
  type: uuid
  publish_to: orders

order_number:
  type: int

total_amount:
  type: float

created_at:
  type: timestamp

# Person fields
first_name:
  type: first_name

last_name:
  type: last_name

full_name:
  type: name

email:
  type: email

phone:
  type: phone

# Distributions
discount:
  type: range
  min: 0
  max: 25

shipping:
  type: normal
  mean: 15.0
  stddev: 3.5
  min: 0.0

items_count:
  type: poisson
  lambda: 3.0

# References
customer_id:
  type: pool
  pool: customers

# Weighted enum
order_status:
  type: weighted
  values:
    CREATED: 20
    CONFIRMED: 20
    PROCESSING: 25
    COMPLETED: 30
    CANCELLED: 5

# Conditional
completed_at:
  type: conditional
  rules:
    - when: order_status == COMPLETED
      then:
        type: timestamp
  default: null

# Chaos config moves inline (already in EntityProfile)
chaos:
  drop_percentage: 1.0
  corrupt_fields:
    email:
      rate: 5.0
```

**Changes**: `EntityProfile.Fields` type changes from `map[string][]ProfileWeightedItem` to `map[string]FieldConfig`. All 10 profile YAMLs updated. Generator updated to produce new format.

---

## Step 3: Extract `internal/app/`

Move wiring logic from `main.go` into `internal/app/app.go`.

**`cmd/producer/main.go` becomes**:
```go
package main
import "go-kafka-simulator/internal/app"
func main() { app.Run() }
```

**Config unification**: Single `Config` struct populated from YAML base + env var overrides + defaults. All `os.Getenv` calls removed from `main.go` and `transport/*`.

**`app.Run()` handles**: config loading, profile loading, mode/transport selection, channel creation, lifecycle (ctx, wg, signal handling).

---

## Step 4: Decouple transport from engine

**New file**: `internal/pool/buffer.go`

```go
type BufferPool interface {
    Get() []byte
    Put([]byte)
}
```

**Changes**:
- `DataPublisher` removes `SetSimulator(sim *engine.Simulator)`, adds `SetBufferPool(pool pool.BufferPool)`
- `KafkaPublisher` / `FilePublisher` use `BufferPool` instead of `sim.ReleaseBuffer`
- Engine holds its own `BufferPool` and passes it to publisher

---

## Step 5: Split `engine/` into focused files

| New file | Extracted from `simulator.go` |
|----------|-------------------------------|
| `event.go` | `DataEvent` type |
| `worker.go` | `runWorker` method |
| `scaler.go` | `getTrafficScale` function |
| `dashboard.go` | `StartDashboard` method |

`simulator.go` keeps: `Simulator` struct, `NewSimulator`, `Start`, `ReleaseBuffer`.

---

## Step 6: Docker adjustments

After all steps, verify:
- `go build ./...` passes
- `CGO_ENABLED=1 go build -tags musl -ldflags="-w -s" -o kafka-producer ./cmd/producer/main.go` passes
- `docker compose up` works

Update `Dockerfile` and `docker-compose.yml` if any file paths changed.
