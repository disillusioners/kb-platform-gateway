package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Services ServicesConfig
	Database DatabaseConfig
	S3       S3Config
	Temporal TemporalConfig
}

type ServerConfig struct {
	Host string
	Port int
	Mode string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

type S3Config struct {
	Bucket          string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	Endpoint        string // Optional for S3-compatible services
}

type TemporalConfig struct {
	Host      string
	Port      int
	Namespace string
}

type ServicesConfig struct {
	PythonCoreHost string
	PythonCorePort int
	TemporalHost   string
	TemporalPort   int
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
	}

	cfg := &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnvAsInt("SERVER_PORT", 8080),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Services: ServicesConfig{
			PythonCoreHost: getEnv("PYTHON_CORE_HOST", "python-llama-core"),
			PythonCorePort: getEnvAsInt("PYTHON_CORE_PORT", 8000),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "postgres"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "kb_user"),
			Password: getEnv("DB_PASSWORD", "kb_password"),
			Database: getEnv("DB_NAME", "kb_platform"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		S3: S3Config{
			Bucket:          getEnv("S3_BUCKET", "kb-documents"),
			Region:          getEnv("S3_REGION", "us-east-1"),
			AccessKeyID:     getEnv("S3_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("S3_SECRET_ACCESS_KEY", ""),
			Endpoint:        getEnv("S3_ENDPOINT", ""),
		},
		Temporal: TemporalConfig{
			Host:      getEnv("TEMPORAL_HOST", "temporal"),
			Port:      getEnvAsInt("TEMPORAL_PORT", 7233),
			Namespace: getEnv("TEMPORAL_NAMESPACE", "default"),
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
