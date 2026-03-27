package config

type Config struct {
	Region      string
	ClusterName string
}

func LoadConfig() *Config {
	return &Config{
		Region:      "eu-west-2",
		ClusterName: "stackpulse-cluster",
	}
}
