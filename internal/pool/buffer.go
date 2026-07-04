package pool

import "sync"

type BufferPool interface {
	Get() []byte
	Put([]byte)
}

type SyncPool struct {
	pool sync.Pool
}

func NewSyncPool() *SyncPool {
	return &SyncPool{
		pool: sync.Pool{
			New: func() interface{} { return make([]byte, 0, 1024) },
		},
	}
}

func (p *SyncPool) Get() []byte {
	return p.pool.Get().([]byte)
}

func (p *SyncPool) Put(b []byte) {
	p.pool.Put(b)
}
