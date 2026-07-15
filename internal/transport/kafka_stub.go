//go:build !cgo

package transport

import (
	"errors"

	"go-kafka-simulator/internal/engine"
)

func NewKafkaPublisher(brokers string, inChan chan *engine.DataEvent) (DataPublisher, error) {
	return nil, errors.New("kafka transport requires CGO (use SIMULATOR_MODE=json or rebuild with CGO_ENABLED=1)")
}
