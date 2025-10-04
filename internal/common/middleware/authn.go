package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/therehabstreet/podoai/internal/common/config"
	"github.com/therehabstreet/podoai/internal/common/helpers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthNMiddleware provides JWT-based authentication (token validation only)
type AuthNMiddleware struct {
	config *config.Config
}

// NewAuthNMiddleware creates a new authentication middleware instance
func NewAuthNMiddleware(cfg *config.Config) *AuthNMiddleware {
	return &AuthNMiddleware{
		config: cfg,
	}
}

// UnaryInterceptor creates a gRPC unary interceptor for JWT authentication
func (am *AuthNMiddleware) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip auth for certain methods
		if am.shouldSkipAuth(info.FullMethod) {
			return handler(ctx, req)
		}

		// Extract token from metadata
		token, err := am.extractTokenFromContext(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "missing or invalid token: %v", err)
		}

		// Validate token
		claims, err := helpers.ValidateToken(am.config, token)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		// Add claims to context
		ctx = am.addClaimsToContext(ctx, claims)

		// Continue to next handler (authorization will be handled by service-specific interceptors)
		return handler(ctx, req)
	}
}

// shouldSkipAuth determines if authentication should be skipped for a method
func (am *AuthNMiddleware) shouldSkipAuth(method string) bool {
	skipMethods := []string{
		"/podoai.CommonService/RequestOtp",
		"/podoai.CommonService/VerifyOtp",
	}

	for _, skipMethod := range skipMethods {
		if method == skipMethod {
			return true
		}
	}
	return false
}

// extractTokenFromContext extracts JWT token from gRPC metadata
func (am *AuthNMiddleware) extractTokenFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("missing metadata")
	}

	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		return "", fmt.Errorf("missing authorization header")
	}

	// Check for Bearer token format
	token := authHeader[0]
	if strings.HasPrefix(token, "Bearer ") {
		return strings.TrimPrefix(token, "Bearer "), nil
	}

	return token, nil
}

// addClaimsToContext adds JWT claims to the context
func (am *AuthNMiddleware) addClaimsToContext(ctx context.Context, claims *helpers.JWTClaims) context.Context {
	ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
	ctx = context.WithValue(ctx, RolesKey, claims.Roles)
	ctx = context.WithValue(ctx, TokenTypeKey, claims.TokenType)
	ctx = context.WithValue(ctx, OwnerEntityIDKey, claims.OwnerEntityID)
	return ctx
}
