package handlers

import (
	"github.com/therehabstreet/podoai/internal/clinical/clients"
	pb "github.com/therehabstreet/podoai/proto/clinical"
	"google.golang.org/grpc"
)

type ClinicalServer struct {
	pb.UnimplementedClinicalServiceServer
	DBClient clients.DBClient
}

func NewClinicalServer(dbClient clients.DBClient) *ClinicalServer {
	return &ClinicalServer{
		DBClient: dbClient,
	}
}

func RegisterClinicalServer(grpcServer *grpc.Server, server *ClinicalServer) {
	pb.RegisterClinicalServiceServer(grpcServer, server)
}
