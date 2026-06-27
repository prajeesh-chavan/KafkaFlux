package transport

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"go-kafka-simulator/internal/engine"
)

type FilePublisher struct {
	inChan     chan *engine.DataEvent
	sim        *engine.Simulator
	format     string // "json" or "csv"
	outputPath string
}

func NewFilePublisher(format string, outputPath string, inChan chan *engine.DataEvent) *FilePublisher {
	return &FilePublisher{
		format:     format,
		outputPath: outputPath,
		inChan:     inChan,
	}
}

func (fp *FilePublisher) SetSimulator(sim *engine.Simulator) {
	fp.sim = sim
}

func (fp *FilePublisher) Start(ctx context.Context, wg *sync.WaitGroup, parallelWorkers int) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		ext := ".json"
		if fp.format == "csv" {
			ext = ".csv"
		}
		
		fullPath := fp.outputPath + ext

		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("File Engine Error: Failed to create output directory %s: %v\n", dir, err)
			return
		}

		file, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("File Engine Error: Failed to open output file: %v\n", err)
			return
		}
		defer file.Close()

		var csvWriter *csv.Writer
		headerWritten := false
		var orderedKeys []string

		if fp.format == "csv" {
			csvWriter = csv.NewWriter(file)
			defer csvWriter.Flush()
		}

		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-fp.inChan:
				if !ok {
					return
				}

				if fp.format == "json" {
					_, _ = file.Write(event.Data)
					_, _ = file.WriteString("\n") 
				} else if fp.format == "csv" {
					var payload map[string]interface{}
					if err := json.Unmarshal(event.Data, &payload); err == nil {
						
						// 1. Establish static column definitions on the very first entry
						if !headerWritten {
							for k := range payload {
								orderedKeys = append(orderedKeys, k)
							}
							// Sort columns alphabetically to eliminate map sequence randomness completely!
							sort.Strings(orderedKeys) 
							
							_ = csvWriter.Write(orderedKeys)
							headerWritten = true
						}

						// 2. Iterate using the locked, sorted key sequence
						var row []string
						for _, key := range orderedKeys {
							val, exists := payload[key]
							if exists {
								row = append(row, fmt.Sprintf("%v", val))
							} else {
								row = append(row, "") 
							}
						}
						
						_ = csvWriter.Write(row)
						csvWriter.Flush()
					}
				}

				if fp.sim != nil {
					fp.sim.ReleaseBuffer(event.Data)
				}
			}
		}
	}()
}

func (fp *FilePublisher) Close() {}