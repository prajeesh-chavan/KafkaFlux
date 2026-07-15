package engine

import (
	"context"
	"sync"
	"time"

	"go-kafka-simulator/internal/config"
	"go-kafka-simulator/internal/pool"
	"go-kafka-simulator/internal/telemetry"
)

type Simulator struct {
	profiles      []*config.EntityProfile
	outChan       chan *DataEvent
	bufPool       pool.BufferPool
	metrics       *telemetry.Metrics
	Registry      *StateRegistry
	StartTime     time.Time
	EventCounters map[string]*uint64
	CurrentEPS    map[string]*uint64
	seed          int64
	batchSize     int64
}

func NewSimulator(profiles []*config.EntityProfile, outChan chan *DataEvent, bufPool pool.BufferPool, metrics *telemetry.Metrics, seed, batchSize int64) *Simulator {
	counters := make(map[string]*uint64)
	epsTracker := make(map[string]*uint64)

	for _, prof := range profiles {
		var c uint64 = 0
		var e uint64 = 0
		counters[prof.Entity] = &c
		epsTracker[prof.Entity] = &e
	}

	return &Simulator{
		profiles:      profiles,
		outChan:       outChan,
		Registry:      NewStateRegistry(),
		StartTime:     time.Now(),
		EventCounters: counters,
		CurrentEPS:    epsTracker,
		bufPool:       bufPool,
		metrics:       metrics,
		seed:          seed,
		batchSize:     batchSize,
	}
}

func (s *Simulator) Start(ctx context.Context, wg *sync.WaitGroup) {
	for i, prof := range s.profiles {
		wg.Add(1)
		go s.runWorker(ctx, wg, prof, i)
	}
}
