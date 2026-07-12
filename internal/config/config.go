package config

import (
	"encoding/json"
	"os"
	"time"
)

type ServerConfig struct {
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Host     string `json:"host"`
}

type HealthCheckerConfig struct {
	Interval time.Duration `json:"interval"`
	Timeout  time.Duration `json:"timeout"`
}

type Config struct {
	Port        int                 `json:"port"`
	Servers     []ServerConfig      `json:"servers"`
	HealthCheck HealthCheckerConfig `json:"health_check"`
}

func Load() (Config, error) {
	file, err := os.Open("config.json")
	if err != nil {
		return Config{}, err
	}

	defer file.Close()

	decoder := json.NewDecoder(file)
	cfg := Config{}
	err = decoder.Decode(&cfg)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}
