package models

import (
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// OTP model for storing OTP verification codes
type OTP struct {
	ID          string    `bson:"_id,omitempty"`
	PhoneNumber string    `bson:"phone_number"`
	Code        string    `bson:"code"`
	CreatedAt   time.Time `bson:"created_at"`
	ExpiresAt   time.Time `bson:"expires_at"`
	IsUsed      bool      `bson:"is_used"`
	Attempts    int       `bson:"attempts"`
	MaxAttempts int       `bson:"max_attempts"`
}

// IsExpired checks if the OTP has expired
func (o *OTP) IsExpired() bool {
	return time.Now().After(o.ExpiresAt)
}

// IsValid checks if the OTP is valid for verification
func (o *OTP) IsValid() bool {
	return !o.IsUsed && !o.IsExpired() && o.Attempts < o.MaxAttempts
}

// NewOTP creates a new OTP instance
func NewOTP(phoneNumber, code string, expiryMinutes int) *OTP {
	now := time.Now()
	maxAttempts := getOTPMaxAttempts()

	return &OTP{
		ID:          uuid.NewString(),
		PhoneNumber: phoneNumber,
		Code:        code,
		CreatedAt:   now,
		ExpiresAt:   now.Add(time.Duration(expiryMinutes) * time.Minute),
		IsUsed:      false,
		Attempts:    0,
		MaxAttempts: maxAttempts,
	}
}

// getOTPMaxAttempts gets the maximum attempts from environment variable
func getOTPMaxAttempts() int {
	if envValue := os.Getenv("OTP_MAX_ATTEMPTS"); envValue != "" {
		if val, err := strconv.Atoi(envValue); err == nil {
			return val
		}
	}
	return 3 // Default to 3 attempts
}
