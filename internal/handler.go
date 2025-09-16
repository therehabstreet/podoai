package podoai

import (
	"context"

	proto "github.com/therehabstreet/podoai/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements the PodoAIServiceServer interface.
type Server struct {
	proto.UnimplementedPodoAIServiceServer
}

func NewServer() *Server {
	return &Server{} 
}

func (s *Server) RequestOtp(
	ctx context.Context,
	req *proto.RequestOtpRequest,
) (*proto.RequestOtpResponse, error) {
	// TODO: implement your logic here
	return nil, status.Errorf(codes.Unimplemented, "method RequestOtp not implemented")
}

func (s *Server) VerifyOtp(ctx context.Context, req *proto.VerifyOtpRequest) (*proto.LoginResponse, error) {
	// TODO: implement your logic here
	return nil, status.Errorf(codes.Unimplemented, "method VerifyOtp not implemented")
}

func (s *Server) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	// TODO: implement your logic here
	return nil, status.Errorf(codes.Unimplemented, "method Login not implemented")
}

func (s *Server) ValidateToken(ctx context.Context, req *proto.ValidateTokenRequest) (*proto.ValidateTokenResponse, error) {
	// TODO: implement your logic here
	return nil, status.Errorf(codes.Unimplemented, "method ValidateToken not implemented")
}
