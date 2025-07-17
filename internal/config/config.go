package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server        ServerConfig        `json:"server"`
	HomeAssistant HomeAssistantConfig `json:"home_assistant"`
	LLM           LLMConfig           `json:"llm"`
	Storage       StorageConfig       `json:"storage"`
	LogLevel      string              `json:"log_level"`
}

type ServerConfig struct {
	Port         int           `json:"port"`
	Host         string        `json:"host"`
	Mode         string        `json:"mode"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
}

type HomeAssistantConfig struct {
	URL     string `json:"url"`
	Token   string `json:"token"`
	Timeout int    `json:"timeout"`
}

type LLMConfig struct {
	OllamaURL   string  `json:"ollama_url"`
	Model       string  `json:"model"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float32 `json:"temperature"`
	TopP        float32 `json:"top_p"`
	TopK        int     `json:"top_k"`
	Timeout     int     `json:"timeout"`
}

type StorageConfig struct {
	Type     string `json:"type"`
	Path     string `json:"path"`
	InMemory bool   `json:"in_memory"`
}

func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Port:         getEnvAsInt("SERVER_PORT", 8080),
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Mode:         getEnv("SERVER_MODE", "debug"),
			ReadTimeout:  time.Duration(getEnvAsInt("SERVER_READ_TIMEOUT", 10)) * time.Second,
			WriteTimeout: time.Duration(getEnvAsInt("SERVER_WRITE_TIMEOUT", 10)) * time.Second,
		},
		HomeAssistant: HomeAssistantConfig{
			URL:     getEnv("HA_URL", "http://homeassistant.local:8123"),
			Token:   getEnv("HA_TOKEN", ""),
			Timeout: getEnvAsInt("HA_TIMEOUT", 30),
		},
		LLM: LLMConfig{
			OllamaURL:   getEnv("OLLAMA_URL", "http://localhost:11434"),
			Model:       getEnv("OLLAMA_MODEL", "llama3.2"),
			MaxTokens:   getEnvAsInt("LLM_MAX_TOKENS", 512),
			Temperature: getEnvAsFloat32("LLM_TEMPERATURE", 0.7),
			TopP:        getEnvAsFloat32("LLM_TOP_P", 0.9),
			TopK:        getEnvAsInt("LLM_TOP_K", 40),
			Timeout:     getEnvAsInt("LLM_TIMEOUT", 30),
		},
		Storage: StorageConfig{
			Type:     getEnv("STORAGE_TYPE", "memory"),
			Path:     getEnv("STORAGE_PATH", "./data"),
			InMemory: getEnvAsBool("STORAGE_IN_MEMORY", true),
		},
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsFloat32(key string, defaultValue float32) float32 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 32); err == nil {
			return float32(floatValue)
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
