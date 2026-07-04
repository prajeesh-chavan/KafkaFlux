package telemetry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type Metrics struct {
	mu             sync.Mutex
	eventsTotal    map[string]*int64
	eventsDropped  atomic.Int64
	deliveryFail   atomic.Int64
	marshalErrors  atomic.Int64
	bufferFill     atomic.Int64
	bufferCap      atomic.Int64
	eps            map[string]*atomic.Int64
	startTime      time.Time
}

func NewMetrics() *Metrics {
	return &Metrics{
		eventsTotal: make(map[string]*int64),
		eps:         make(map[string]*atomic.Int64),
		startTime:   time.Now(),
	}
}

func (m *Metrics) IncEventsTotal(entity string) {
	m.mu.Lock()
	ptr, ok := m.eventsTotal[entity]
	if !ok {
		var v int64
		ptr = &v
		m.eventsTotal[entity] = ptr
	}
	m.mu.Unlock()
	atomic.AddInt64(ptr, 1)
}

func (m *Metrics) IncEventsDropped() {
	m.eventsDropped.Add(1)
}

func (m *Metrics) IncDeliveryFail() {
	m.deliveryFail.Add(1)
}

func (m *Metrics) IncMarshalErrors() {
	m.marshalErrors.Add(1)
}

func (m *Metrics) SetBufferFill(fill, cap int) {
	m.bufferFill.Store(int64(fill))
	m.bufferCap.Store(int64(cap))
}

func (m *Metrics) SetEPS(entity string, eps float64) {
	m.mu.Lock()
	ptr, ok := m.eps[entity]
	if !ok {
		ptr = new(atomic.Int64)
		m.eps[entity] = ptr
	}
	m.mu.Unlock()
	ptr.Store(int64(eps * 1000))
}

func (m *Metrics) PrometheusHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		uptime := int64(time.Since(m.startTime).Seconds())

		fmt.Fprint(w, "# HELP kafkaflux_events_total Total events generated per entity\n")
		fmt.Fprint(w, "# TYPE kafkaflux_events_total counter\n")

		m.mu.Lock()
		entities := make([]string, 0, len(m.eventsTotal))
		for e := range m.eventsTotal {
			entities = append(entities, e)
		}
		sort.Strings(entities)
		for _, e := range entities {
			val := atomic.LoadInt64(m.eventsTotal[e])
			fmt.Fprintf(w, "kafkaflux_events_total{entity=%q} %d\n", e, val)
		}

		fmt.Fprint(w, "\n# HELP kafkaflux_events_dropped_total Events dropped by chaos injection\n")
		fmt.Fprint(w, "# TYPE kafkaflux_events_dropped_total counter\n")
		fmt.Fprintf(w, "kafkaflux_events_dropped_total %d\n", m.eventsDropped.Load())

		fmt.Fprint(w, "\n# HELP kafkaflux_delivery_failures_total Kafka delivery failures\n")
		fmt.Fprint(w, "# TYPE kafkaflux_delivery_failures_total counter\n")
		fmt.Fprintf(w, "kafkaflux_delivery_failures_total %d\n", m.deliveryFail.Load())

		fmt.Fprint(w, "\n# HELP kafkaflux_marshal_errors_total JSON marshal errors\n")
		fmt.Fprint(w, "# TYPE kafkaflux_marshal_errors_total counter\n")
		fmt.Fprintf(w, "kafkaflux_marshal_errors_total %d\n", m.marshalErrors.Load())

		fmt.Fprint(w, "\n# HELP kafkaflux_buffer_fill Buffer channel fill level\n")
		fmt.Fprint(w, "# TYPE kafkaflux_buffer_fill gauge\n")
		fmt.Fprintf(w, "kafkaflux_buffer_fill %d\n", m.bufferFill.Load())

		fmt.Fprint(w, "\n# HELP kafkaflux_buffer_cap Buffer channel capacity\n")
		fmt.Fprint(w, "# TYPE kafkaflux_buffer_cap gauge\n")
		fmt.Fprintf(w, "kafkaflux_buffer_cap %d\n", m.bufferCap.Load())

		fmt.Fprint(w, "\n# HELP kafkaflux_current_eps Current events per second per entity\n")
		fmt.Fprint(w, "# TYPE kafkaflux_current_eps gauge\n")
		for _, e := range entities {
			if ptr, ok := m.eps[e]; ok {
				fmt.Fprintf(w, "kafkaflux_current_eps{entity=%q} %.3f\n", e, float64(ptr.Load())/1000)
			}
		}

		fmt.Fprint(w, "\n# HELP kafkaflux_uptime_seconds System uptime\n")
		fmt.Fprint(w, "# TYPE kafkaflux_uptime_seconds counter\n")
		fmt.Fprintf(w, "kafkaflux_uptime_seconds %d\n", uptime)

		m.mu.Unlock()
	})
}

func (m *Metrics) StatusJSON() map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	entities := make([]string, 0, len(m.eventsTotal))
	for e := range m.eventsTotal {
		entities = append(entities, e)
	}
	sort.Strings(entities)

	perEntity := make([]map[string]interface{}, 0, len(entities))
	var totalEvents int64
	for _, e := range entities {
		ev := atomic.LoadInt64(m.eventsTotal[e])
		totalEvents += ev
		eps := float64(0)
		if ptr, ok := m.eps[e]; ok {
			eps = float64(ptr.Load()) / 1000
		}
		perEntity = append(perEntity, map[string]interface{}{
			"entity": e,
			"events": ev,
			"eps":    eps,
		})
	}

	return map[string]interface{}{
		"uptime_seconds":   int64(time.Since(m.startTime).Seconds()),
		"total_events":     totalEvents,
		"events_dropped":   m.eventsDropped.Load(),
		"delivery_failures": m.deliveryFail.Load(),
		"marshal_errors":   m.marshalErrors.Load(),
		"buffer_used":      m.bufferFill.Load(),
		"buffer_capacity":  m.bufferCap.Load(),
		"entities":         perEntity,
	}
}

func (m *Metrics) StatusHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m.StatusJSON())
	})
}
