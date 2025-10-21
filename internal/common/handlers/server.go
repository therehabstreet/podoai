package handlers

import (
	"github.com/therehabstreet/podoai/internal/common/clients"
	"github.com/therehabstreet/podoai/internal/common/config"
	"github.com/therehabstreet/podoai/internal/common/workers"
	pb "github.com/therehabstreet/podoai/proto/common"
	"google.golang.org/grpc"
)

type CommonServer struct {
	pb.UnimplementedCommonServiceServer
	Config             *config.Config
	DBClient           clients.DBClient
	MessagingClient    clients.MessagingClient
	StorageClient      clients.StorageClient
	ScanResultWorkflow *workers.WorkflowEngine
}

func NewCommonServer(cfg *config.Config, dbClient clients.DBClient, messagingClient clients.MessagingClient, storageClient clients.StorageClient, scanResultWorkflow *workers.WorkflowEngine) *CommonServer {
	return &CommonServer{
		Config:             cfg,
		DBClient:           dbClient,
		MessagingClient:    messagingClient,
		StorageClient:      storageClient,
		ScanResultWorkflow: scanResultWorkflow,
	}
}

func RegisterCommonServer(grpcServer *grpc.Server, server *CommonServer) {
	pb.RegisterCommonServiceServer(grpcServer, server)
}
