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
	inChan       chan *engine.DataEvent
	sim          *engine.Simulator
	format       string // "json" or "csv"
	outputDir    string // Changed from outputPath to outputDir
	
	// Track file handles, writers, and headers per topic dynamically
	files         map[string]*os.File
	csvWriters    map[string]*csv.Writer
	orderedKeys   map[string][]string
	headersReady  map[string]bool
	mu            sync.Mutex 
}

func NewFilePublisher(format string, outputDir string, inChan chan *engine.DataEvent) *FilePublisher {
	return &FilePublisher{
		format:       format,
		outputDir:    outputDir,
		inChan:       inChan,
		files:        make(map[string]*os.File),
		csvWriters:   make(map[string]*csv.Writer),
		orderedKeys:  make(map[string][]string),
		headersReady: make(map[string]bool),
	}
}

func (fp *FilePublisher) SetSimulator(sim *engine.Simulator) {
	fp.sim = sim
}

func (fp *FilePublisher) Start(ctx context.Context, wg *sync.WaitGroup, parallelWorkers int) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Ensure baseline output directory exists safely
		if err := os.MkdirAll(fp.outputDir, 0755); err != nil {
			fmt.Printf("File Engine Error: Failed to create output directory %s: %v\n", fp.outputDir, err)
			return
		}

		ext := ".json"
		if fp.format == "csv" {
			ext = ".csv"
		}

		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-fp.inChan:
				if !ok {
					return
				}

				topic := event.Topic
				if topic == "" {
					topic = "unknown_topic"
				}

				// Thread-safely get or create the specific file handle for this topic
				fp.mu.Lock()
				file, exists := fp.files[topic]
				if !exists {
					fullPath := filepath.Join(fp.outputDir, topic+ext)
					var err error
					file, err = os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
					if err != nil {
						fmt.Printf("File Engine Error: Failed to open file for topic %s: %v\n", topic, err)
						fp.mu.Unlock()
						continue
					}
					fp.files[topic] = file

					if fp.format == "csv" {
						fp.csvWriters[topic] = csv.NewWriter(file)
					}
				}
				fp.mu.Unlock()

				// Write processing block
				if fp.format == "json" {
					_, _ = file.Write(event.Data)
					_, _ = file.WriteString("\n")
				} else if fp.format == "csv" {
					var payload map[string]interface{}
					if err := json.Unmarshal(event.Data, &payload); err == nil {
						writer := fp.csvWriters[topic]

						fp.mu.Lock()
						// Initialize specific structured headers for this file topic
						if !fp.headersReady[topic] {
							var keys []string
							for k := range payload {
								keys = append(keys, k)
							}
							sort.Strings(keys)
							fp.orderedKeys[topic] = keys
							
							_ = writer.Write(keys)
							fp.headersReady[topic] = true
						}

						// Iterate strictly using this topic's key sequence mapping layout
						var row []string
						for _, key := range fp.orderedKeys[topic] {
							val, exists := payload[key]
							if exists {
								row = append(row, fmt.Sprintf("%v", val))
							} else {
								row = append(row, "")
							}
						}
						fp.mu.Unlock()

						_ = writer.Write(row)
						writer.Flush()
					}
				}

				if fp.sim != nil {
					fp.sim.ReleaseBuffer(event.Data)
				}
			}
		}
	}()
}

func (fp *FilePublisher) Close() {
	fp.mu.Lock()
	defer fp.mu.Unlock()

	// Flush and close every open file handle smoothly on exit shutdown signals
	for topic, writer := range fp.csvWriters {
		writer.Flush()
		_ = topic
	}
	for _, file := range fp.files {
		_ = file.Sync()
		_ = file.Close()
	}
	fmt.Println("[TRANSPORT] All topic file sinks safely flushed and offline.")
}