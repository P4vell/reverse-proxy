package config

import (
	"encoding/json"
	"os"
	"time"
)

type Server struct {
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Host     string `json:"host"`
}

type HealthChecker struct {
	Interval time.Duration `json:"interval"`
	Timeout  time.Duration `json:"timeout"`
}

type Config struct {
	Port        int           `json:"port"`
	Servers     []Server      `json:"servers"`
	HealthCheck HealthChecker `json:"health_check"`
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
