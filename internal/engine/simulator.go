package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go-kafka-simulator/internal/config"
)

type DataEvent struct {
	Topic string
	Data  []byte
}

type Simulator struct {
	profiles      []*config.EntityProfile
	outChan       chan *DataEvent
	bytePool      *sync.Pool
	Registry      *StateRegistry
	StartTime     time.Time
	EventCounters map[string]*uint64 
	CurrentEPS    map[string]*uint64 
}

func NewSimulator(profiles []*config.EntityProfile, outChan chan *DataEvent) *Simulator {
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

func getTrafficScale(startTime time.Time) float64 {
	duration := time.Since(startTime)
	const periodSeconds = 600.0 
	seconds := math.Mod(duration.Seconds(), periodSeconds)
	radians := (2.0 * math.Pi * seconds / periodSeconds) - (math.Pi / 2.0)
	scale := 1.0 + (0.6 * math.Sin(radians))
	if scale < 0.1 {
		return 0.1 
	}
	return scale
}

func (s *Simulator) ReleaseBuffer(buf []byte) {
	s.bytePool.Put(buf)
}

func (s *Simulator) runWorker(ctx context.Context, wg *sync.WaitGroup, prof *config.EntityProfile) {
	defer wg.Done()

	localRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	startTime := time.Now()

	var currentEPS float64
	var interval time.Duration

	adjustTicker := time.NewTicker(2 * time.Second)
	defer adjustTicker.Stop()

	currentEPS = float64(prof.TargetEPS)
	interval = time.Second / time.Duration(currentEPS)

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
			// Chaos Injection: Drop events
			if prof.Chaos.DropPercentage > 0 && localRand.Float64()*100.0 < prof.Chaos.DropPercentage {
				time.Sleep(interval)
				continue
			}

			loopStart := time.Now()

			buf := s.bytePool.Get().([]byte)
			buf = buf[:0]

			payload := make(map[string]interface{})
			payload["__registry"] = s.Registry

			for _, fieldOrder := range prof.Compiled {
				val := fieldOrder.Gen(localRand, payload)
				
				// Chaos Injection: Field Corruption
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

				if rules, ok := prof.Fields[fieldOrder.Name]; ok && len(rules) == 1 {
					if rules[0].StateAction == "publish" && rules[0].StatePool != "" {
						s.Registry.Publish(rules[0].StatePool, fmt.Sprintf("%v", val))
					}
				}
			}

			delete(payload, "__registry")

			data, err := json.Marshal(payload)
			if err == nil {
				buf = append(buf, data...)
				s.outChan <- &DataEvent{
					Topic: prof.Topic,
					Data:  buf,
				}
				atomic.AddUint64(s.EventCounters[prof.Entity], 1)
				atomic.StoreUint64(s.CurrentEPS[prof.Entity], uint64(currentEPS))
			} else {
				s.bytePool.Put(buf)
			}

			elapsed := time.Since(loopStart)
			if elapsed < interval {
				time.Sleep(interval - elapsed)
			}
		}
	}
}

func (s *Simulator) StartDashboard(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		mode := os.Getenv("SIMULATOR_MODE")
		if mode == "" {
			mode = "kafka"
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				fmt.Print("\033[2J\033[1;1H")

				uptime := time.Since(s.StartTime).Round(time.Second)
				chanLen := len(s.outChan)
				chanCap := cap(s.outChan)
				chanPercent := int((float64(chanLen) / float64(chanCap)) * 100)

				barSize := 20
				filledSize := int((float64(chanLen) / float64(chanCap)) * float64(barSize))
				bar := ""
				for i := 0; i < barSize; i++ {
					if i < filledSize {
						bar += "█"
					} else {
						bar += "."
					}
				}

				fmt.Println("======================================================================")
				fmt.Println("     KAFKAFLUX ENTERPRISE EVENT STREAM SIMULATOR")
				fmt.Println("======================================================================")
				fmt.Printf(" System Uptime: %s | Profiles: %d | Transport: %s\n", uptime, len(s.profiles), strings.ToUpper(mode))
				fmt.Printf(" Buffer Channel Load: [%s] %d%% (%d / %d)\n", bar, chanPercent, chanLen, chanCap)
				fmt.Println("----------------------------------------------------------------------")
				fmt.Printf("%-15s %-30s %-12s %-15s\n", "ENTITY", "TOPIC", "CURR_EPS", "TOTAL_EVENTS")
				fmt.Println("----------------------------------------------------------------------")

				for _, prof := range s.profiles {
					total := atomic.LoadUint64(s.EventCounters[prof.Entity])
					eps := atomic.LoadUint64(s.CurrentEPS[prof.Entity])
					
					waveIndicator := ""
					if prof.DynamicScaling {
						waveIndicator = "[Dynamic]"
					}

					fmt.Printf("%-15s %-30s %-12d %-15d%s\n", 
						prof.Entity, 
						prof.Topic, 
						eps, 
						total,
						waveIndicator,
					)
				}
				fmt.Println("----------------------------------------------------------------------")
				fmt.Println(" [Press Ctrl+C to safely flush buffers and exit system metrics]")
			}
		}
	}()
}