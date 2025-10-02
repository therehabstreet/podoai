package handlers

import (
	"github.com/therehabstreet/podoai/internal/consumer/clients"
	pb "github.com/therehabstreet/podoai/proto/consumer"
	"google.golang.org/grpc"
)

type ConsumerServer struct {
	pb.UnimplementedConsumerServiceServer
	DBClient clients.DBClient
}

func NewConsumerServer(dbClient clients.DBClient) *ConsumerServer {
	return &ConsumerServer{
		DBClient: dbClient,
	}
}

func RegisterConsumerServer(grpcServer *grpc.Server, server *ConsumerServer) {
	pb.RegisterConsumerServiceServer(grpcServer, server)
}
