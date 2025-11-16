package config

import (
	"log"
	"os"
	"time"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Env         string `yaml:"env"`
	StoragePath string `yaml:"storage_path"`
	Server      struct {
		Port         int           `yaml:"port"`
		ReadTimeout  time.Duration `yaml:"readTimeout"`
		WriteTimeout time.Duration `yaml:"writeTimeout"`
		IdleTimeout  time.Duration `yaml:"idleTimeout"`
	} `yaml:"server"`

	Database struct {
		Host            string        `yaml:"host"`
		Port            int           `yaml:"port"`
		User            string        `yaml:"user"`
		Password        string        `yaml:"password"`
		Name            string        `yaml:"name"`
		SSLMode         string        `yaml:"sslMode"`
		MaxOpenConns    int           `yaml:"maxOpenConns"`
		MaxIdleConns    int           `yaml:"maxIdleConns"`
		ConnMaxLifetime time.Duration `yaml:"connMaxLifetime"`
	} `yaml:"database"`

	Logging struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
	} `yaml:"logging"`

	App struct {
		ReviewerCount int `yaml:"reviewerCount"`
		RandomSeed    int `yaml:"randomSeed"`
	} `yaml:"app"`
}

func Load(path string) *Config {
	cfg := &Config{}

	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read config: %s", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		log.Fatalf("Failed to Unmarshal config: %s", err)
	}

	return cfg
}
