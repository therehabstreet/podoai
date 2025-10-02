package helpers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// JWTClaims represents the claims structure for our JWT tokens
type JWTClaims struct {
	UserID       string `json:"user_id"`
	MobileNumber string `json:"mobile_number"`
	TokenType    string `json:"token_type"` // "access" or "refresh"
	ExpiresAt    int64  `json:"exp"`
	IssuedAt     int64  `json:"iat"`
	NotBefore    int64  `json:"nbf"`
	Issuer       string `json:"iss"`
	Subject      string `json:"sub"`
	ID           string `json:"jti"`
}

// JWTHeader represents the JWT header
type JWTHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret           string
	AccessExpiryMin  int
	RefreshExpiryMin int
}

// LoadJWTConfig loads JWT configuration from environment variables
func LoadJWTConfig() JWTConfig {
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

	return JWTConfig{
		Secret:           getEnvWithDefault("JWT_SECRET", "your-secret-key-change-in-production"),
		AccessExpiryMin:  accessExpiryMin,
		RefreshExpiryMin: refreshExpiryMin,
	}
}

// GenerateAccessToken generates a JWT access token
func GenerateAccessToken(userID, mobileNumber string) (string, error) {
	config := LoadJWTConfig()
	now := time.Now()

	claims := JWTClaims{
		UserID:       userID,
		MobileNumber: mobileNumber,
		TokenType:    "access",
		ExpiresAt:    now.Add(time.Duration(config.AccessExpiryMin) * time.Minute).Unix(),
		IssuedAt:     now.Unix(),
		NotBefore:    now.Unix(),
		Issuer:       "podoai",
		Subject:      userID,
		ID:           generateTokenID(),
	}

	return generateJWT(claims, config.Secret)
}

// GenerateRefreshToken generates a JWT refresh token
func GenerateRefreshToken(userID, mobileNumber string) (string, error) {
	config := LoadJWTConfig()
	now := time.Now()

	claims := JWTClaims{
		UserID:       userID,
		MobileNumber: mobileNumber,
		TokenType:    "refresh",
		ExpiresAt:    now.Add(time.Duration(config.RefreshExpiryMin) * time.Minute).Unix(),
		IssuedAt:     now.Unix(),
		NotBefore:    now.Unix(),
		Issuer:       "podoai",
		Subject:      userID,
		ID:           generateTokenID(),
	}

	return generateJWT(claims, config.Secret)
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string) (*JWTClaims, error) {
	config := LoadJWTConfig()

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
	expectedSignature := generateSignature(parts[0]+"."+parts[1], config.Secret)
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
func IsTokenExpired(tokenString string) bool {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return true
	}

	return time.Now().Unix() > claims.ExpiresAt
}

// RefreshAccessToken generates a new access token from a valid refresh token
func RefreshAccessToken(refreshTokenString string) (string, error) {
	claims, err := ValidateToken(refreshTokenString)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %v", err)
	}

	if claims.TokenType != "refresh" {
		return "", fmt.Errorf("token is not a refresh token")
	}

	if time.Now().Unix() > claims.ExpiresAt {
		return "", fmt.Errorf("refresh token is expired")
	}

	return GenerateAccessToken(claims.UserID, claims.MobileNumber)
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

// getEnvWithDefault gets environment variable with a default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
