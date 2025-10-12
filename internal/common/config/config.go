package config

import (
	"os"
	"strconv"
)

type Config struct {
	WhatsApp *WhatsAppConfig
	JWT      *JWTConfig
	GCS      *GCSConfig
}

type WhatsAppConfig struct {
	APIKey    string
	APIURL    string
	FromPhone string
}

type JWTConfig struct {
	Secret           string
	AccessExpiryMin  int
	RefreshExpiryMin int
}

type GCSConfig struct {
	BucketName                string
	ServiceAccountKeyPath     string // Path to service account JSON file
	DefaultSignedURLExpiryMin int    // Default expiry for signed URLs in minutes
}

func NewConfig() *Config {
	accessExpiryMin := 60 * 24       // 24 hours default
	refreshExpiryMin := 60 * 24 * 30 // 30 days default

	if envValue := os.Getenv("JWT_ACCESS_EXPIRY_MINUTES"); envValue != "" {
		if val, err := strconv.Atoi(envValue); err == nil {
			accessExpiryMin = val
		}
	}

	if envValue := os.Getenv("JWT_REFRESH_EXPIRY_MINUTES"); envValue != "" {
		if val, err := strconv.Atoi(envValue); err == nil {
			refreshExpiryMin = val
		}
	}

	// GCS configuration
	gcsSignedURLExpiryMin := 15 // 15 minutes default
	if envValue := os.Getenv("GCS_SIGNED_URL_EXPIRY_MINUTES"); envValue != "" {
		if val, err := strconv.Atoi(envValue); err == nil {
			gcsSignedURLExpiryMin = val
		}
	}

	return &Config{
		WhatsApp: &WhatsAppConfig{
			APIKey:    getEnvWithDefault("WHATSAPP_API_KEY", ""),
			APIURL:    getEnvWithDefault("WHATSAPP_API_URL", "https://graph.facebook.com/v18.0"),
			FromPhone: getEnvWithDefault("WHATSAPP_FROM_PHONE", ""),
		},
		JWT: &JWTConfig{
			Secret:           getEnvWithDefault("JWT_SECRET", "your-secret-key-change-in-production"),
			AccessExpiryMin:  accessExpiryMin,
			RefreshExpiryMin: refreshExpiryMin,
		},
		GCS: &GCSConfig{
			BucketName:                getEnvWithDefault("GCS_BUCKET_NAME", "podoai-scans"),
			ServiceAccountKeyPath:     getEnvWithDefault("GCS_SERVICE_ACCOUNT_KEY_PATH", ""),
			DefaultSignedURLExpiryMin: gcsSignedURLExpiryMin,
		},
	}
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
