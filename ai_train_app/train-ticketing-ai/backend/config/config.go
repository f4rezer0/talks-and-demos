package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// AI Provider
	AIProvider      string
	OpenAIAPIKey    string
	OpenAIBaseURL   string
	AnthropicAPIKey string
	OllamaURL       string
	OllamaModel     string

	// Server
	ServerPort string
}

// Load loads configuration from environment variables
func Load() *Config {
	// Try to load .env file (optional for local development)
	_ = godotenv.Load()

	config := &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "trainpass123"),
		DBName:     getEnv("DB_NAME", "traintickets"),

		AIProvider:      getEnv("AI_PROVIDER", "ollama"),
		OpenAIAPIKey:    os.Getenv("OPENAI_API_KEY"),
		OpenAIBaseURL:   getEnv("OPENAI_BASE_URL", "https://api.openai.com"),
		AnthropicAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
		OllamaURL:       getEnv("OLLAMA_URL", "http://localhost:11434"),
		OllamaModel:     getEnv("OLLAMA_MODEL", "llama2"),

		ServerPort: getEnv("SERVER_PORT", "8080"),
	}

	// Validate AI provider configuration
	switch config.AIProvider {
	case "openai":
		if config.OpenAIAPIKey == "" {
			log.Println("WARNING: OPENAI_API_KEY not set")
		}
	case "anthropic":
		if config.AnthropicAPIKey == "" {
			log.Println("WARNING: ANTHROPIC_API_KEY not set")
		}
	case "ollama":
		if config.OllamaURL == "" {
			log.Println("WARNING: OLLAMA_URL not set")
		}
	default:
		log.Printf("WARNING: Unknown AI_PROVIDER: %s (using ollama as fallback)\n", config.AIProvider)
		config.AIProvider = "ollama"
	}

	return config
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
