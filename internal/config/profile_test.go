package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadProfiles(t *testing.T) {
	dir, err := os.MkdirTemp("", "profiles-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	profileContent := `entity: test_entity
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
	if err := os.WriteFile(filepath.Join(dir, "test.yaml"), []byte(profileContent), 0644); err != nil {
		t.Fatal(err)
	}

	profiles, err := LoadProfiles(dir)
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

	profiles, err := LoadProfiles(dir)
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

	_, err = LoadProfiles(dir)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}
