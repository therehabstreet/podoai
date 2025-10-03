package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/therehabstreet/podoai/internal/common/config"
	"github.com/therehabstreet/podoai/internal/common/helpers"
	pb "github.com/therehabstreet/podoai/proto/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Context key types to avoid collisions
type contextKey string

const (
	UserIDKey        contextKey = "user_id"
	RolesKey         contextKey = "roles"
	TokenTypeKey     contextKey = "token_type"
	OwnerEntityIDKey contextKey = "owner_entity_id"
)

// AuthMiddleware provides JWT-based authentication and authorization
type AuthMiddleware struct {
	config *config.Config
}

// NewAuthMiddleware creates a new auth middleware instance
func NewAuthMiddleware(cfg *config.Config) *AuthMiddleware {
	return &AuthMiddleware{
		config: cfg,
	}
}

// UnaryInterceptor creates a gRPC unary interceptor for JWT authentication and authorization
func (am *AuthMiddleware) UnaryInterceptor() grpc.UnaryServerInterceptor {
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

		// Perform authorization check
		if err := am.authorize(ctx, info.FullMethod, req, claims); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

// RequireRole creates a middleware that requires specific roles
func (am *AuthMiddleware) RequireRole(requiredRoles ...pb.Role) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Extract token from metadata
		token, err := am.extractTokenFromContext(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "missing or invalid token: %v", err)
		}

		// Check if user has any of the required roles
		if !helpers.HasAnyRole(am.config, token, requiredRoles) {
			return nil, status.Errorf(codes.PermissionDenied, "insufficient permissions: requires one of %v", requiredRoles)
		}

		return handler(ctx, req)
	}
}

