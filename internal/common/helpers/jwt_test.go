package helpers

import (
	"testing"
)

func TestGenerateAndValidateTokens(t *testing.T) {
	userID := "test-user-123"
	mobileNumber := "+1234567890"

	// Test access token generation
	accessToken, err := GenerateAccessToken(userID, mobileNumber)
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	// Test refresh token generation
	refreshToken, err := GenerateRefreshToken(userID, mobileNumber)
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	// Test access token validation
	accessClaims, err := ValidateToken(accessToken)
	if err != nil {
		t.Fatalf("Failed to validate access token: %v", err)
	}

	if accessClaims.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, accessClaims.UserID)
	}

	if accessClaims.MobileNumber != mobileNumber {
		t.Errorf("Expected MobileNumber %s, got %s", mobileNumber, accessClaims.MobileNumber)
	}

	if accessClaims.TokenType != "access" {
		t.Errorf("Expected TokenType 'access', got %s", accessClaims.TokenType)
	}

	// Test refresh token validation
	refreshClaims, err := ValidateToken(refreshToken)
	if err != nil {
		t.Fatalf("Failed to validate refresh token: %v", err)
	}

	if refreshClaims.TokenType != "refresh" {
		t.Errorf("Expected TokenType 'refresh', got %s", refreshClaims.TokenType)
	}

	// Test token expiration check
	if IsTokenExpired(accessToken) {
		t.Error("Access token should not be expired")
	}

	if IsTokenExpired(refreshToken) {
		t.Error("Refresh token should not be expired")
	}

	// Test refresh access token
	newAccessToken, err := RefreshAccessToken(refreshToken)
	if err != nil {
		t.Fatalf("Failed to refresh access token: %v", err)
	}

	newAccessClaims, err := ValidateToken(newAccessToken)
	if err != nil {
		t.Fatalf("Failed to validate new access token: %v", err)
	}

	if newAccessClaims.UserID != userID {
		t.Errorf("Expected UserID %s in new token, got %s", userID, newAccessClaims.UserID)
	}
}

func TestInvalidToken(t *testing.T) {
	// Test invalid token format
	_, err := ValidateToken("invalid.token")
	if err == nil {
		t.Error("Expected error for invalid token format")
	}

	// Test completely invalid token
	_, err = ValidateToken("not-a-token")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

func TestExpiredToken(t *testing.T) {
	// This test would require mocking time or creating a token with very short expiry
	// For now, we'll test the IsTokenExpired function with an invalid token
	if !IsTokenExpired("invalid-token") {
		t.Error("Invalid token should be considered expired")
	}
}
