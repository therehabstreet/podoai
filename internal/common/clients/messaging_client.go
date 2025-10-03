package clients

import (
	"context"
	"fmt"

	"github.com/therehabstreet/podoai/internal/common/config"
)

// MessagingClient interface for sending OTP via different channels
type MessagingClient interface {
	SendOTP(ctx context.Context, phoneNumber, otp string) error
	GetProviderName() string
}

// WhatsAppClient implements OTPSender for WhatsApp
type WhatsAppClient struct {
	apiKey    string
	apiURL    string
	fromPhone string
}

// NewWhatsAppClient creates a new WhatsApp OTP sender
func NewWhatsAppClient(config *config.Config) *WhatsAppClient {
	return &WhatsAppClient{
		apiKey:    config.WhatsApp.APIKey,
		apiURL:    config.WhatsApp.APIURL,
		fromPhone: config.WhatsApp.FromPhone,
	}
}

// SendOTP sends OTP via WhatsApp
func (w *WhatsAppClient) SendOTP(ctx context.Context, phoneNumber, otp string) error {
	// TODO: Implement actual WhatsApp API integration
	// This is a stub implementation

	message := fmt.Sprintf("Your PodoAI verification code is: %s\n\nThis code will expire in 5 minutes. Do not share this code with anyone.", otp)

	// TODO: Format phone number properly (add country code if needed)
	formattedPhone := formatPhoneNumber(phoneNumber)

	// TODO: Make HTTP request to WhatsApp Business API
	// Example structure:
	// payload := map[string]interface{}{
	//     "messaging_product": "whatsapp",
	//     "to": formattedPhone,
	//     "type": "text",
	//     "text": map[string]string{
	//         "body": message,
	//     },
	// }

	// For now, just log the message
	fmt.Printf("[WhatsApp OTP] Sending to %s: %s\n", formattedPhone, message)

	// TODO: Handle WhatsApp API response and errors
	// TODO: Implement retry logic for failed sends
	// TODO: Track delivery status

	return nil
}

// GetProviderName returns the provider name
func (w *WhatsAppClient) GetProviderName() string {
	return "WhatsApp"
}

// formatPhoneNumber formats phone number for WhatsApp API
func formatPhoneNumber(phoneNumber string) string {
	// TODO: Implement proper phone number formatting
	// Remove non-digit characters and ensure proper country code
	// For now, return as-is
	return phoneNumber
}

// validatePhoneNumber validates if phone number is suitable for WhatsApp
func (w *WhatsAppClient) validatePhoneNumber(phoneNumber string) error {
	// TODO: Implement phone number validation
	// Check if number is valid for WhatsApp messaging
	if phoneNumber == "" {
		return fmt.Errorf("phone number cannot be empty")
	}

	// TODO: Add more validation logic
	// - Check if number has proper country code
	// - Validate number format
	// - Check if number is WhatsApp enabled

	return nil
}

// Additional helper methods for WhatsApp integration

// buildWhatsAppPayload creates the payload for WhatsApp API
func (w *WhatsAppClient) buildWhatsAppPayload(phoneNumber, message string) map[string]interface{} {
	// TODO: Implement proper payload structure for WhatsApp Business API
	return map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                phoneNumber,
		"type":              "text",
		"text": map[string]string{
			"body": message,
		},
	}
}

// handleWhatsAppResponse processes the response from WhatsApp API
func (w *WhatsAppClient) handleWhatsAppResponse(response []byte) error {
	// TODO: Parse WhatsApp API response
	// TODO: Handle different response scenarios
	// TODO: Extract message ID for tracking
	return nil
}
