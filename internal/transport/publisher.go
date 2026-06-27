package transport

import (
	"context"
	"go-kafka-simulator/internal/engine"
	"sync"
)

// DataPublisher unifies Kafka and file-based workers
type DataPublisher interface {
	Start(ctx context.Context, wg *sync.WaitGroup, parallelWorkers int)
	SetSimulator(sim *engine.Simulator)
	Close()
}