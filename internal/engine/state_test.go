package engine

import (
	"testing"
)

func TestNewStateRegistry(t *testing.T) {
	sr := NewStateRegistry()
	if sr == nil {
		t.Fatal("StateRegistry should not be nil")
	}
}

func TestPublishAndFetch(t *testing.T) {
	sr := NewStateRegistry()
	sr.Publish("testpool", "value1")
	sr.Publish("testpool", "value2")

	v, ok := sr.Fetch("testpool")
	if !ok {
		t.Fatal("expected to find value in pool")
	}
	if v != "value1" && v != "value2" {
		t.Fatalf("unexpected value: %s", v)
	}
}

func TestFetchEmptyPool(t *testing.T) {
	sr := NewStateRegistry()
	_, ok := sr.Fetch("nonexistent")
	if ok {
		t.Fatal("expected false for nonexistent pool")
	}
}

func TestFetchAfterCap(t *testing.T) {
	sr := NewStateRegistry()
	for i := 0; i < 10010; i++ {
		sr.Publish("bigpool", "val")
	}
	_, ok := sr.Fetch("bigpool")
	if !ok {
		t.Fatal("pool should still have values after cap")
	}
}
