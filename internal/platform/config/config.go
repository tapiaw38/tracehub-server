package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Claude   ClaudeConfig
}

type ServerConfig struct {
	Host string
	Port string
	Mode string // debug or release
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type AuthConfig struct {
	SecretKey string
}

type ClaudeConfig struct {
	APIKey    string
	Model     string
	MaxTokens int
}

var config *Config

// Load loads configuration from environment variables
func Load() error {
	// Load .env file if exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config = &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "tracehub"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Auth: AuthConfig{
			SecretKey: getEnv("JWT_SECRET", "your-secret-key-change-me"),
		},
		Claude: ClaudeConfig{
			APIKey:    getEnv("ANTHROPIC_API_KEY", ""),
			Model:     getEnv("CLAUDE_MODEL", "claude-sonnet-4-20250514"),
			MaxTokens: getEnvInt("CLAUDE_MAX_TOKENS", 4000),
		},
	}

	return config.Validate()
}

// Get returns the current configuration
func Get() *Config {
	if config == nil {
		log.Fatal("Configuration not loaded. Call Load() first.")
	}
	return config
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	if c.Auth.SecretKey == "your-secret-key-change-me" {
		log.Println("WARNING: Using default JWT secret key. Set JWT_SECRET in production!")
	}
	return nil
}

// GetDatabaseURL returns the PostgreSQL connection string
func (c *Config) GetDatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}
