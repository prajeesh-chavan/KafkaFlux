package engine

import (
	"testing"
	"time"
)

func TestGetTrafficScale(t *testing.T) {
	now := time.Now()
	scale := getTrafficScale(now)
	if scale < 0.1 {
		t.Fatalf("scale below minimum: %f", scale)
	}
	if scale > 1.6 {
		t.Fatalf("scale above maximum: %f", scale)
	}
}

func TestGetTrafficScaleAfterPeriod(t *testing.T) {
	start := time.Now().Add(-601 * time.Second)
	scale := getTrafficScale(start)
	if scale < 0.1 || scale > 1.6 {
		t.Fatalf("scale out of range after period: %f", scale)
	}
}
