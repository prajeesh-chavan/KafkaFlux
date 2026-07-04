package transport

import (
	"context"
	"fmt"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"go-kafka-simulator/internal/engine"
	"go-kafka-simulator/internal/pool"
)

type KafkaPublisher struct {
	producer *kafka.Producer
	inChan   chan *engine.DataEvent
	bufPool  pool.BufferPool
}

func NewKafkaPublisher(brokers string, inChan chan *engine.DataEvent) (*KafkaPublisher, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":            brokers,
		"acks":                         "1",
		"compression.type":             "snappy",
		"linger.ms":                    "20",
		"queue.buffering.max.messages": "500000",
	})
	if err != nil {
		return nil, err
	}

	return &KafkaPublisher{
		producer: p,
		inChan:   inChan,
	}, nil
}

func (kp *KafkaPublisher) SetBufferPool(p pool.BufferPool) {
	kp.bufPool = p
}

func (kp *KafkaPublisher) Start(ctx context.Context, wg *sync.WaitGroup, parallelWorkers int) {
	go func() {
		for e := range kp.producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition.Error)
				}
				if kp.bufPool != nil && ev.Opaque != nil {
					if buf, ok := ev.Opaque.([]byte); ok {
						kp.bufPool.Put(buf)
					}
				}
			}
		}
	}()

	for i := 0; i < parallelWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case event, ok := <-kp.inChan:
					if !ok {
						return
					}
					_ = kp.producer.Produce(&kafka.Message{
						TopicPartition: kafka.TopicPartition{Topic: &event.Topic, Partition: kafka.PartitionAny},
						Value:          event.Data,
						Opaque:         event.Data,
					}, nil)
				}
			}
		}()
	}
}

func (kp *KafkaPublisher) Close() {
	kp.producer.Flush(15000)
	kp.producer.Close()
}
