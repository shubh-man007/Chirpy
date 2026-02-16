package config

import (
	"net/http"
	"os"
	"time"
)

// Config holds runtime configuration for the Chirpy TUI.
type Config struct {
	BaseURL    string
	HTTPClient *http.Client
}

func New() Config {
	baseURL := os.Getenv("CHIRPY_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	httpClient := &http.Client{
		Timeout: 15 * time.Second,
	}

	return Config{
		BaseURL:    baseURL,
		HTTPClient: httpClient,
	}
}
