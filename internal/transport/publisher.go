package transport

import (
	"context"
	"sync"

	"go-kafka-simulator/internal/pool"
)

type DataPublisher interface {
	Start(ctx context.Context, wg *sync.WaitGroup, parallelWorkers int)
	SetBufferPool(p pool.BufferPool)
	Close()
}
