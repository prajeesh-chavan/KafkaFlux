package transport

import (
	"context"
	"sync"

	"go-kafka-simulator/internal/pool"
	"go-kafka-simulator/internal/telemetry"
)

type DataPublisher interface {
	Start(ctx context.Context, wg *sync.WaitGroup, parallelWorkers int)
	SetBufferPool(p pool.BufferPool)
	SetMetrics(m *telemetry.Metrics)
	Close()
}
