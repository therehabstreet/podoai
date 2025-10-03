package handlers

import (
	"github.com/therehabstreet/podoai/internal/common/clients"
	"github.com/therehabstreet/podoai/internal/common/config"
	pb "github.com/therehabstreet/podoai/proto/common"
	"google.golang.org/grpc"
)

type CommonServer struct {
	pb.UnimplementedCommonServiceServer
	Config          *config.Config
	DBClient        clients.DBClient
	MessagingClient clients.MessagingClient
}

func NewCommonServer(cfg *config.Config, dbClient clients.DBClient, messagingClient clients.MessagingClient) *CommonServer {
	return &CommonServer{
		Config:          cfg,
		DBClient:        dbClient,
		MessagingClient: messagingClient,
	}
}

func RegisterCommonServer(grpcServer *grpc.Server, server *CommonServer) {
	pb.RegisterCommonServiceServer(grpcServer, server)
}
