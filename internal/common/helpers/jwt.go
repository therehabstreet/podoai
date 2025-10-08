package helpers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/therehabstreet/podoai/internal/common/config"
	pb "github.com/therehabstreet/podoai/proto/common"
)

// JWTClaims represents the claims structure for our JWT tokens
type JWTClaims struct {
	UserID        string   `json:"user_id"`
	Roles         []string `json:"roles"`
	TokenType     string   `json:"token_type"` // "access" or "refresh"
	AppType       string   `json:"app_type"`   // "clinical" or "consumer"
	OwnerEntityID string   `json:"owner_entity_id"`
	ExpiresAt     int64    `json:"exp"`
	IssuedAt      int64    `json:"iat"`
	NotBefore     int64    `json:"nbf"`
	Issuer        string   `json:"iss"`
	Subject       string   `json:"sub"`
	ID            string   `json:"jti"`
}

// JWTHeader represents the JWT header
type JWTHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

// GenerateAccessToken generates a JWT access token
// Accepts string roles directly to avoid conversion overhead
func GenerateAccessToken(cfg *config.Config, userID string, roles []string, appType string) (string, error) {
	now := time.Now()

	claims := JWTClaims{
		UserID:    userID,
		Roles:     roles,
		TokenType: "access",
		AppType:   appType,
		ExpiresAt: now.Add(time.Duration(cfg.JWT.AccessExpiryMin) * time.Minute).Unix(),
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
		Issuer:    "podoai",
		Subject:   userID,
		ID:        generateTokenID(),
	}

	return generateJWT(claims, cfg.JWT.Secret)
}

// GenerateRefreshToken generates a JWT refresh token
// Accepts string roles directly to avoid conversion overhead
func GenerateRefreshToken(cfg *config.Config, userID string, roles []string, appType string) (string, error) {
	now := time.Now()

	claims := JWTClaims{
		UserID:    userID,
		Roles:     roles,
		TokenType: "refresh",
		AppType:   appType,
		ExpiresAt: now.Add(time.Duration(cfg.JWT.RefreshExpiryMin) * time.Minute).Unix(),
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
		Issuer:    "podoai",
		Subject:   userID,
		ID:        generateTokenID(),
	}

	return generateJWT(claims, cfg.JWT.Secret)
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(cfg *config.Config, tokenString string) (*JWTClaims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// Decode header
	headerData, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid header encoding")
	}

	var header JWTHeader
	if err := json.Unmarshal(headerData, &header); err != nil {
		return nil, fmt.Errorf("invalid header format")
	}

	if header.Algorithm != "HS256" {
		return nil, fmt.Errorf("unsupported algorithm: %s", header.Algorithm)
	}

	// Decode payload
	payloadData, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid payload encoding")
	}

	var claims JWTClaims
	if err := json.Unmarshal(payloadData, &claims); err != nil {
		return nil, fmt.Errorf("invalid payload format")
	}

	// Verify signature
	expectedSignature := generateSignature(parts[0]+"."+parts[1], cfg.JWT.Secret)
	if parts[2] != expectedSignature {
		return nil, fmt.Errorf("invalid signature")
	}

	// Check expiration
	if time.Now().Unix() > claims.ExpiresAt {
		return nil, fmt.Errorf("token is expired")
	}

	return &claims, nil
}

// IsTokenExpired checks if a token is expired
func IsTokenExpired(cfg *config.Config, tokenString string) bool {
	claims, err := ValidateToken(cfg, tokenString)
	if err != nil {
		return true
	}

	return time.Now().Unix() > claims.ExpiresAt
}

// RefreshAccessToken generates a new access token from a valid refresh token
func RefreshAccessToken(cfg *config.Config, refreshTokenString string) (string, error) {
	claims, err := ValidateToken(cfg, refreshTokenString)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %v", err)
	}

	if claims.TokenType != "refresh" {
		return "", fmt.Errorf("token is not a refresh token")
	}

	if time.Now().Unix() > claims.ExpiresAt {
		return "", fmt.Errorf("refresh token is expired")
	}

	// Generate new access token with string roles directly
	return GenerateAccessToken(cfg, claims.UserID, claims.Roles, claims.AppType)
}

func GetRolesFromToken(cfg *config.Config, tokenString string) ([]string, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// Validate the token first
	claims, err := ValidateToken(cfg, tokenString)
	if err != nil {
		return nil, err
	}

	return claims.Roles, nil
}

// GetProtoRolesFromToken extracts roles from a JWT token as proto roles
// Use this when you need proto roles for API responses
func GetProtoRolesFromToken(cfg *config.Config, tokenString string) ([]pb.Role, error) {
	roleStrs, err := GetRolesFromToken(cfg, tokenString)
	if err != nil {
		return nil, err
	}
	return StringsToRoles(roleStrs), nil
}

// HasRole checks if the token contains a specific role
// Accepts proto role for API convenience but works with strings internally
func HasRole(cfg *config.Config, tokenString string, role pb.Role) bool {
	roles, err := GetRolesFromToken(cfg, tokenString)
	if err != nil {
		return false
	}

	// Convert proto role to string for comparison
	roleStr := strings.ToLower(role.String())

	for _, r := range roles {
		if r == roleStr {
			return true
		}
	}
	return false
}

// HasRoleString checks if the token contains a specific role by string
// Use this when working with string roles directly
func HasRoleString(cfg *config.Config, tokenString string, role string) bool {
	roles, err := GetRolesFromToken(cfg, tokenString)
	if err != nil {
		return false
	}

	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the token contains any of the specified roles
// Accepts proto roles for API convenience but works with strings internally
func HasAnyRole(cfg *config.Config, tokenString string, requiredRoles []pb.Role) bool {
	userRoles, err := GetRolesFromToken(cfg, tokenString)
	if err != nil {
		return false
	}

	// Convert proto roles to strings for comparison
	requiredRoleStrs := RolesToStrings(requiredRoles)

	for _, userRole := range userRoles {
		for _, requiredRole := range requiredRoleStrs {
			if userRole == requiredRole {
				return true
			}
		}
	}
	return false
}

// HasAnyRoleString checks if the token contains any of the specified roles by strings
// Use this when working with string roles directly
func HasAnyRoleString(cfg *config.Config, tokenString string, requiredRoles []string) bool {
	userRoles, err := GetRolesFromToken(cfg, tokenString)
	if err != nil {
		return false
	}

	for _, userRole := range userRoles {
		for _, requiredRole := range requiredRoles {
			if userRole == requiredRole {
				return true
			}
		}
	}
	return false
}

// generateJWT creates a JWT token with the given claims and secret
func generateJWT(claims JWTClaims, secret string) (string, error) {
	header := JWTHeader{
		Algorithm: "HS256",
		Type:      "JWT",
	}

	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	headerEncoded := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(headerBytes)
	claimsEncoded := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(claimsBytes)

	payload := headerEncoded + "." + claimsEncoded
	signature := generateSignature(payload, secret)

	return payload + "." + signature, nil
}

// generateSignature generates HMAC-SHA256 signature
func generateSignature(payload, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(h.Sum(nil))
}

// generateTokenID generates a unique token ID
func generateTokenID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
