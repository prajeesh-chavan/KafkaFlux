package engine

import (
	"testing"
)

func TestDataEventCreation(t *testing.T) {
	e := DataEvent{
		Topic: "test.topic",
		Data:  []byte(`{"key":"value"}`),
	}
	if e.Topic != "test.topic" {
		t.Fatalf("unexpected topic: %s", e.Topic)
	}
	if string(e.Data) != `{"key":"value"}` {
		t.Fatalf("unexpected data: %s", string(e.Data))
	}
}
