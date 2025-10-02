package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/therehabstreet/podoai/internal/clinical/clients"
	clinicalHandlers "github.com/therehabstreet/podoai/internal/clinical/handlers"
	commonClients "github.com/therehabstreet/podoai/internal/common/clients"
	"github.com/therehabstreet/podoai/internal/common/config"
	commonHandlers "github.com/therehabstreet/podoai/internal/common/handlers"
	consumerClients "github.com/therehabstreet/podoai/internal/consumer/clients"
	consumerHandlers "github.com/therehabstreet/podoai/internal/consumer/handlers"

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

	consumerMongoClient := &consumerClients.MongoDBClient{
		Client: clinicalMongoClient.Client, // Reuse the same connection
	}

	whatsappClient := commonClients.NewWhatsAppClient(
		config.WhatsApp.APIKey,
		config.WhatsApp.APIURL,
		config.WhatsApp.FromPhone,
	)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	// Register clinical server
	clinicalServer := clinicalHandlers.NewClinicalServer(clinicalMongoClient)
	clinicalHandlers.RegisterClinicalServer(grpcServer, clinicalServer)

	// Register common server
	commonServer := commonHandlers.NewCommonServer(commonMongoClient, whatsappClient)
	commonHandlers.RegisterCommonServer(grpcServer, commonServer)

	// Register consumer server
	consumerServer := consumerHandlers.NewConsumerServer(consumerMongoClient)
	consumerHandlers.RegisterConsumerServer(grpcServer, consumerServer)

	log.Println("gRPC server listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
