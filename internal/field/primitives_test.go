package field

import (
	"math/rand"
	"strings"
	"testing"
	"time"
)

func TestGenUUID(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genUUID()
	v := fn(r, nil).(string)
	if len(v) != 36 {
		t.Fatalf("expected 36 chars, got %d: %s", len(v), v)
	}
	parts := strings.Split(v, "-")
	if len(parts) != 5 {
		t.Fatalf("expected 5 parts, got %d", len(parts))
	}
}

func TestGenInt(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genInt()
	for i := 0; i < 100; i++ {
		v := fn(r, nil).(int)
		if v < 1000 || v >= 91000 {
			t.Fatalf("int out of range: %d", v)
		}
	}
}

func TestGenFloat(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genFloat()
	for i := 0; i < 100; i++ {
		v := fn(r, nil).(float64)
		if v < 5.0 || v > 500.0 {
			t.Fatalf("float out of range: %f", v)
		}
	}
}

func TestGenTimestamp(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genTimestamp()
	v := fn(r, nil).(string)
	if _, err := time.Parse(time.RFC3339, v); err != nil {
		t.Fatalf("invalid timestamp: %s: %v", v, err)
	}
}
