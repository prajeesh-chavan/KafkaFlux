package engine

import (
	"context"
	"encoding/json"
	"go-kafka-simulator/internal/config"
	"math/rand"
	"sync"
	"time"
	"fmt"
	"math"
	"sync/atomic"
	"os"
	"strings"
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
	StartTime    time.Time
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

// getTrafficScale computes a sinusoidal multiplier based on uptime.
// Emulates a full 24-hour business cycle every 10 minutes in real-time.
func getTrafficScale(startTime time.Time) float64 {
	duration := time.Since(startTime)
	
	// Period = 10 minutes (600 seconds) for an accelerated virtual day
	const periodSeconds = 600.0 
	seconds := math.Mod(duration.Seconds(), periodSeconds)
	
	// Shift by pi/2 so the system boots at a neutral middle baseline (Scale 1.0)
	radians := (2.0 * math.Pi * seconds / periodSeconds) - (math.Pi / 2.0)
	
	// Sine ranges from -1 to +1. Amplitude of 0.6 means scale bounces between 0.4 and 1.6
	scale := 1.0 + (0.6 * math.Sin(radians))
	
	if scale < 0.1 {
		return 0.1 // Baseline floor barrier safety limit
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

	// Track dynamic ticker evaluation intervals
	var currentEPS float64
	var interval time.Duration

	// Velocity controller evaluator timer running every 2 seconds
	adjustTicker := time.NewTicker(2 * time.Second)
	defer adjustTicker.Stop()

	// Initial configuration baseline setup
	currentEPS = float64(prof.TargetEPS)
	interval = time.Second / time.Duration(currentEPS)

	for {
		select {
		case <-ctx.Done():
			return

		case <-adjustTicker.C:
					// Only apply the wave calculation if dynamic_scaling is enabled in the YAML
					if prof.DynamicScaling {
						scale := getTrafficScale(startTime)
						newEPS := float64(prof.TargetEPS) * scale
						
						if int(newEPS) != int(currentEPS) && newEPS > 0 {
							currentEPS = newEPS
							interval = time.Second / time.Duration(currentEPS)
						}
					} else {
						// Fallback/Reset to exact baseline if it was toggled or defaults to flat
						if int(currentEPS) != prof.TargetEPS {
							currentEPS = float64(prof.TargetEPS)
							interval = time.Second / time.Duration(currentEPS)
						}
					}

		default:
			// Process event emission loop iteration matching dynamic velocity rates
			loopStart := time.Now()

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

			// Execution tracking time compensation window throttling control
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
		ticker := time.NewTicker(500 * time.Millisecond) // Refresh twice per second
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
				// ANSI Escape Codes: Clear screen (\033[H matches top left, \033[2J clears screen)
				fmt.Print("\033[2J\033[1;1H")

				uptime := time.Since(s.StartTime).Round(time.Second)
				chanLen := len(s.outChan)
				chanCap := cap(s.outChan)
				chanPercent := int((float64(chanLen) / float64(chanCap)) * 100)

				// Generate dynamic loading bar for ring-buffer
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
					
					// Visually show if traffic scaling is active
					waveIndicator := ""
					if prof.DynamicScaling {
						waveIndicator = " [Dynamic]"
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