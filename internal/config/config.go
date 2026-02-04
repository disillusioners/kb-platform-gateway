package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	JWT      JWTConfig
	Services ServicesConfig
}

type ServerConfig struct {
	Host string
	Port int
	Mode string
}

type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

type ServicesConfig struct {
	PythonCoreHost string
	PythonCorePort int
	TemporalHost   string
	TemporalPort   int
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnvAsInt("SERVER_PORT", 8080),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "your-secret-key"),
			Expiration: getEnvAsDuration("JWT_EXPIRATION", 24*time.Hour),
		},
		Services: ServicesConfig{
			PythonCoreHost: getEnv("PYTHON_CORE_HOST", "python-llama-core"),
			PythonCorePort: getEnvAsInt("PYTHON_CORE_PORT", 8000),
			TemporalHost:   getEnv("TEMPORAL_HOST", "temporal"),
			TemporalPort:   getEnvAsInt("TEMPORAL_PORT", 7233),
		},
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
