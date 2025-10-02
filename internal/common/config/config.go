package config

import "os"

type Config struct {
	WhatsApp WhatsAppConfig
}

type WhatsAppConfig struct {
	APIKey    string
	APIURL    string
	FromPhone string
}

func NewConfig() Config {
	return Config{
		WhatsApp: WhatsAppConfig{
			APIKey:    getEnvWithDefault("WHATSAPP_API_KEY", ""),
			APIURL:    getEnvWithDefault("WHATSAPP_API_URL", "https://graph.facebook.com/v18.0"),
			FromPhone: getEnvWithDefault("WHATSAPP_FROM_PHONE", ""),
		},
	}
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
