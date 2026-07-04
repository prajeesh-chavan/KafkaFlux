package field

import (
	"math/rand"
	"strings"
	"testing"
	"time"
)

func TestGenWord(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genWord()
	v := fn(r, nil).(string)
	if v == "" {
		t.Fatal("word should not be empty")
	}
}

func TestGenSentence(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genSentence()
	v := fn(r, nil).(string)
	if !strings.HasSuffix(v, ".") {
		t.Fatalf("sentence should end with period: %s", v)
	}
}

func TestGenParagraph(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genParagraph()
	v := fn(r, nil).(string)
	if len(v) < 20 {
		t.Fatalf("paragraph too short: %s", v)
	}
}

func TestGenProductName(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genProductName()
	v := fn(r, nil).(string)
	if !strings.Contains(v, " v") {
		t.Fatalf("product name should contain version: %s", v)
	}
}

func TestGenSKU(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genSKU()
	v := fn(r, nil).(string)
	parts := strings.Split(v, "-")
	if len(parts) != 3 {
		t.Fatalf("expected 3 parts, got %d in %s", len(parts), v)
	}
}

func TestGenPastTimestamp(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genPastTimestamp()
	v := fn(r, nil).(string)
	ts, err := time.Parse(time.RFC3339, v)
	if err != nil {
		t.Fatalf("invalid timestamp: %s: %v", v, err)
	}
	if !ts.Before(time.Now()) {
		t.Fatal("past_timestamp should be in the past")
	}
}

func TestGenFutureTimestamp(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genFutureTimestamp()
	v := fn(r, nil).(string)
	ts, err := time.Parse(time.RFC3339, v)
	if err != nil {
		t.Fatalf("invalid timestamp: %s: %v", v, err)
	}
	if !ts.After(time.Now()) {
		t.Fatal("future_timestamp should be in the future")
	}
}

func TestGenDate(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genDate()
	v := fn(r, nil).(string)
	if _, err := time.Parse("2006-01-02", v); err != nil {
		t.Fatalf("invalid date: %s: %v", v, err)
	}
}