// shouldSkipAuth determines if authentication should be skipped for a method
func (am *AuthMiddleware) shouldSkipAuth(method string) bool {
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
func (am *AuthMiddleware) extractTokenFromContext(ctx context.Context) (string, error) {
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
func (am *AuthMiddleware) addClaimsToContext(ctx context.Context, claims *helpers.JWTClaims) context.Context {
	ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
	ctx = context.WithValue(ctx, RolesKey, claims.Roles)
	ctx = context.WithValue(ctx, TokenTypeKey, claims.TokenType)
	ctx = context.WithValue(ctx, OwnerEntityIDKey, claims.OwnerEntityID)
	return ctx
}

// authorize performs authorization checks based on method and user roles
func (am *AuthMiddleware) authorize(ctx context.Context, method string, req any, claims *helpers.JWTClaims) error {
	userRoles := claims.Roles
	ownerEntityID := claims.OwnerEntityID

	switch method {
	// Common Service - Products, Exercises, Therapies (open to all authenticated users)
	case "/podoai.CommonService/GetProduct",
		"/podoai.CommonService/GetExercise",
		"/podoai.CommonService/GetTherapy":
		return nil

	// Common Service - Scan operations
	case "/podoai.CommonService/CreateScan":
		if r, ok := req.(*pb.CreateScanRequest); ok {
			if r.GetScan() != nil && r.GetScan().GetOwnerEntityId() != ownerEntityID {
				return status.Errorf(codes.PermissionDenied, "unauthorized to create scans")
			}
		}
		return nil

	case "/podoai.CommonService/GetScan":
		if r, ok := req.(*pb.GetScanRequest); ok {
			if r.GetOwnerEntityId() != ownerEntityID {
				return status.Errorf(codes.PermissionDenied, "unauthorized to access scans")
			}
		}
		return nil

	case "/podoai.CommonService/GetScans":
		if r, ok := req.(*pb.GetScansRequest); ok {
			if r.GetOwnerEntityId() != ownerEntityID {
				return status.Errorf(codes.PermissionDenied, "unauthorized to access scans")
			}
		}
		return nil

	case "/podoai.CommonService/DeleteScan":
		// Only clinic admins can delete scans from their clinic
		if am.hasRole(userRoles, pb.Role_CLINIC_ADMIN.String()) {
			if r, ok := req.(*pb.DeleteScanRequest); ok {
				if r.GetOwnerEntityId() != ownerEntityID {
					return status.Errorf(codes.PermissionDenied, "unauthorized to delete scans")
				}
			}
			return nil
		}
		return status.Errorf(codes.PermissionDenied, "unauthorized to delete scan")

	// Common Service - Patient operations
	case "/podoai.CommonService/GetPatients":
		if r, ok := req.(*pb.GetPatientsRequest); ok {
			if r.GetOwnerEntityId() != ownerEntityID {
				return status.Errorf(codes.PermissionDenied, "unauthorized to access patients")
			}
		}
		return nil

	case "/podoai.CommonService/GetPatient":
		if r, ok := req.(*pb.GetPatientRequest); ok {
			if r.GetOwnerEntityId() != ownerEntityID {
				return status.Errorf(codes.PermissionDenied, "unauthorized to access patient")
			}
		}
		return nil

	case "/podoai.CommonService/SearchPatient":
		if r, ok := req.(*pb.SearchPatientRequest); ok {
			if r.GetOwnerEntityId() != ownerEntityID {
				return status.Errorf(codes.PermissionDenied, "unauthorized to search patients")
			}
		}
		return nil

	case "/podoai.CommonService/CreatePatient":
		if r, ok := req.(*pb.CreatePatientRequest); ok {
			if r.GetPatient() != nil && r.GetPatient().GetOwnerEntityId() != ownerEntityID {
				return status.Errorf(codes.PermissionDenied, "unauthorized to create patient")
			}
		}
		return nil

	case "/podoai.CommonService/DeletePatient":
		if r, ok := req.(*pb.DeletePatientRequest); ok {
			if r.GetOwnerEntityId() != ownerEntityID {
				return status.Errorf(codes.PermissionDenied, "unauthorized to delete patient")
			}
		}
		return nil

	// Clinical Service - Clinic operations
	case "/podoai_clinical.ClinicalService/GetClinic":
		// Clinic staff and admins can view clinic info
		if am.hasRole(userRoles, pb.Role_CLINIC_STAFF.String()) || am.hasRole(userRoles, pb.Role_CLINIC_ADMIN.String()) {
			// TODO: Add clinic-specific validation once proto messages are confirmed
			return nil
		}
		return status.Errorf(codes.PermissionDenied, "insufficient permissions to access clinic data")

	// Clinical Service - Clinic User management (CRUDL)
	case "/podoai_clinical.ClinicalService/CreateClinicUser":
		// Only clinic admins can manage clinic users (for their own clinic)
		if am.hasRole(userRoles, pb.Role_CLINIC_ADMIN.String()) {
			// TODO: Add clinic user-specific validation once proto messages are confirmed
			return nil
		}
		return status.Errorf(codes.PermissionDenied, "insufficient permissions to manage clinic users")

	case "/podoai_clinical.ClinicalService/GetClinicUser":
		// Only clinic admins can manage clinic users (for their own clinic)
		if am.hasRole(userRoles, pb.Role_CLINIC_ADMIN.String()) {
			// TODO: Add clinic user-specific validation once proto messages are confirmed
			return nil
		}
		return status.Errorf(codes.PermissionDenied, "insufficient permissions to manage clinic users")

	case "/podoai_clinical.ClinicalService/UpdateClinicUser":
		// Only clinic admins can manage clinic users (for their own clinic)
		if am.hasRole(userRoles, pb.Role_CLINIC_ADMIN.String()) {
			// TODO: Add clinic user-specific validation once proto messages are confirmed
			return nil
		}
		return status.Errorf(codes.PermissionDenied, "insufficient permissions to manage clinic users")

	case "/podoai_clinical.ClinicalService/DeleteClinicUser":
		// Only clinic admins can manage clinic users (for their own clinic)
		if am.hasRole(userRoles, pb.Role_CLINIC_ADMIN.String()) {
			// TODO: Add clinic user-specific validation once proto messages are confirmed
			return nil
		}
		return status.Errorf(codes.PermissionDenied, "insufficient permissions to manage clinic users")

	case "/podoai_clinical.ClinicalService/ListClinicUsers":
		// Only clinic admins can manage clinic users (for their own clinic)
		if am.hasRole(userRoles, pb.Role_CLINIC_ADMIN.String()) {
			// TODO: Add clinic user-specific validation once proto messages are confirmed
			return nil
		}
		return status.Errorf(codes.PermissionDenied, "insufficient permissions to manage clinic users")

	// Consumer Service - User operations
	case "/podoai_consumer.ConsumerService/GetUser":
		// Consumers can only access their own data
		if am.hasRole(userRoles, "consumer") {
			// TODO: Validate that the requested user_id matches the token's user_id once proto messages are confirmed
			return nil
		}
		return status.Errorf(codes.PermissionDenied, "insufficient permissions to access user data")

	case "/podoai_consumer.ConsumerService/UpdateUser":
		// Consumers can only update their own data
		if am.hasRole(userRoles, "consumer") {
			// TODO: Validate that the user_id in request matches the token's user_id once proto messages are confirmed
			return nil
		}
		return status.Errorf(codes.PermissionDenied, "insufficient permissions to update user data")

	default:
		// For any unknown methods, deny access by default
		return status.Errorf(codes.PermissionDenied, "access denied for method: %s", method)
	}
}

// hasRole checks if the user has a specific role
func (am *AuthMiddleware) hasRole(userRoles []string, role string) bool {
	for _, userRole := range userRoles {
		if strings.EqualFold(userRole, role) {
			return true
		}
	}
	return false
}

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
