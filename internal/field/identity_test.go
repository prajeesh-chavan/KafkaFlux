package field

import (
	"math/rand"
	"strings"
	"testing"
	"time"
)

func TestGenBoolean(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genBoolean()
	for i := 0; i < 100; i++ {
		v := fn(r, nil).(bool)
		if v != true && v != false {
			t.Fatal("boolean must be true or false")
		}
	}
}

func TestGenCompany(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genCompany()
	v := fn(r, nil).(string)
	if v == "" {
		t.Fatal("company should not be empty")
	}
}

func TestGenCompanyEmail(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genCompanyEmail()
	v := fn(r, nil).(string)
	if !strings.Contains(v, "@") {
		t.Fatalf("invalid email: %s", v)
	}
}

func TestGenJobTitle(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genJobTitle()
	v := fn(r, nil).(string)
	if v == "" {
		t.Fatal("job title should not be empty")
	}
}

func TestGenIP(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genIP()
	v := fn(r, nil).(string)
	parts := strings.Split(v, ".")
	if len(parts) != 4 {
		t.Fatalf("expected 4 octets, got %d", len(parts))
	}
}

func TestGenIPv6(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genIPv6()
	v := fn(r, nil).(string)
	parts := strings.Split(v, ":")
	if len(parts) != 8 {
		t.Fatalf("expected 8 groups, got %d", len(parts))
	}
}

func TestGenUserAgent(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genUserAgent()
	v := fn(r, nil).(string)
	if v == "" {
		t.Fatal("user agent should not be empty")
	}
}

func TestGenMAC(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genMAC()
	v := fn(r, nil).(string)
	parts := strings.Split(v, ":")
	if len(parts) != 6 {
		t.Fatalf("expected 6 octets, got %d", len(parts))
	}
}

func TestGenCreditCard(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genCreditCard()
	v := fn(r, nil).(string)
	parts := strings.Split(v, " ")
	if len(parts) != 4 {
		t.Fatalf("expected 4 groups, got %d", len(parts))
	}
}

func TestGenHexColor(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genHexColor()
	v := fn(r, nil).(string)
	if len(v) != 7 || v[0] != '#' {
		t.Fatalf("invalid hex color: %s", v)
	}
}

func TestGenSSN(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genSSN()
	v := fn(r, nil).(string)
	parts := strings.Split(v, "-")
	if len(parts) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(parts))
	}
}

func TestGenCurrency(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genCurrency()
	v := fn(r, nil).(string)
	if len(v) != 3 {
		t.Fatalf("expected 3-letter code, got %s", v)
	}
}

func TestGenHTTPStatus(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genHTTPStatus()
	for i := 0; i < 100; i++ {
		v := fn(r, nil).(int)
		if v < 100 || v > 599 {
			t.Fatalf("invalid status code: %d", v)
		}
	}
}

func TestGenURL(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fn := genURL()
	v := fn(r, nil).(string)
	if !strings.HasPrefix(v, "http") {
		t.Fatalf("invalid URL: %s", v)
	}
}
