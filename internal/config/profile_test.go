package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeProfile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

const baseProfile = `entity: test_entity
topic: "telemetry.test"
target_eps: 10
dynamic_scaling: false

chaos:
  drop_percentage: 0.0
  corrupt_fields: {}

fields:
  id:
    type: uuid
  name:
    type: first_name
`

func TestLoadProfiles(t *testing.T) {
	dir, err := os.MkdirTemp("", "profiles-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	writeProfile(t, dir, "test.yaml", baseProfile)

	profiles, err := LoadProfiles(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(profiles))
	}
	if profiles[0].Entity != "test_entity" {
		t.Fatalf("unexpected entity: %s", profiles[0].Entity)
	}
	if len(profiles[0].Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(profiles[0].Fields))
	}
	if len(profiles[0].Compiled) != 2 {
		t.Fatalf("expected 2 compiled fields, got %d", len(profiles[0].Compiled))
	}
}

func TestLoadProfilesEmptyDir(t *testing.T) {
	dir, err := os.MkdirTemp("", "empty-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	profiles, err := LoadProfiles(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 0 {
		t.Fatalf("expected 0 profiles, got %d", len(profiles))
	}
}

func TestLoadProfilesInvalidYAML(t *testing.T) {
	dir, err := os.MkdirTemp("", "bad-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	os.WriteFile(filepath.Join(dir, "bad.yaml"), []byte("invalid: [yaml: broken"), 0644)

	_, err = LoadProfiles(dir, nil)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadProfilesSubdirectory(t *testing.T) {
	dir, err := os.MkdirTemp("", "subdir-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	if err := os.MkdirAll(filepath.Join(dir, "ecommerce"), 0755); err != nil {
		t.Fatal(err)
	}
	writeProfile(t, dir, "ecommerce/orders.yaml", baseProfile)

	profiles, err := LoadProfiles(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile from subdirectory, got %d", len(profiles))
	}
	if profiles[0].Entity != "test_entity" {
		t.Fatalf("unexpected entity: %s", profiles[0].Entity)
	}
}

func TestLoadProfilesDisabled(t *testing.T) {
	dir, err := os.MkdirTemp("", "disabled-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	writeProfile(t, dir, "enabled.yaml", baseProfile)

	disabledContent := `entity: disabled_entity
enabled: false
topic: "telemetry.test"
target_eps: 10
dynamic_scaling: false

chaos:
  drop_percentage: 0.0
  corrupt_fields: {}

fields:
  id:
    type: uuid
`
	writeProfile(t, dir, "disabled.yaml", disabledContent)

	profiles, err := LoadProfiles(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected 1 enabled profile, got %d", len(profiles))
	}
	if profiles[0].Entity != "test_entity" {
		t.Fatalf("expected test_entity, got %s", profiles[0].Entity)
	}
}

func TestLoadProfilesFilterByEntityName(t *testing.T) {
	dir, err := os.MkdirTemp("", "filter-name-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	writeProfile(t, dir, "orders.yaml", `entity: orders
topic: "telemetry.orders"
target_eps: 10
chaos:
  drop_percentage: 0.0
  corrupt_fields: {}
fields:
  id:
    type: uuid
`)
	writeProfile(t, dir, "payments.yaml", `entity: payments
topic: "telemetry.payments"
target_eps: 10
chaos:
  drop_percentage: 0.0
  corrupt_fields: {}
fields:
  id:
    type: uuid
`)

	profiles, err := LoadProfiles(dir, []string{"orders"})
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected 1 filtered profile, got %d", len(profiles))
	}
	if profiles[0].Entity != "orders" {
		t.Fatalf("expected orders, got %s", profiles[0].Entity)
	}
}

func TestLoadProfilesFilterByGlob(t *testing.T) {
	dir, err := os.MkdirTemp("", "filter-glob-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	if err := os.MkdirAll(filepath.Join(dir, "ecommerce"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "iot"), 0755); err != nil {
		t.Fatal(err)
	}

	writeProfile(t, dir, "ecommerce/orders.yaml", `entity: orders
topic: "telemetry.orders"
target_eps: 10
chaos:
  drop_percentage: 0.0
  corrupt_fields: {}
fields:
  id:
    type: uuid
`)
	writeProfile(t, dir, "ecommerce/payments.yaml", `entity: payments
topic: "telemetry.payments"
target_eps: 10
chaos:
  drop_percentage: 0.0
  corrupt_fields: {}
fields:
  id:
    type: uuid
`)
	writeProfile(t, dir, "iot/sensors.yaml", `entity: sensors
topic: "telemetry.sensors"
target_eps: 10
chaos:
  drop_percentage: 0.0
  corrupt_fields: {}
fields:
  id:
    type: uuid
`)

	profiles, err := LoadProfiles(dir, []string{"ecommerce/*"})
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 2 {
		t.Fatalf("expected 2 ecommerce profiles, got %d", len(profiles))
	}
}

func TestLoadProfilesFilterDisabledOverriddenByFilter(t *testing.T) {
	dir, err := os.MkdirTemp("", "filter-override-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	writeProfile(t, dir, "orders.yaml", `entity: orders
enabled: false
topic: "telemetry.orders"
target_eps: 10
chaos:
  drop_percentage: 0.0
  corrupt_fields: {}
fields:
  id:
    type: uuid
`)

	profiles, err := LoadProfiles(dir, []string{"orders"})
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile (filter overrides disabled), got %d", len(profiles))
	}
	if profiles[0].Entity != "orders" {
		t.Fatalf("expected orders, got %s", profiles[0].Entity)
	}
}
