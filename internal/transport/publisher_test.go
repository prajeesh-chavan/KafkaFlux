package transport

import (
	"context"
	"sync"
	"testing"

	"go-kafka-simulator/internal/engine"
	"go-kafka-simulator/internal/pool"
	"go-kafka-simulator/internal/telemetry"
)

type mockPublisher struct {
	started    bool
	closed     bool
	bufPoolSet bool
	metricsSet bool
}

func (m *mockPublisher) Start(_ context.Context, _ *sync.WaitGroup, _ int) {
	m.started = true
}

func (m *mockPublisher) SetBufferPool(_ pool.BufferPool) {
	m.bufPoolSet = true
}

func (m *mockPublisher) SetMetrics(_ *telemetry.Metrics) {
	m.metricsSet = true
}

func (m *mockPublisher) Close() {
	m.closed = true
}

func TestPublisherInterface(t *testing.T) {
	var pub DataPublisher = &mockPublisher{}
	if pub == nil {
		t.Fatal("DataPublisher should be assignable")
	}
}

func TestNewFilePublisher(t *testing.T) {
	ch := make(chan *engine.DataEvent, 10)
	fp := NewFilePublisher("json", "./test_out", ch)
	if fp == nil {
		t.Fatal("FilePublisher should not be nil")
	}
	if fp.format != "json" {
		t.Fatalf("expected json format, got %s", fp.format)
	}
}

func TestFilePublisherSetBufferPool(t *testing.T) {
	ch := make(chan *engine.DataEvent, 10)
	fp := NewFilePublisher("json", "./test_out", ch)
	bp := pool.NewSyncPool()
	fp.SetBufferPool(bp)
	if fp.bufPool == nil {
		t.Fatal("buffer pool should be set")
	}
}

func TestFilePublisherClose(t *testing.T) {
	ch := make(chan *engine.DataEvent, 10)
	fp := NewFilePublisher("json", "./test_out", ch)
	fp.Close()
}

func TestDataEventChannel(t *testing.T) {
	ch := make(chan *engine.DataEvent, 1)
	ch <- &engine.DataEvent{Topic: "test", Data: []byte(`{}`)}
	close(ch)

	evt := <-ch
	if evt.Topic != "test" {
		t.Fatalf("unexpected topic: %s", evt.Topic)
	}
}
