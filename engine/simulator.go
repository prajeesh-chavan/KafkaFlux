package engine

import (
	"context"
	"encoding/json"
	"go-kafka-simulator/internal/config"
	"math/rand"
	"sync"
	"time"
)

type DataEvent struct {
	Topic string
	Data  []byte
}

type Simulator struct {
	profiles []*config.EntityProfile
	outChan  chan *DataEvent
	bytePool *sync.Pool
}

func NewSimulator(profiles []*config.EntityProfile, outChan chan *DataEvent) *Simulator {
	return &Simulator{
		profiles: profiles,
		outChan:  outChan,
		bytePool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, 1024)
			},
		},
	}
}

func (s *Simulator) Start(ctx context.Context, wg *sync.WaitGroup) {
	for _, prof := range s.profiles {
		wg.Add(1)
		go s.runWorker(ctx, wg, prof)
	}
}

func (s *Simulator) runWorker(ctx context.Context, wg *sync.WaitGroup, prof *config.EntityProfile) {
	defer wg.Done()

	localRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	interval := time.Second / time.Duration(prof.TargetEPS)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			buf := s.bytePool.Get().([]byte)
			buf = buf[:0]

			// Evaluate schema sequence while continuously maintaining our calculation states
			payload := make(map[string]interface{})
			for _, fieldOrder := range prof.Compiled {
				payload[fieldOrder.Name] = fieldOrder.Gen(localRand, payload)
			}

			data, err := json.Marshal(payload)
			if err != nil {
				continue
			}
			buf = append(buf, data...)

			s.outChan <- &DataEvent{
				Topic: prof.Topic,
				Data:  buf,
			}
		}
	}
}

func (s *Simulator) ReleaseBuffer(buf []byte) {
	s.bytePool.Put(buf)
}