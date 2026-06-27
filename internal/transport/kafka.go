package transport

import (
	"context"
	"fmt"
	"go-kafka-simulator/internal/engine"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaPublisher struct {
	producer *kafka.Producer
	inChan   chan *engine.DataEvent
	sim      *engine.Simulator
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

func (kp *KafkaPublisher) SetSimulator(sim *engine.Simulator) {
	kp.sim = sim
}

func (kp *KafkaPublisher) Start(ctx context.Context, wg *sync.WaitGroup, parallelWorkers int) {
	// Background Delivery Report Loop Listener
	go func() {
		for e := range kp.producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition.Error)
				}
				
				// Safely extract the original pristine buffer using Opaque metadata tracking
				if kp.sim != nil && ev.Opaque != nil {
					if originalBuf, ok := ev.Opaque.([]byte); ok {
						kp.sim.ReleaseBuffer(originalBuf)
					}
				}
			}
		}
	}()

	// Parallel Network I/O Pipeline Workers
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
					
					// Attach the original slice address to the Opaque interface pointer field
					_ = kp.producer.Produce(&kafka.Message{
						TopicPartition: kafka.TopicPartition{Topic: &event.Topic, Partition: kafka.PartitionAny},
						Value:          event.Data,
						Opaque:         event.Data, // Pass the reference forward through the CGO bounds safely
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