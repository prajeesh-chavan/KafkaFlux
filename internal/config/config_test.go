package config

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })

	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	return tmpFile.Name()
}

func TestLoadConfig(t *testing.T) {
	content := `simulator:
  workers: 4
  profiles_dir: "./test_profiles"
  kafka_servers: "localhost:9092"
  metrics_port: 9999
  log_level: "debug"
`
	path := writeTempConfig(t, content)

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Simulator.Workers != 4 {
		t.Fatalf("expected 4 workers, got %d", cfg.Simulator.Workers)
	}
	if cfg.Simulator.ProfilesDir != "./test_profiles" {
		t.Fatalf("unexpected profiles_dir: %s", cfg.Simulator.ProfilesDir)
	}
	if cfg.Simulator.KafkaServers != "localhost:9092" {
		t.Fatalf("unexpected kafka_servers: %s", cfg.Simulator.KafkaServers)
	}
	if cfg.Simulator.MetricsPort != 9999 {
		t.Fatalf("expected metrics_port 9999, got %d", cfg.Simulator.MetricsPort)
	}
	if cfg.Simulator.LogLevel != "debug" {
		t.Fatalf("expected log_level debug, got %s", cfg.Simulator.LogLevel)
	}
}

func TestLoadConfigInvalidFile(t *testing.T) {
	_, err := Load("./nonexistent.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestLoadRuntimeDefaults(t *testing.T) {
	content := `simulator:
  workers: 8
  profiles_dir: "./profiles"
  kafka_servers: "kafka:29092"
  metrics_port: 9099
  log_level: "info"
`
	path := writeTempConfig(t, content)

	os.Setenv("SIMULATOR_MODE", "json")
	os.Setenv("KAFKA_BROKERS", "localhost:9092")
	os.Setenv("OUTPUT_FILE_PATH", "./output")
	t.Cleanup(func() {
		os.Unsetenv("SIMULATOR_MODE")
		os.Unsetenv("KAFKA_BROKERS")
		os.Unsetenv("OUTPUT_FILE_PATH")
	})

	// Use a modified Load that always reads our temp file
	rc, err := loadRuntimeFrom(path)
	if err != nil {
		t.Fatal(err)
	}
	if rc.Mode != "json" {
		t.Fatalf("expected json mode, got %s", rc.Mode)
	}
	if rc.Broker != "localhost:9092" {
		t.Fatalf("expected localhost:9092, got %s", rc.Broker)
	}
	if rc.OutputPath != "./output" {
		t.Fatalf("expected ./output, got %s", rc.OutputPath)
	}
}

func loadRuntimeFrom(configPath string) (*RuntimeConfig, error) {
	cfg, err := Load(configPath)
	if err != nil {
		return nil, err
	}

	mode := os.Getenv("SIMULATOR_MODE")
	if mode == "" {
		mode = "kafka"
	}

	broker := os.Getenv("KAFKA_BROKERS")
	if broker == "" {
		broker = cfg.Simulator.KafkaServers
	}

	outputPath := os.Getenv("OUTPUT_FILE_PATH")
	if outputPath == "" {
		outputPath = "./data_output"
	}

	rc := &RuntimeConfig{
		Config:     *cfg,
		Mode:       mode,
		Broker:     broker,
		OutputPath: outputPath,
		Profiles:   cfg.Simulator.Profiles,
		Seed:       cfg.Simulator.Seed,
		BatchSize:  cfg.Simulator.BatchSize,
	}

	envProfiles := os.Getenv("PROFILES")
	if envProfiles != "" {
		parsed := strings.Split(envProfiles, ",")
		for i := range parsed {
			parsed[i] = strings.TrimSpace(parsed[i])
		}
		rc.Profiles = parsed
	}

	if envSeed := os.Getenv("SIMULATOR_SEED"); envSeed != "" {
		if parsed, err := strconv.ParseInt(envSeed, 10, 64); err == nil {
			rc.Seed = parsed
		}
	}

	if envBatch := os.Getenv("BATCH_SIZE"); envBatch != "" {
		if parsed, err := strconv.ParseInt(envBatch, 10, 64); err == nil {
			rc.BatchSize = parsed
		}
	}

	return rc, nil
}

func TestLoadRuntimeProfilesEnvVar(t *testing.T) {
	content := `simulator:
  workers: 8
  profiles_dir: "./profiles"
  kafka_servers: "kafka:29092"
  metrics_port: 9099
  log_level: "info"
`
	path := writeTempConfig(t, content)

	os.Setenv("PROFILES", "orders,payments")
	t.Cleanup(func() {
		os.Unsetenv("PROFILES")
	})

	rc, err := loadRuntimeFrom(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(rc.Profiles) != 2 {
		t.Fatalf("expected 2 profiles from env, got %d", len(rc.Profiles))
	}
	if rc.Profiles[0] != "orders" {
		t.Fatalf("expected orders, got %s", rc.Profiles[0])
	}
	if rc.Profiles[1] != "payments" {
		t.Fatalf("expected payments, got %s", rc.Profiles[1])
	}
}
