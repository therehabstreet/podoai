package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
	// Update the import path below to the correct location of your generated proto files, for example:
	"github.com/therehabstreet/podoai/internal"
	pb "github.com/therehabstreet/podoai/proto"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	server := podoai.NewServer()
	pb.RegisterPodoAIServiceServer(grpcServer, server)

	log.Println("gRPC server listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
