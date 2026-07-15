package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Simulator SimulatorConfig `yaml:"simulator"`
}

type SimulatorConfig struct {
	Workers      int      `yaml:"workers"`
	ProfilesDir  string   `yaml:"profiles_dir"`
	Profiles     []string `yaml:"profiles"`
	KafkaServers string   `yaml:"kafka_servers"`
	MetricsPort  int      `yaml:"metrics_port"`
	LogLevel     string   `yaml:"log_level"`
	Seed         int64    `yaml:"seed"`
	BatchSize    int64    `yaml:"batch_size"`
}

type RuntimeConfig struct {
	Config
	Mode       string
	Broker     string
	OutputPath string
	Profiles   []string
	Seed       int64
	BatchSize  int64
}

func LoadRuntime() (*RuntimeConfig, error) {
	cfg, err := Load("config.yaml")
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