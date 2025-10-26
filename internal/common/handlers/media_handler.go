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
	action := req.GetAction()

	if scanID == "" {
		return nil, fmt.Errorf("missing scan ID")
	}

	var actionString string
	switch action {
	case pb.SignedUrlAction_READ:
		actionString = action.String()
	case pb.SignedUrlAction_WRITE:
		actionString = action.String()
	default:
		return nil, fmt.Errorf("invalid url sign action specified")
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
		signedURL, expiresAt, err := cs.generateGCSSignedURL(gcsPath, actionString)
		if err != nil {
			return nil, fmt.Errorf("failed to generate signed URL for %s: %v", typeString, err)
		}

		thumbnailURL, _, err := cs.generateGCSSignedURL(thumbnailPath, actionString)
		if err != nil {
			return nil, fmt.Errorf("failed to generate thumbnail URL for %s: %v", typeString, err)
		}

		images = append(images, &pb.Image{
			Type:               imageType,
			SignedUrl:          signedURL,
			ThumbnailSignedUrl: thumbnailURL,
			Path:               gcsPath,
			ThumbnailPath:      thumbnailPath,
			ExpiresAt:          timestamppb.New(expiresAt),
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
		signedURL, expiresAt, err := cs.generateGCSSignedURL(gcsPath, actionString)
		if err != nil {
			return nil, fmt.Errorf("failed to generate signed URL for %s: %v", typeString, err)
		}

		thumbnailURL, _, err := cs.generateGCSSignedURL(thumbnailPath, actionString)
		if err != nil {
			return nil, fmt.Errorf("failed to generate thumbnail URL for %s: %v", typeString, err)
		}

		videos = append(videos, &pb.Video{
			Type:               videoType,
			SignedUrl:          signedURL,
			ThumbnailSignedUrl: thumbnailURL,
			Path:               gcsPath,
			ThumbnailPath:      thumbnailPath,
			ExpiresAt:          timestamppb.New(expiresAt),
		})
	}

	return &pb.GenerateMediaSignedUrlsResponse{
		Images: images,
		Videos: videos,
	}, nil
}

// generateGCSSignedURL generates a signed URL for the given GCS path
func (cs *CommonServer) generateGCSSignedURL(gcsPath string, action string) (string, time.Time, error) {
	ctx := context.Background()
	return cs.StorageClient.GenerateSignedURL(ctx, gcsPath, action)
}
