package config

type Config struct {
	Simulator SimulatorConfig `yaml:"simulator"`
}

type SimulatorConfig struct {
	Workers        int    `yaml:"workers"`
	ProfilesDir    string `yaml:"profiles_dir"`
	KafkaServers   string `yaml:"kafka_servers"`
}