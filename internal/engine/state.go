package engine

import (
	"math/rand"
	"sync"
	"time"
)

type StateRegistry struct {
	mu    sync.RWMutex
	pools map[string][]string
}

func NewStateRegistry() *StateRegistry {
	return &StateRegistry{
		pools: make(map[string][]string),
	}
}

// Publish adds a generated value to a specific reference pool
func (sr *StateRegistry) Publish(poolName string, value string) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	sr.pools[poolName] = append(sr.pools[poolName], value)

	// Cap memory leak safety buffer per pool to the latest 10,000 values
	if len(sr.pools[poolName]) > 10000 {
		sr.pools[poolName] = sr.pools[poolName][1:]
	}
}

// Fetch retrieves a random existing value from a reference pool
func (sr *StateRegistry) Fetch(poolName string) (string, bool) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	pool, exists := sr.pools[poolName]
	if !exists || len(pool) == 0 {
		return "", false
	}

	// Safe pseudo-random index selection
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	return pool[r.Intn(len(pool))], true
}