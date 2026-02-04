package config

import "os"

// Config holds the application configuration
type Config struct {
	Port        string
	SessionPath string
	WebhookURL  string
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	sessionPath := os.Getenv("SESSION_PATH")
	if sessionPath == "" {
		sessionPath = "./sessions"
	}

	webhookURL := os.Getenv("WEBHOOK_URL")

	return &Config{
		Port:        port,
		SessionPath: sessionPath,
		WebhookURL:  webhookURL,
	}
}
