package handlers

import (
	"github.com/therehabstreet/podoai/internal/common/clients"
	pb "github.com/therehabstreet/podoai/proto/common"
	"google.golang.org/grpc"
)

type CommonServer struct {
	pb.UnimplementedCommonServiceServer
	DBClient  clients.DBClient
	OTPSender clients.MessagingClient
}

func NewCommonServer(dbClient clients.DBClient, otpSender clients.MessagingClient) *CommonServer {
	return &CommonServer{
		DBClient:  dbClient,
		OTPSender: otpSender,
	}
}

func RegisterCommonServer(grpcServer *grpc.Server, server *CommonServer) {
	pb.RegisterCommonServiceServer(grpcServer, server)
}
