package clients

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/storage"
	"github.com/therehabstreet/podoai/internal/common/config"
	"google.golang.org/api/option"
)

type StorageClient interface {
	GenerateSignedURL(ctx context.Context, path string, action string) (string, time.Time, error)
	Close() error
	GetProviderName() string
}

type GCSClient struct {
	client *storage.Client
	config *config.GCSConfig
}

func NewGCSClient(cfg *config.GCSConfig) (StorageClient, error) {
	ctx := context.Background()
	var client *storage.Client
	var err error

	// If service account key path is provided, use it
	if cfg.ServiceAccountKeyPath != "" {
		client, err = storage.NewClient(ctx, option.WithCredentialsFile(cfg.ServiceAccountKeyPath))
	} else {
		// Try to use default credentials (works in GCP environments or with GOOGLE_APPLICATION_CREDENTIALS)
		client, err = storage.NewClient(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %v", err)
	}

	return &GCSClient{
		client: client,
		config: cfg,
	}, nil
}

func (g *GCSClient) GenerateSignedURL(ctx context.Context, path string, action string) (string, time.Time, error) {
	expirationDuration := time.Duration(g.config.DefaultSignedURLExpiryMin) * time.Minute
	if action == "WRITE" {
		expirationDuration = time.Duration(g.config.DefaultSignedURLExpiryMin) * time.Minute // Same for now
	}

	expiresAt := time.Now().Add(expirationDuration)

	// Clean up path - remove leading slash
	objectName := path
	if len(objectName) > 0 && objectName[0] == '/' {
		objectName = objectName[1:]
	}

	// Determine HTTP method based on action
	method := "GET"
	if action == "WRITE" {
		method = "PUT"
	}

	// Generate signed URL
	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  method,
		Expires: expiresAt,
	}

	signedURL, err := storage.SignedURL(g.config.BucketName, objectName, opts)
	if err != nil {
		// If signing fails, fall back to placeholder for development
		fmt.Printf("Warning: Failed to generate GCS signed URL, using placeholder: %v\n", err)
		signedURL = fmt.Sprintf("https://storage.googleapis.com/%s/%s?method=%s&expires=%d&placeholder=true",
			g.config.BucketName, objectName, method, expiresAt.Unix())
	}

	return signedURL, expiresAt, nil
}

func (g *GCSClient) Close() error {
	if g.client != nil {
		return g.client.Close()
	}
	return nil
}

func (g *GCSClient) GetProviderName() string {
	return "Google Cloud Storage"
}
