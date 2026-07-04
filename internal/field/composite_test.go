package field

import (
	"math/rand"
	"testing"
	"time"
)

func TestCompileWeightedChoice(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := compileWeightedChoice(map[string]float64{"A": 50, "B": 30, "C": 20})
	counts := map[string]int{}
	for i := 0; i < 10000; i++ {
		v := fn(r, nil).(string)
		counts[v]++
	}
	if counts["A"] == 0 || counts["B"] == 0 || counts["C"] == 0 {
		t.Fatalf("not all values selected: %v", counts)
	}
}

func TestGenPool(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genPool("testpool")
	reg := &mockPoolFetcher{vals: map[string][]string{
		"testpool": {"val1", "val2"},
	}}
	v := fn(r, map[string]interface{}{"__registry": reg}).(string)
	if v != "val1" && v != "val2" {
		t.Fatalf("unexpected pool value: %s", v)
	}
}

func TestGenPoolFallback(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genPool("empty")
	v := fn(r, nil).(string)
	if v == "" {
		t.Fatal("pool fallback should generate a value")
	}
}

func TestCompileConditional(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	cfg := FieldConfig{
		Type: "conditional",
		Rules: []ConditionalRule{
			{
				When: "status == ACTIVE",
				Then: &FieldConfig{Type: "boolean"},
			},
		},
		Default: &FieldConfig{Value: "N/A"},
	}
	fn, err := compileConditional(cfg)
	if err != nil {
		t.Fatal(err)
	}
	v := fn(r, map[string]interface{}{"status": "ACTIVE"})
	if _, ok := v.(bool); !ok {
		t.Fatalf("expected bool for ACTIVE status, got %T: %v", v, v)
	}
	v2 := fn(r, map[string]interface{}{"status": "INACTIVE"})
	if v2 != "N/A" {
		t.Fatalf("expected N/A for INACTIVE, got %v", v2)
	}
}

type mockPoolFetcher struct {
	vals map[string][]string
}

func (m *mockPoolFetcher) Fetch(poolName string) (string, bool) {
	pool, ok := m.vals[poolName]
	if !ok || len(pool) == 0 {
		return "", false
	}
	return pool[0], true
}
