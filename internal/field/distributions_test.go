package field

import (
	"math/rand"
	"testing"
	"time"
)

func TestGenRange(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genRange(10, 20)
	for i := 0; i < 100; i++ {
		v := fn(r, nil).(int)
		if v < 10 || v > 20 {
			t.Fatalf("range value out of bounds: %d", v)
		}
	}
}

func TestCompileNormalDistribution(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn, err := compileNormalDistribution(50.0, 10.0, floatPtr(0.0))
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		v := fn(r, nil).(float64)
		if v < 0 {
			t.Fatalf("normal value below min clamp: %f", v)
		}
	}
}

func TestCompilePoissonDistribution(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn, err := compilePoissonDistribution(3.5)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		v := fn(r, nil).(int)
		if v < 0 {
			t.Fatalf("poisson value negative: %d", v)
		}
	}
}

func floatPtr(f float64) *float64 {
	return &f
}
