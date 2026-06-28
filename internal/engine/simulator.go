package engine

import (
	"context"
	"encoding/json"
	"go-kafka-simulator/internal/config"
	"math/rand"
	"sync"
	"time"
	"fmt"
)

type DataEvent struct {
	Topic string
	Data  []byte
}

type Simulator struct {
	profiles []*config.EntityProfile
	outChan  chan *DataEvent
	bytePool *sync.Pool
	Registry *StateRegistry
}

func NewSimulator(profiles []*config.EntityProfile, outChan chan *DataEvent) *Simulator {
	return &Simulator{
		profiles: profiles,
		outChan:  outChan,
		Registry: NewStateRegistry(),
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

					payload := make(map[string]interface{})
					// Inject the registry into the payload map context safely
					payload["__registry"] = s.Registry 

					// Evaluate schema sequence while maintaining states
					for _, fieldOrder := range prof.Compiled {
						// 1. Generate the field value
						val := fieldOrder.Gen(localRand, payload)
						payload[fieldOrder.Name] = val

						// 2. Dynamic Capture: Check if this specific field wants to publish to a pool
						if rules, ok := prof.Fields[fieldOrder.Name]; ok && len(rules) == 1 {
							if rules[0].StateAction == "publish" && rules[0].StatePool != "" {
								s.Registry.Publish(rules[0].StatePool, fmt.Sprintf("%v", val))
							}
						}
					}

					// Clean up the context pointer before encoding to JSON output
					delete(payload, "__registry")

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