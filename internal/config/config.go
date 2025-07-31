package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	// Server Configuration
	Server ServerConfig `mapstructure:"server"`
	
	// WhatsApp Configuration
	WhatsApp WhatsAppConfig `mapstructure:"whatsapp"`
	
	// Fonnte Configuration
	Fonnte FontteConfig `mapstructure:"fonnte"`
	
	// OpenAI Configuration
	OpenAI OpenAIConfig `mapstructure:"openai"`
	
	// Image Generation Configuration
	Image ImageConfig `mapstructure:"image"`
	
	// Database Configuration
	Database DatabaseConfig `mapstructure:"database"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

type WhatsAppConfig struct {
	SessionPath string `mapstructure:"session_path"`
	LogLevel    string `mapstructure:"log_level"`
}

type FontteConfig struct {
	APIKey     string `mapstructure:"api_key"`
	WebhookURL string `mapstructure:"webhook_url"`
}

type OpenAIConfig struct {
	APIKey    string `mapstructure:"api_key"`
	BaseURL   string `mapstructure:"base_url"`
	Model     string `mapstructure:"model"`
	MaxTokens int    `mapstructure:"max_tokens"`
}

type ImageConfig struct {
	Provider string `mapstructure:"provider"`
	APIKey   string `mapstructure:"api_key"`
}

type DatabaseConfig struct {
	URL string `mapstructure:"url"`
}

func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// .env file is optional, so we don't return error
		fmt.Println("No .env file found, using environment variables")
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// Set environment variable mappings
	viper.SetEnvPrefix("")
	viper.AutomaticEnv()

	// Set defaults
	setDefaults()

	// Read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate required fields
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", "8080")

	// WhatsApp defaults
	viper.SetDefault("whatsapp.session_path", "./sessions")
	viper.SetDefault("whatsapp.log_level", "INFO")

	// OpenAI defaults
	viper.SetDefault("openai.model", "gpt-4-turbo-preview")
	viper.SetDefault("openai.max_tokens", 1000)

	// Image defaults
	viper.SetDefault("image.provider", "openai")

	// Database defaults
	viper.SetDefault("database.url", "sqlite://./bot.db")

	// Bind environment variables
	viper.BindEnv("server.host", "SERVER_HOST")
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("whatsapp.session_path", "WHATSAPP_SESSION_PATH")
	viper.BindEnv("whatsapp.log_level", "WHATSAPP_LOG_LEVEL")
	viper.BindEnv("fonnte.api_key", "FONNTE_API_KEY")
	viper.BindEnv("fonnte.webhook_url", "FONNTE_WEBHOOK_URL")
	viper.BindEnv("openai.api_key", "OPENAI_API_KEY")
	viper.BindEnv("openai.base_url", "OPENAI_BASE_URL")
	viper.BindEnv("openai.model", "OPENAI_MODEL")
	viper.BindEnv("openai.max_tokens", "OPENAI_MAX_TOKENS")
	viper.BindEnv("image.provider", "IMAGE_API_PROVIDER")
	viper.BindEnv("image.api_key", "IMAGE_API_KEY")
	viper.BindEnv("database.url", "DATABASE_URL")
}

func validateConfig(config *Config) error {
	if config.OpenAI.APIKey == "" || config.OpenAI.APIKey == "your_openai_api_key" {
		return fmt.Errorf("OPENAI_API_KEY is required")
	}

	if config.Fonnte.APIKey == "" || config.Fonnte.APIKey == "your_fonnte_api_key" {
		return fmt.Errorf("FONNTE_API_KEY is required")
	}

	// Create sessions directory if it doesn't exist
	if err := os.MkdirAll(config.WhatsApp.SessionPath, 0755); err != nil {
		return fmt.Errorf("failed to create sessions directory: %w", err)
	}

	return nil
}