package middleware

import (
	"context"
	"strings"

	commonMiddleware "github.com/therehabstreet/podoai/internal/common/middleware"
	pb "github.com/therehabstreet/podoai/proto/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthZMiddleware provides authorization for ClinicalService methods
type AuthZMiddleware struct{}

// NewAuthZMiddleware creates a new clinical service authorization middleware
func NewAuthZMiddleware() *AuthZMiddleware {
	return &AuthZMiddleware{}
}

// UnaryInterceptor creates a gRPC unary interceptor for ClinicalService authorization
func (am *AuthZMiddleware) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Only authorize ClinicalService methods
		if !strings.HasPrefix(info.FullMethod, "/podoai_clinical.ClinicalService/") {
			return handler(ctx, req)
		}

		// Get user data from context (set by authentication middleware)
		userRoles, ok := commonMiddleware.GetRolesFromContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "missing user roles in context")
		}

		ownerEntityID, ok := commonMiddleware.GetOwnerEntityIDFromContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "missing owner entity in context")
		}

		// Perform authorization check
		if err := am.authorize(ctx, info.FullMethod, req, userRoles, ownerEntityID); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

// authorize performs authorization checks for ClinicalService methods
func (am *AuthZMiddleware) authorize(ctx context.Context, method string, req any, userRoles []string, ownerEntityID string) error {
	switch method {
	// Clinic operations
	case "/podoai_clinical.ClinicalService/GetClinic":
		// Clinic staff and admins can view clinic info
		if am.hasRole(userRoles, pb.Role_CLINIC_STAFF.String()) || am.hasRole(userRoles, pb.Role_CLINIC_ADMIN.String()) {
			// TODO: Add clinic-specific validation once proto messages are confirmed
			// Should validate that the requested clinic_id matches the user's clinic
			return nil
		}
		return status.Errorf(codes.PermissionDenied, "insufficient permissions to access clinic data")

	case "/podoai_clinical.ClinicalService/UpdateClinic":
		// Only clinic admins can update clinic info
		if am.hasRole(userRoles, pb.Role_CLINIC_ADMIN.String()) {
			// TODO: Add clinic-specific validation once proto messages are confirmed
			// Should validate that the clinic_id in request matches the user's clinic
			return nil
		}
		return status.Errorf(codes.PermissionDenied, "only clinic admins can update clinic information")

	// Clinic User management (CRUDL)
	case "/podoai_clinical.ClinicalService/CreateClinicUser":
		// Only clinic admins can manage clinic users (for their own clinic)
		if am.hasRole(userRoles, pb.Role_CLINIC_ADMIN.String()) {
			// TODO: Add clinic user-specific validation once proto messages are confirmed
			// Should validate that the clinic_id in request matches the user's clinic
			return nil
		}
		return status.Errorf(codes.PermissionDenied, "only clinic admins can create clinic users")

	case "/podoai_clinical.ClinicalService/GetClinicUser":
		// Clinic staff can view clinic users, but only within their clinic
		if am.hasRole(userRoles, pb.Role_CLINIC_STAFF.String()) || am.hasRole(userRoles, pb.Role_CLINIC_ADMIN.String()) {
			// TODO: Add clinic user-specific validation once proto messages are confirmed
			// Should validate that the clinic_id in request matches the user's clinic
			return nil
		}
		return status.Errorf(codes.PermissionDenied, "insufficient permissions to view clinic users")

	case "/podoai_clinical.ClinicalService/UpdateClinicUser":
		// Only clinic admins can update clinic users (for their own clinic)
		if am.hasRole(userRoles, pb.Role_CLINIC_ADMIN.String()) {
			// TODO: Add clinic user-specific validation once proto messages are confirmed
			// Should validate that the clinic_id in request matches the user's clinic
			return nil
		}
		return status.Errorf(codes.PermissionDenied, "only clinic admins can update clinic users")

	case "/podoai_clinical.ClinicalService/DeleteClinicUser":
		// Only clinic admins can delete clinic users (for their own clinic)
		if am.hasRole(userRoles, pb.Role_CLINIC_ADMIN.String()) {
			// TODO: Add clinic user-specific validation once proto messages are confirmed
			// Should validate that the clinic_id in request matches the user's clinic
			return nil
		}
		return status.Errorf(codes.PermissionDenied, "only clinic admins can delete clinic users")

	case "/podoai_clinical.ClinicalService/ListClinicUsers":
		// Clinic staff can list clinic users, but only within their clinic
		if am.hasRole(userRoles, pb.Role_CLINIC_STAFF.String()) || am.hasRole(userRoles, pb.Role_CLINIC_ADMIN.String()) {
			// TODO: Add clinic user-specific validation once proto messages are confirmed
			// Should validate that the clinic_id in request matches the user's clinic
			return nil
		}
		return status.Errorf(codes.PermissionDenied, "insufficient permissions to list clinic users")

	default:
		// For any unknown ClinicalService methods, deny access by default
		return status.Errorf(codes.PermissionDenied, "access denied for method: %s", method)
	}
}

// hasRole checks if the user has a specific role
func (am *AuthZMiddleware) hasRole(userRoles []string, role string) bool {
	for _, userRole := range userRoles {
		if strings.EqualFold(userRole, role) {
			return true
		}
	}
	return false
}
