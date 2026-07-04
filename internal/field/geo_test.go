package field

import (
	"math/rand"
	"strings"
	"testing"
	"time"
)

func TestGenLatitude(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genLatitude()
	for i := 0; i < 100; i++ {
		v := fn(r, nil).(float64)
		if v < -90 || v > 90 {
			t.Fatalf("latitude out of range: %f", v)
		}
	}
}

func TestGenLongitude(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genLongitude()
	for i := 0; i < 100; i++ {
		v := fn(r, nil).(float64)
		if v < -180 || v > 180 {
			t.Fatalf("longitude out of range: %f", v)
		}
	}
}

func TestGenCoordinatePair(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genCoordinatePair()
	v := fn(r, nil).(string)
	parts := strings.Split(v, ",")
	if len(parts) != 2 {
		t.Fatalf("expected 2 coordinates, got %s", v)
	}
}

func TestGenTimezone(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genTimezone()
	v := fn(r, nil).(string)
	if !strings.Contains(v, "/") {
		t.Fatalf("invalid timezone: %s", v)
	}
}
