package app

import (
	"fmt"
	"os"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
)

// Config holds the application's configuration settings.
// Each field corresponds to an environment variable and has a default value.
type Config struct {
	AppPort                 int
	DatabaseURL             string
	DebugMode               bool
	GOOGLE_CLIENT_ID        string
	OPENAI_API_KEY          string
	GOOGLE_MOBILE_CLIENT_ID string
	UPLOAD_DIR              string
	EXPORT_DIR              string
	STRIPE_API_KEY          string
}

func NewMockConfig() *Config {
	return &Config{
		AppPort:                 7070,
		DatabaseURL:             "mongodb://localhost:27017",
		DebugMode:               true,
		GOOGLE_CLIENT_ID:        "mock-google-client-id",
		OPENAI_API_KEY:          "mock-openai-api-key",
		GOOGLE_MOBILE_CLIENT_ID: "mock-google-mobile-client-id",
		UPLOAD_DIR:              "uploads",
		EXPORT_DIR:              "exports",
		STRIPE_API_KEY:          "mock-stripe-api-key",
	}
}

// LoadConfig loads configuration from environment variables,
// falling back to default values if variables are not set.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		// Set default values
		AppPort:                 7070,
		DatabaseURL:             "",
		DebugMode:               true,
		GOOGLE_CLIENT_ID:        "",
		OPENAI_API_KEY:          "",
		GOOGLE_MOBILE_CLIENT_ID: "",
		UPLOAD_DIR:              "uploads",
		EXPORT_DIR:              "exports",
		STRIPE_API_KEY:          "",
	}

	// Load AppPort from environment variable "APP_PORT"
	if envPort, found := os.LookupEnv("APP_PORT"); found {
		port, err := strconv.Atoi(envPort)
		if err != nil {
			return nil, fmt.Errorf("invalid APP_PORT environment variable: %w", err)
		}
		cfg.AppPort = port
	}

	// Load DatabaseURL from environment variable "DATABASE_URL"
	if envDBURL, found := os.LookupEnv("DATABASE_URL"); found {
		cfg.DatabaseURL = envDBURL
	}

	// Load DebugMode from environment variable "DEBUG_MODE"
	// Any non-empty string is considered true, or specifically "true"
	if envDebug, found := os.LookupEnv("DEBUG_MODE"); found {
		// More robust check: consider "true", "1", "yes" as true
		cfg.DebugMode = (envDebug == "true" || envDebug == "1" || envDebug == "yes")
	}

	// Load GOOGLE_CLIENT_ID from environment variable "GOOGLE_CLIENT_ID"
	if envGoogleClientID, found := os.LookupEnv("GOOGLE_CLIENT_ID"); found {
		cfg.GOOGLE_CLIENT_ID = envGoogleClientID
	}

	// Load OPENAI_API_KEY from environment variable "OPENAI_API_KEY"
	if envOpenAIKey, found := os.LookupEnv("OPENAI_API_KEY"); found {
		cfg.OPENAI_API_KEY = envOpenAIKey
	}

	// Load GOOGLE_MOBILE_CLIENT_ID from environment variable "GOOGLE_MOBILE_CLIENT_ID"
	if envGoogleMobileClientID, found := os.LookupEnv("GOOGLE_MOBILE_CLIENT_ID"); found {
		cfg.GOOGLE_MOBILE_CLIENT_ID = envGoogleMobileClientID
	}

	// Load UPLOAD_DIR from environment variable "UPLOAD_DIR"
	if envUploadDir, found := os.LookupEnv("UPLOAD_DIR"); found {
		cfg.UPLOAD_DIR = envUploadDir
	}

	// Load EXPORT_DIR from environment variable "EXPORT_DIR"
	if envExportDir, found := os.LookupEnv("EXPORT_DIR"); found {
		cfg.EXPORT_DIR = envExportDir
	}

	// Load STRIPE_API_KEY from environment variable "STRIPE_API_KEY"
	if envStripeAPIKey, found := os.LookupEnv("STRIPE_API_KEY"); found {
		cfg.STRIPE_API_KEY = envStripeAPIKey
	}

	return cfg, nil
}
