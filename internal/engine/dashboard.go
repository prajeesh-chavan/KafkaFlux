package engine

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

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
