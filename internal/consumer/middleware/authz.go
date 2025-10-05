package middleware

import (
	"context"
	"strings"

	commonMiddleware "github.com/therehabstreet/podoai/internal/common/middleware"
	pbCommon "github.com/therehabstreet/podoai/proto/common"
	pb "github.com/therehabstreet/podoai/proto/consumer"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthZMiddleware provides authorization for ConsumerService methods
type AuthZMiddleware struct{}

// NewAuthZMiddleware creates a new consumer service authorization middleware
func NewAuthZMiddleware() *AuthZMiddleware {
	return &AuthZMiddleware{}
}

// UnaryInterceptor creates a gRPC unary interceptor for ConsumerService authorization
func (am *AuthZMiddleware) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Only authorize ConsumerService methods
		if !strings.HasPrefix(info.FullMethod, "/podoai_consumer.ConsumerService/") {
			return handler(ctx, req)
		}

		// Get user data from context (set by authentication middleware)
		userRoles, ok := commonMiddleware.GetRolesFromContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "missing user roles in context")
		}

		userID, ok := commonMiddleware.GetUserIDFromContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "missing user ID in context")
		}

		// Perform authorization check
		if err := am.authorize(ctx, info.FullMethod, req, userRoles, userID); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

// authorize performs authorization checks for ConsumerService methods
func (am *AuthZMiddleware) authorize(ctx context.Context, method string, req any, userRoles []string, userID string) error {
	switch method {
	// User operations
	case "/podoai_consumer.ConsumerService/GetUser":
		// Consumers can only access their own data
		if am.hasRole(userRoles, pbCommon.Role_CONSUMER.String()) {
			if r, ok := req.(*pb.GetUserRequest); ok {
				if r.GetUserId() == userID {
					return nil
				}
			}
		}
		return status.Errorf(codes.PermissionDenied, "only consumers can access user data")

	case "/podoai_consumer.ConsumerService/UpdateUser":
		// Consumers can only update their own data
		if am.hasRole(userRoles, pbCommon.Role_CONSUMER.String()) {
			if r, ok := req.(*pb.UpdateUserRequest); ok {
				if r.GetUser() != nil && r.GetUser().GetId() == userID {
					return nil
				}
			}
		}
		return status.Errorf(codes.PermissionDenied, "only consumers can update their own data")

	case "/podoai_consumer.ConsumerService/CreateUser":
		return status.Errorf(codes.PermissionDenied, "unauthorized to create user")

	case "/podoai_consumer.ConsumerService/DeleteUser":
		return status.Errorf(codes.PermissionDenied, "consumers can not delete their account")

	default:
		// For any unknown ConsumerService methods, deny access by default
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
