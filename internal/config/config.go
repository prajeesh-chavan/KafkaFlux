package config

import "os"

type Config struct {
	Simulator SimulatorConfig `yaml:"simulator"`
}

type SimulatorConfig struct {
	Workers        int    `yaml:"workers"`
	ProfilesDir    string `yaml:"profiles_dir"`
	KafkaServers   string `yaml:"kafka_servers"`
}

type RuntimeConfig struct {
	Config
	Mode       string
	Broker     string
	OutputPath string
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

	return &RuntimeConfig{
		Config:     *cfg,
		Mode:       mode,
		Broker:     broker,
		OutputPath: outputPath,
	}, nil
}