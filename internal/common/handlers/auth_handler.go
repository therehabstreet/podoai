package handlers

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/therehabstreet/podoai/internal/common/helpers"
	"github.com/therehabstreet/podoai/internal/common/models"
	pb "github.com/therehabstreet/podoai/proto/common"
)

// RequestOtp handles OTP request
func (cs *CommonServer) RequestOtp(ctx context.Context, req *pb.RequestOtpRequest) (*pb.RequestOtpResponse, error) {
	mobileNumber := req.GetMobileNumber()

	// Validate mobile number
	if mobileNumber == "" {
		return &pb.RequestOtpResponse{
			Success: false,
			Message: "Mobile number is required",
		}, nil
	}

	// Generate OTP
	otp, err := generateOTP()
	if err != nil {
		return &pb.RequestOtpResponse{
			Success: false,
			Message: "Failed to generate OTP",
		}, nil
	}

	// Store OTP in database with expiration
	otpModel := models.NewOTP(mobileNumber, otp, 5) // 5 minutes expiry
	err = cs.DBClient.StoreOTP(ctx, otpModel)
	if err != nil {
		return &pb.RequestOtpResponse{
			Success: false,
			Message: "Failed to store OTP",
		}, nil
	}

	// Send OTP via configured sender (WhatsApp, SMS, etc.)
	if cs.OTPSender != nil {
		err = cs.OTPSender.SendOTP(ctx, mobileNumber, otp)
		if err != nil {
			return &pb.RequestOtpResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to send OTP via %s", cs.OTPSender.GetProviderName()),
			}, nil
		}
	} else {
		// Fallback: log the OTP if no sender is configured
		fmt.Printf("OTP for %s: %s (no sender configured)\n", mobileNumber, otp)
	}

	return &pb.RequestOtpResponse{
		Success: true,
		Message: "OTP sent successfully",
	}, nil
}

// VerifyOtp handles OTP verification
func (cs *CommonServer) VerifyOtp(ctx context.Context, req *pb.VerifyOtpRequest) (*pb.LoginResponse, error) {
	mobileNumber := req.GetMobileNumber()
	otp := req.GetOtp()

	// Validate inputs
	if mobileNumber == "" || otp == "" {
		return nil, fmt.Errorf("mobile number and OTP are required")
	}

	// Get OTP from database
	storedOTP, err := cs.DBClient.GetOTPByMobileNumber(ctx, mobileNumber)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired OTP")
	}

	// Increment attempt count
	err = cs.DBClient.IncrementOTPAttempts(ctx, storedOTP.ID)
	if err != nil {
		// Log error but continue
		fmt.Printf("Failed to increment OTP attempts: %v\n", err)
	}

	// Check if OTP is valid
	if !storedOTP.IsValid() {
		return nil, fmt.Errorf("OTP is expired or has exceeded maximum attempts")
	}

	// Verify OTP code
	if storedOTP.Code != otp {
		return nil, fmt.Errorf("invalid OTP")
	}

	// Mark OTP as used
	err = cs.DBClient.MarkOTPAsUsed(ctx, storedOTP.ID)
	if err != nil {
		// Log error but continue
		fmt.Printf("Failed to mark OTP as used: %v\n", err)
	}

	// TODO: Get or create user based on mobile number
	userID := mobileNumber // For now, use mobile number as user ID

	// Generate JWT token
	token, err := helpers.GenerateAccessToken(userID, mobileNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	refreshToken, err := helpers.GenerateRefreshToken(userID, mobileNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %v", err)
	}

	return &pb.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

// Helper functions

// generateOTP generates a 6-digit OTP
func generateOTP() (string, error) {
	max := big.NewInt(999999)
	min := big.NewInt(100000)

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	// Ensure it's at least 6 digits
	otp := n.Int64() + min.Int64()
	if otp > 999999 {
		otp = otp - 900000
	}

	return fmt.Sprintf("%06d", otp), nil
}
