package handlers

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	clinicalModels "github.com/therehabstreet/podoai/internal/clinical/models"
	"github.com/therehabstreet/podoai/internal/common/helpers"
	"github.com/therehabstreet/podoai/internal/common/models"
	pb "github.com/therehabstreet/podoai/proto/common"
)

// RequestOtp handles OTP request with different flows for clinical and consumer
func (cs *CommonServer) RequestOtp(ctx context.Context, req *pb.RequestOtpRequest) (*pb.RequestOtpResponse, error) {
	phoneNumber := req.GetPhoneNumber()

	// Validate phone number
	if phoneNumber == "" {
		return &pb.RequestOtpResponse{
			Success: false,
			Message: "Phone number is required",
		}, nil
	}

	// Check if this is a clinical app and validate user existence
	// For consumer apps, we allow any phone number and create user if needed
	if helpers.IsClinicalApp(ctx) {
		exists, err := cs.DBClient.ClinicalUserExists(ctx, phoneNumber)
		if err != nil {
			return &pb.RequestOtpResponse{
				Success: false,
				Message: "Failed to validate user",
			}, nil
		}
		if !exists {
			return &pb.RequestOtpResponse{
				Success: false,
				Message: "User not found. Please contact your admin.",
			}, nil
		}
	}

	// Check for recent OTP requests to prevent abuse
	existingOTP, err := cs.DBClient.GetOTPByPhoneNumber(ctx, phoneNumber)
	if err == nil && existingOTP != nil {
		// Check if there's a recent valid OTP (within last 60 seconds)
		timeSinceCreation := time.Since(existingOTP.CreatedAt)
		if timeSinceCreation < 60*time.Second && !existingOTP.IsUsed {
			return &pb.RequestOtpResponse{
				Success: false,
				Message: fmt.Sprintf("Please wait %d seconds before requesting another OTP", 60-int(timeSinceCreation.Seconds())),
			}, nil
		}
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
	otpModel := models.NewOTP(phoneNumber, otp, 5) // 5 minutes expiry
	otpModel.Code = "000000"                       // TODO: For testing purposes, override with fixed OTP
	err = cs.DBClient.StoreOTP(ctx, otpModel)
	if err != nil {
		return &pb.RequestOtpResponse{
			Success: false,
			Message: "Failed to store OTP",
		}, nil
	}

	err = cs.MessagingClient.SendOTP(ctx, phoneNumber, otp)
	if err != nil {
		return &pb.RequestOtpResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to send OTP via %s", cs.MessagingClient.GetProviderName()),
		}, nil
	}

	return &pb.RequestOtpResponse{
		Success: true,
		Message: "OTP sent successfully",
	}, nil
}

// VerifyOtp handles OTP verification
func (cs *CommonServer) VerifyOtp(ctx context.Context, req *pb.VerifyOtpRequest) (*pb.LoginResponse, error) {
	phoneNumber := req.GetPhoneNumber()
	otp := req.GetOtp()

	// Validate inputs
	if phoneNumber == "" || otp == "" {
		return nil, fmt.Errorf("phone number and OTP are required")
	}

	// Get OTP from database
	storedOTP, err := cs.DBClient.GetOTPByPhoneNumber(ctx, phoneNumber)
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

	// Get or create user based on app type
	var userID string
	var roles []string

	if helpers.IsClinicalApp(ctx) {
		// For clinical apps, user must exist
		user, err := cs.DBClient.GetUserByPhoneNumber(ctx, phoneNumber)
		if err != nil {
			return nil, fmt.Errorf("clinical user not found")
		}
		clinicalUser := user.(clinicalModels.ClinicUser)
		userID = clinicalUser.ID
		roles = clinicalUser.Roles
		if len(roles) == 0 {
			roles = clinicalModels.DefaultClinicUserRoles()
		}
	} else {
		// For consumer apps, get or create user
		user, err := cs.DBClient.GetUserByPhoneNumber(ctx, phoneNumber)
		if err == nil {
			// User exists
			consumerUser := user.(models.User)
			userID = consumerUser.ID
			roles = consumerUser.Roles
			if len(roles) == 0 {
				roles = models.DefaultConsumerRoles()
			}
		} else {
			// Create new consumer user
			newUser := &models.User{
				ID:          uuid.NewString(),
				PhoneNumber: phoneNumber,
				Roles:       models.DefaultConsumerRoles(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			userID, err = cs.DBClient.CreateUser(ctx, newUser)
			if err != nil {
				return nil, fmt.Errorf("failed to create user: %v", err)
			}
			roles = newUser.Roles
		}
	}

	// Get app type from context
	appType := helpers.GetAppTypeFromContext(ctx)

	token, err := helpers.GenerateAccessToken(cs.Config, userID, roles, appType)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	refreshToken, err := helpers.GenerateRefreshToken(cs.Config, userID, roles, appType)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %v", err)
	}

	return &pb.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		Roles:        helpers.StringsToRoles(roles),
		UserId:       userID,
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
