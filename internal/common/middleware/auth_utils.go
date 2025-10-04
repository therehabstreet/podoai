package middleware

import "context"

// Context key types to avoid collisions
type contextKey string

const (
	UserIDKey        contextKey = "user_id"
	RolesKey         contextKey = "roles"
	TokenTypeKey     contextKey = "token_type"
	OwnerEntityIDKey contextKey = "owner_entity_id"
)

// Helper functions to extract data from context

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// GetRolesFromContext extracts roles from context
func GetRolesFromContext(ctx context.Context) ([]string, bool) {
	roles, ok := ctx.Value(RolesKey).([]string)
	return roles, ok
}

// GetOwnerEntityIDFromContext extracts owner entity ID from context
func GetOwnerEntityIDFromContext(ctx context.Context) (string, bool) {
	ownerEntityID, ok := ctx.Value(OwnerEntityIDKey).(string)
	return ownerEntityID, ok
}

// GetTokenTypeFromContext extracts token type from context
func GetTokenTypeFromContext(ctx context.Context) (string, bool) {
	tokenType, ok := ctx.Value(TokenTypeKey).(string)
	return tokenType, ok
}
