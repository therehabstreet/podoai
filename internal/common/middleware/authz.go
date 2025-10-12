package middleware

import (
	"context"
	"strings"

	pb "github.com/therehabstreet/podoai/proto/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthZMiddleware provides authorization for CommonService methods
type AuthZMiddleware struct{}

// NewAuthZMiddleware creates a new common service authorization middleware
func NewAuthZMiddleware() *AuthZMiddleware {
	return &AuthZMiddleware{}
}

// UnaryInterceptor creates a gRPC unary interceptor for CommonService authorization
func (am *AuthZMiddleware) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Only authorize CommonService methods
		if !strings.HasPrefix(info.FullMethod, "/podoai.CommonService/") {
			return handler(ctx, req)
		}

		// Skip auth for certain methods
		if am.shouldSkipAuth(info.FullMethod) {
			return handler(ctx, req)
		}

		// Get user data from context (set by authentication middleware)
		userRoles, ok := GetRolesFromContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "missing user roles in context")
		}

		ownerEntityID, ok := GetOwnerEntityIDFromContext(ctx)
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

// shouldSkipAuth determines if authorization should be skipped for a method
func (am *AuthZMiddleware) shouldSkipAuth(method string) bool {
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

// authorize performs authorization checks for CommonService methods
func (am *AuthZMiddleware) authorize(ctx context.Context, method string, req any, userRoles []string, ownerEntityID string) error {
	switch method {
	// Products, Exercises, Therapies (open to all authenticated users)
	case "/podoai.CommonService/GetProduct",
		"/podoai.CommonService/GetExercise",
		"/podoai.CommonService/GetTherapy":
		return nil

	// Scan operations
	case "/podoai.CommonService/CreateScan":
		if r, ok := req.(*pb.CreateScanRequest); ok {
			if r.GetScan() != nil && r.GetScan().GetOwnerEntityId() == ownerEntityID {
				return nil
			}
		}
		return status.Errorf(codes.PermissionDenied, "unauthorized to create scans for different owner entity")

	case "/podoai.CommonService/GetScan":
		if r, ok := req.(*pb.GetScanRequest); ok {
			if r.GetOwnerEntityId() == ownerEntityID {
				return nil
			}
		}
		return status.Errorf(codes.PermissionDenied, "unauthorized to access scans for different owner entity")

	case "/podoai.CommonService/GetScans":
		if r, ok := req.(*pb.GetScansRequest); ok {
			if r.GetOwnerEntityId() == ownerEntityID {
				return nil
			}
		}
		return status.Errorf(codes.PermissionDenied, "unauthorized to access scans for different owner entity")

	case "/podoai.CommonService/DeleteScan":
		// Only clinic admins can delete scans from their clinic
		if am.hasRole(userRoles, pb.Role_CLINIC_ADMIN.String()) {
			if r, ok := req.(*pb.DeleteScanRequest); ok {
				if r.GetOwnerEntityId() == ownerEntityID {
					return nil
				}
			}
		}
		return status.Errorf(codes.PermissionDenied, "only clinic admins can delete scans")

	case "/podoai.CommonService/GenerateMediaSignedUrls":
		if r, ok := req.(*pb.GenerateMediaSignedUrlsRequest); ok {
			if r.GetOwnerEntityId() == ownerEntityID {
				return nil
			}
		}
		return status.Errorf(codes.PermissionDenied, "unauthorized to access media for different owner entity")

	// Patient operations
	case "/podoai.CommonService/GetPatients":
		if r, ok := req.(*pb.GetPatientsRequest); ok {
			if r.GetOwnerEntityId() == ownerEntityID {
				return nil
			}
		}
		return status.Errorf(codes.PermissionDenied, "unauthorized to access patients for different owner entity")

	case "/podoai.CommonService/GetPatient":
		if r, ok := req.(*pb.GetPatientRequest); ok {
			if r.GetOwnerEntityId() == ownerEntityID {
				return nil
			}
		}
		return status.Errorf(codes.PermissionDenied, "unauthorized to access patient for different owner entity")

	case "/podoai.CommonService/SearchPatient":
		if r, ok := req.(*pb.SearchPatientRequest); ok {
			if r.GetOwnerEntityId() == ownerEntityID {
				return nil
			}
		}
		return status.Errorf(codes.PermissionDenied, "unauthorized to search patients for different owner entity")

	case "/podoai.CommonService/CreatePatient":
		if r, ok := req.(*pb.CreatePatientRequest); ok {
			if r.GetPatient() != nil && r.GetPatient().GetOwnerEntityId() == ownerEntityID {
				return nil
			}
		}
		return status.Errorf(codes.PermissionDenied, "unauthorized to create patient for different owner entity")

	case "/podoai.CommonService/DeletePatient":
		if r, ok := req.(*pb.DeletePatientRequest); ok {
			if r.GetOwnerEntityId() == ownerEntityID {
				return nil
			}
		}
		return status.Errorf(codes.PermissionDenied, "unauthorized to delete patient for different owner entity")

	default:
		// For any unknown CommonService methods, deny access by default
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
