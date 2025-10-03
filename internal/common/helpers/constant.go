package helpers

import "context"

// App type constants for context values and collection naming
const (
	AppTypeClinical = "clinical"
	AppTypeConsumer = "consumer"
	AppTypeKey      = "appType"
)

// GetAppTypeFromContext extracts the app type from context
func GetAppTypeFromContext(ctx context.Context) string {
	if appType, ok := ctx.Value(AppTypeKey).(string); ok {
		return appType
	}
	return AppTypeConsumer // Default to consumer if not specified
}

// IsClinicalApp checks if the current context is for clinical app
func IsClinicalApp(ctx context.Context) bool {
	return GetAppTypeFromContext(ctx) == AppTypeClinical
}

// IsConsumerApp checks if the current context is for consumer app
func IsConsumerApp(ctx context.Context) bool {
	return GetAppTypeFromContext(ctx) == AppTypeConsumer
}
