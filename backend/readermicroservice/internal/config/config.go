package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type DBConfig struct {
	Host     string `yaml:"host" env:"DB_HOST"`
	Port     int    `yaml:"port" env:"DB_PORT"`
	User     string `yaml:"user" env:"DB_USER"`
	Password string `yaml:"password" env:"DB_PASSWORD"`
	Database string `yaml:"dbname" env:"DB_NAME"`
}

type KafkaConfig struct {
	Brokers []string `yaml:"brokers" env:"KAFKA_BROKERS"`
	Topic   string   `yaml:"topic" env:"KAFKA_TOPIC"`
}

type CacheConfig struct {
	MaxSize         int           `yaml:"max_size" env:"CACHE_MAX_SIZE"`
	DefaultTTL      time.Duration `yaml:"default_ttl" env:"CACHE_DEFAULT_TTL"`
	CleanupInterval time.Duration `yaml:"cleanup_interval" env:"CACHE_CLEANUP_INTERVAL"`
}

type RetryConfig struct {
	MaxRetries int           `yaml:"max_retries" env:"RETRY_MAX_RETRIES"`
	BaseDelay  time.Duration `yaml:"base_delay" env:"RETRY_BASE_DELAY"`
}

type AppConfig struct {
	DB    DBConfig    `yaml:"db"`
	Kafka KafkaConfig `yaml:"kafka"`
	Cache CacheConfig `yaml:"cache"`
	Retry RetryConfig `yaml:"retry"`
}

var RLogger *log.Logger

func Configure() {
	RLogger = log.New(os.Stdout, "LOGGER: ", log.LstdFlags)
}

func LoadConfig(path string) (*AppConfig, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config AppConfig
	if err := yaml.Unmarshal(file, &config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Override with environment variables
	if err := overrideWithEnv(&config); err != nil {
		return nil, fmt.Errorf("error overriding with env: %w", err)
	}

	return &config, nil
}

func overrideWithEnv(cfg *AppConfig) error {
	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.DB.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		p, err := strconv.Atoi(port)
		if err != nil {
			return fmt.Errorf("invalid DB_PORT: %w", err)
		}
		cfg.DB.Port = p
	}
	if user := os.Getenv("DB_USER"); user != "" {
		cfg.DB.User = user
	}
	if pass := os.Getenv("DB_PASSWORD"); pass != "" {
		cfg.DB.Password = pass
	}
	if db := os.Getenv("DB_NAME"); db != "" {
		cfg.DB.Database = db
	}

	return nil
}
