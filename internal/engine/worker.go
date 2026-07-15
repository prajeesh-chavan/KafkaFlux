package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"go-kafka-simulator/internal/config"
	"go-kafka-simulator/internal/field"
)

func (s *Simulator) runWorker(ctx context.Context, wg *sync.WaitGroup, prof *config.EntityProfile, idx int) {
	defer wg.Done()

	var seed int64
	if s.seed != 0 {
		seed = s.seed + int64(idx)
	} else {
		seed = time.Now().UnixNano()
	}
	localRand := rand.New(rand.NewSource(seed))
	startTime := time.Now()

	var currentEPS float64
	var interval time.Duration

	adjustTicker := time.NewTicker(2 * time.Second)
	defer adjustTicker.Stop()

	currentEPS = float64(prof.TargetEPS)
	interval = time.Second / time.Duration(currentEPS)

	batchSize := prof.BatchSize
	if batchSize == 0 {
		batchSize = s.batchSize
	}
	var produced int64

	for {
		select {
		case <-ctx.Done():
			return

		case <-adjustTicker.C:
			if prof.DynamicScaling {
				scale := getTrafficScale(startTime)
				newEPS := float64(prof.TargetEPS) * scale
				if int(newEPS) != int(currentEPS) && newEPS > 0 {
					currentEPS = newEPS
					interval = time.Second / time.Duration(currentEPS)
				}
			} else {
				if int(currentEPS) != prof.TargetEPS {
					currentEPS = float64(prof.TargetEPS)
					interval = time.Second / time.Duration(currentEPS)
				}
			}

		default:
			if prof.Chaos.DropPercentage > 0 && localRand.Float64()*100.0 < prof.Chaos.DropPercentage {
				if s.metrics != nil {
					s.metrics.IncEventsDropped()
				}
				time.Sleep(interval)
				continue
			}

			loopStart := time.Now()

			buf := s.bufPool.Get()
			buf = buf[:0]

			payload := make(map[string]interface{})
			payload["__registry"] = s.Registry
			payload["__data"] = field.GetDataLoader()

			for _, fieldOrder := range prof.Compiled {
				val := fieldOrder.Gen(localRand, payload)

				if corruptCfg, exists := prof.Chaos.CorruptFields[fieldOrder.Name]; exists {
					if localRand.Float64()*100.0 < corruptCfg.Rate {
						if localRand.Float64() > 0.5 {
							val = "NULL"
						} else {
							val = "CHAOS_CORRUPTION_ERR"
						}
					}
				}

				payload[fieldOrder.Name] = val

				if cfg, ok := prof.Fields[fieldOrder.Name]; ok {
					if cfg.PublishTo != "" {
						s.Registry.Publish(cfg.PublishTo, toString(val))
					}
				}
			}

			delete(payload, "__registry")
			delete(payload, "__data")
			delete(payload, "__name")

			data, err := json.Marshal(payload)
			if err == nil {
				buf = append(buf, data...)
				s.outChan <- &DataEvent{
					Topic: prof.Topic,
					Data:  buf,
				}
				atomic.AddUint64(s.EventCounters[prof.Entity], 1)
				atomic.StoreUint64(s.CurrentEPS[prof.Entity], uint64(currentEPS))
				if s.metrics != nil {
					s.metrics.IncEventsTotal(prof.Entity)
					s.metrics.SetEPS(prof.Entity, currentEPS)
				}
				produced++
				if batchSize > 0 && produced >= batchSize {
					slog.Debug("batch complete", "entity", prof.Entity, "count", produced)
					return
				}
			} else {
				s.bufPool.Put(buf)
				if s.metrics != nil {
					s.metrics.IncMarshalErrors()
				}
				slog.Error("json marshal failed", "entity", prof.Entity, "error", err)
			}

			elapsed := time.Since(loopStart)
			if elapsed < interval {
				time.Sleep(interval - elapsed)
			}
		}
	}
}

func toString(v interface{}) string {
	return fmt.Sprintf("%v", v)
}
