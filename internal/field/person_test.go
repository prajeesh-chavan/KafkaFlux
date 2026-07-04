package field

import (
	"math/rand"
	"strings"
	"testing"
	"time"
)

func TestGenFirstName(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genFirstName()
	v := fn(r, nil).(string)
	if v == "" {
		t.Fatal("first name should not be empty")
	}
}

func TestGenLastName(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genLastName()
	v := fn(r, nil).(string)
	if v == "" {
		t.Fatal("last name should not be empty")
	}
}

func TestGenName(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genName()
	v := fn(r, nil).(string)
	parts := strings.Fields(v)
	if len(parts) < 2 {
		t.Fatalf("expected at least 2 words, got: %s", v)
	}
}

func TestGenEmail(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genEmail()
	state := map[string]interface{}{
		"first_name": "Amit",
		"last_name":  "Sharma",
	}
	v := fn(r, state).(string)
	if !strings.Contains(v, "@") {
		t.Fatalf("invalid email: %s", v)
	}
}

func TestGenPhone(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genPhone()
	v := fn(r, nil).(string)
	if len(v) != 10 {
		t.Fatalf("expected 10 digits, got %d: %s", len(v), v)
	}
	if v[0] != '9' && v[0] != '8' && v[0] != '7' {
		t.Fatalf("invalid starting digit: %c", v[0])
	}
}
