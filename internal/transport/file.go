package transport

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"go-kafka-simulator/internal/engine"
	"go-kafka-simulator/internal/pool"
	"go-kafka-simulator/internal/telemetry"
)

type FilePublisher struct {
	inChan    chan *engine.DataEvent
	bufPool   pool.BufferPool
	metrics   *telemetry.Metrics
	format    string
	outputDir string

	files        map[string]*os.File
	csvWriters   map[string]*csv.Writer
	orderedKeys  map[string][]string
	headersReady map[string]bool
	mu           sync.Mutex
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

func (fp *FilePublisher) SetBufferPool(p pool.BufferPool) {
	fp.bufPool = p
}

func (fp *FilePublisher) SetMetrics(m *telemetry.Metrics) {
	fp.metrics = m
}

func (fp *FilePublisher) Start(ctx context.Context, wg *sync.WaitGroup, parallelWorkers int) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := os.MkdirAll(fp.outputDir, 0755); err != nil {
			slog.Error("failed to create output directory", "dir", fp.outputDir, "error", err)
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

				fp.mu.Lock()
				file, exists := fp.files[topic]
				if !exists {
					fullPath := filepath.Join(fp.outputDir, topic+ext)
					var err error
					file, err = os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
					if err != nil {
						slog.Error("failed to open file for topic", "topic", topic, "error", err)
						fp.mu.Unlock()
						continue
					}
					fp.files[topic] = file

					if fp.format == "csv" {
						fp.csvWriters[topic] = csv.NewWriter(file)
					}
				}
				fp.mu.Unlock()

				if fp.format == "json" {
					_, _ = file.Write(event.Data)
					_, _ = file.WriteString("\n")
				} else if fp.format == "csv" {
					var payload map[string]interface{}
					if err := json.Unmarshal(event.Data, &payload); err == nil {
						writer := fp.csvWriters[topic]

						fp.mu.Lock()
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

				if fp.bufPool != nil {
					fp.bufPool.Put(event.Data)
				}
			}
		}
	}()
}

func (fp *FilePublisher) Close() {
	fp.mu.Lock()
	defer fp.mu.Unlock()

	for topic, writer := range fp.csvWriters {
		writer.Flush()
		_ = topic
	}
	for _, file := range fp.files {
		_ = file.Sync()
		_ = file.Close()
	}
	slog.Info("file publisher closed")
}
