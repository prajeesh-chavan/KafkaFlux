package field

import (
	"math/rand"
	"strings"
	"testing"
	"time"
)

func TestGenStreet(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genStreet()
	v := fn(r, nil).(string)
	if !strings.Contains(v, " ") {
		t.Fatalf("invalid street: %s", v)
	}
}

func TestGenCity(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genCity()
	v := fn(r, map[string]interface{}{"country_code": "US"}).(string)
	if v == "" {
		t.Fatal("city should not be empty")
	}
}

func TestGenCityDefault(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genCity()
	v := fn(r, nil).(string)
	if v == "" {
		t.Fatal("city should not be empty even without country_code")
	}
}

func TestGenState(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genState()
	v := fn(r, map[string]interface{}{"country_code": "US"}).(string)
	if len(v) != 2 {
		t.Fatalf("expected 2-letter state, got %s", v)
	}
}

func TestGenZipCode(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genZipCode()
	v := fn(r, map[string]interface{}{"country_code": "US"}).(string)
	if len(v) != 5 {
		t.Fatalf("expected 5-digit zip, got %s", v)
	}
}

func TestGenCountry(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genCountry()
	v := fn(r, map[string]interface{}{"country_code": "IN"}).(string)
	if v != "India" {
		t.Fatalf("expected India, got %s", v)
	}
}

func TestGenFullAddress(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genFullAddress()
	v := fn(r, nil).(string)
	if len(v) < 10 {
		t.Fatalf("address too short: %s", v)
	}
}
