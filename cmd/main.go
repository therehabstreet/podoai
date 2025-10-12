package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/therehabstreet/podoai/internal/clinical/clients"
	clinicalHandlers "github.com/therehabstreet/podoai/internal/clinical/handlers"
	clinicalMiddleware "github.com/therehabstreet/podoai/internal/clinical/middleware"
	commonClients "github.com/therehabstreet/podoai/internal/common/clients"
	"github.com/therehabstreet/podoai/internal/common/config"
	commonHandlers "github.com/therehabstreet/podoai/internal/common/handlers"
	commonMiddleware "github.com/therehabstreet/podoai/internal/common/middleware"
	consumerClients "github.com/therehabstreet/podoai/internal/consumer/clients"
	consumerHandlers "github.com/therehabstreet/podoai/internal/consumer/handlers"
	consumerMiddleware "github.com/therehabstreet/podoai/internal/consumer/middleware"

	"google.golang.org/grpc"
)

func main() {
	config := config.NewConfig()

	clinicalMongoClient, err := clients.InitClinicalMongoClient("mongodb://localhost:27017")
	if err != nil {
		log.Fatalf("failed to connect to clinical MongoDB: %v", err)
	}

	commonMongoClient, err := commonClients.InitCommonMongoClient("mongodb://localhost:27017")
	if err != nil {
		log.Fatalf("failed to connect to common MongoDB: %v", err)
	}

	consumerMongoClient := &consumerClients.MongoDBClient{
		Client: clinicalMongoClient.Client, // Reuse the same connection
	}

	whatsappClient := commonClients.NewWhatsAppClient(config)

	storageClient, err := commonClients.NewGCSClient(config.GCS)
	if err != nil {
		log.Fatalf("Warning: failed to create GCS client: %v", err)
	}

	// Start periodic cleanup of expired OTPs
	go func() {
		ticker := time.NewTicker(1 * time.Hour) // Cleanup every hour
		defer ticker.Stop()

		for range ticker.C {
			err := commonMongoClient.CleanupExpiredOTPs(context.Background())
			if err != nil {
				log.Printf("Warning: failed to cleanup expired OTPs: %v", err)
			}
		}
	}()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create gRPC server with AuthN + AuthZ middleware
	authN := commonMiddleware.NewAuthNMiddleware(config)
	commonAuthZ := commonMiddleware.NewAuthZMiddleware()
	clinicalAuthZ := clinicalMiddleware.NewAuthZMiddleware()
	consumerAuthZ := consumerMiddleware.NewAuthZMiddleware()

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			authN.UnaryInterceptor(),         // First: Authentication
			commonAuthZ.UnaryInterceptor(),   // Second: Common service authorization
			clinicalAuthZ.UnaryInterceptor(), // Third: Clinical service authorization
			consumerAuthZ.UnaryInterceptor(), // Fourth: Consumer service authorization
		),
	)

	// Register clinical server
	clinicalServer := clinicalHandlers.NewClinicalServer(clinicalMongoClient)
	clinicalHandlers.RegisterClinicalServer(grpcServer, clinicalServer)

	// Register common server
	commonServer := commonHandlers.NewCommonServer(config, commonMongoClient, whatsappClient, storageClient)
	commonHandlers.RegisterCommonServer(grpcServer, commonServer)

	// Register consumer server
	consumerServer := consumerHandlers.NewConsumerServer(consumerMongoClient)
	consumerHandlers.RegisterConsumerServer(grpcServer, consumerServer)

	log.Println("gRPC server listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
