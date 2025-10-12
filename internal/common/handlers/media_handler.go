package handlers

import (
	"context"
	"fmt"
	"time"

	pb "github.com/therehabstreet/podoai/proto/common"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GenerateMediaSignedUrls handles the GenerateMediaSignedUrls gRPC request
// Returns hardcoded URLs separated into images and videos
func (cs *CommonServer) GenerateMediaSignedUrls(ctx context.Context, req *pb.GenerateMediaSignedUrlsRequest) (*pb.GenerateMediaSignedUrlsResponse, error) {
	scanID := req.GetScanId()
	ownerEntityID := req.GetOwnerEntityId()

	if scanID == "" {
		return nil, fmt.Errorf("missing scan ID")
	}
	if ownerEntityID == "" {
		return nil, fmt.Errorf("missing owner entity ID")
	}

	// Verify the scan exists and is owned by the owner entity
	_, err := cs.DBClient.FetchScanByID(ctx, scanID, ownerEntityID)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch scan: %v", err)
	}

	// Hardcoded image types to return
	imageTypes := []pb.ImageType{
		pb.ImageType_LEFT_FOOT_DORSAL,
		pb.ImageType_LEFT_FOOT_MEDIAL_SIDE,
		pb.ImageType_RIGHT_FOOT_DORSAL,
		pb.ImageType_RIGHT_FOOT_MEDIAL_SIDE,
	}

	// Hardcoded video types to return
	videoTypes := []pb.VideoType{
		pb.VideoType_GAIT_POSTERIOR,
	}

	var images []*pb.Image
	var videos []*pb.Video

	// Generate signed URLs for images
	for _, imageType := range imageTypes {
		// Convert enum to string for file naming
		typeString := imageType.String()

		// Build GCS paths
		gcsPath := fmt.Sprintf("/scans/%s/%s/media/%s.jpg", ownerEntityID, scanID, typeString)
		thumbnailPath := fmt.Sprintf("/scans/%s/%s/media/thumbnails/%s.jpg", ownerEntityID, scanID, typeString)

		// Generate signed URLs
		signedURL, expiresAt, err := cs.generateGCSSignedURL(gcsPath, "READ")
		if err != nil {
			return nil, fmt.Errorf("failed to generate signed URL for %s: %v", typeString, err)
		}

		thumbnailURL, _, err := cs.generateGCSSignedURL(thumbnailPath, "READ")
		if err != nil {
			return nil, fmt.Errorf("failed to generate thumbnail URL for %s: %v", typeString, err)
		}

		images = append(images, &pb.Image{
			Type:          imageType,
			SignedUrl:     signedURL,
			ThumbnailUrl:  thumbnailURL,
			GcsPath:       gcsPath,
			ThumbnailPath: thumbnailPath,
			ExpiresAt:     timestamppb.New(expiresAt),
		})
	}

	// Generate signed URLs for videos
	for _, videoType := range videoTypes {
		// Convert enum to string for file naming
		typeString := videoType.String()

		// Build GCS paths
		gcsPath := fmt.Sprintf("/scans/%s/%s/media/%s.mp4", ownerEntityID, scanID, typeString)
		thumbnailPath := fmt.Sprintf("/scans/%s/%s/media/thumbnails/%s.jpg", ownerEntityID, scanID, typeString)

		// Generate signed URLs
		signedURL, expiresAt, err := cs.generateGCSSignedURL(gcsPath, "READ")
		if err != nil {
			return nil, fmt.Errorf("failed to generate signed URL for %s: %v", typeString, err)
		}

		thumbnailURL, _, err := cs.generateGCSSignedURL(thumbnailPath, "READ")
		if err != nil {
			return nil, fmt.Errorf("failed to generate thumbnail URL for %s: %v", typeString, err)
		}

		videos = append(videos, &pb.Video{
			Type:          videoType,
			SignedUrl:     signedURL,
			ThumbnailUrl:  thumbnailURL,
			GcsPath:       gcsPath,
			ThumbnailPath: thumbnailPath,
			ExpiresAt:     timestamppb.New(expiresAt),
		})
	}

	return &pb.GenerateMediaSignedUrlsResponse{
		Images: images,
		Videos: videos,
	}, nil
}

// generateGCSSignedURL generates a signed URL for the given GCS path
func (cs *CommonServer) generateGCSSignedURL(gcsPath string, action string) (string, time.Time, error) {
	// Configuration (should come from environment/config)
	bucketName := "podoai-scans" // TODO: Make this configurable
	expirationDuration := 15 * time.Minute

	if action == "WRITE" {
		expirationDuration = time.Minute * 15 // Shorter expiration for uploads
	}

	expiresAt := time.Now().Add(expirationDuration)

	// Placeholder implementation
	method := "GET"
	if action == "WRITE" {
		method = "PUT"
	}

	signedURL := fmt.Sprintf("https://storage.googleapis.com/%s%s?method=%s&expires=%d&placeholder=true",
		bucketName, gcsPath, method, expiresAt.Unix())

	return signedURL, expiresAt, nil
}
