package helpers

import (
	"strings"
	"time"

	"github.com/therehabstreet/podoai/internal/common/models"
	"github.com/therehabstreet/podoai/proto/common"
	pb "github.com/therehabstreet/podoai/proto/common"
	podoai "github.com/therehabstreet/podoai/proto/common"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Scan Proto <-> Model conversion
func ScanProtoToModel(proto *common.Scan) models.Scan {
	scan := models.Scan{
		ID:            proto.GetId(),
		PatientID:     proto.GetPatientId(),
		OwnerEntityID: proto.GetOwnerEntityId(),
		CreatedAt:     timestampPBToTime(proto.GetCreatedAt()),
		UpdatedAt:     timestampPBToTime(proto.GetUpdatedAt()),
		ScannedByID:   proto.GetScannedById(),
		ReviewedByID:  proto.GetReviewedById(),
	}

	// Convert images
	for _, img := range proto.GetImages() {
		scan.Images = append(scan.Images, ImageProtoToModel(img))
	}

	// Convert videos
	for _, vid := range proto.GetVideos() {
		scan.Videos = append(scan.Videos, VideoProtoToModel(vid))
	}

	// Convert ScanAIResult
	if proto.GetScanAiResult() != nil {
		scan.ScanAIResult = ScanAIResultProtoToModel(proto.GetScanAiResult())
	}

	// Convert ReviewedAt
	if proto.GetReviewedAt() != nil {
		reviewedAt := timestampPBToTime(proto.GetReviewedAt())
		scan.ReviewedAt = &reviewedAt
	}

	return scan
}

func ScanModelToProto(model models.Scan) *common.Scan {
	scan := &common.Scan{
		Id:            model.ID,
		PatientId:     model.PatientID,
		OwnerEntityId: model.OwnerEntityID,
		CreatedAt:     timestamppb.New(model.CreatedAt),
		UpdatedAt:     timestamppb.New(model.UpdatedAt),
		ScannedById:   model.ScannedByID,
		ReviewedById:  model.ReviewedByID,
	}

	// Convert images
	for _, img := range model.Images {
		scan.Images = append(scan.Images, ImageModelToProto(img))
	}

	// Convert videos
	for _, vid := range model.Videos {
		scan.Videos = append(scan.Videos, VideoModelToProto(vid))
	}

	// Convert ScanAIResult
	if model.ScanAIResult != nil {
		scan.ScanAiResult = ScanAIResultModelToProto(model.ScanAIResult)
	}

	// Convert ReviewedAt
	if model.ReviewedAt != nil {
		scan.ReviewedAt = timestamppb.New(*model.ReviewedAt)
	}

	return scan
}

// ScanAIResult conversion
func ScanAIResultProtoToModel(proto *common.ScanAIResult) *models.ScanAIResult {
	result := &models.ScanAIResult{
		ArchType:     proto.GetArchType(),
		Pronation:    proto.GetPronation(),
		BalanceScore: proto.GetBalanceScore(),
	}

	if proto.GetLlmResult() != nil {
		result.LLMResult = &models.ScanLLMResult{
			Summary: proto.GetLlmResult().GetSummary(),
		}
	}

	if proto.GetRecommendation() != nil {
		rec := &models.ScanRecommendation{}

		for _, p := range proto.GetRecommendation().GetProducts() {
			rec.Products = append(rec.Products, ProductProtoToModel(p))
		}

		for _, e := range proto.GetRecommendation().GetExercises() {
			rec.Exercises = append(rec.Exercises, ExerciseProtoToModel(e))
		}

		for _, t := range proto.GetRecommendation().GetTherapies() {
			rec.Therapies = append(rec.Therapies, TherapyProtoToModel(t))
		}

		result.Recommendation = rec
	}

	return result
}

func ScanAIResultModelToProto(model *models.ScanAIResult) *common.ScanAIResult {
	result := &common.ScanAIResult{
		ArchType:     model.ArchType,
		Pronation:    model.Pronation,
		BalanceScore: model.BalanceScore,
	}

	if model.LLMResult != nil {
		result.LlmResult = &common.ScanLLMResult{
			Summary: model.LLMResult.Summary,
		}
	}

	if model.Recommendation != nil {
		rec := &common.ScanRecommendation{}

		for _, p := range model.Recommendation.Products {
			rec.Products = append(rec.Products, ProductModelToProto(p))
		}

		for _, e := range model.Recommendation.Exercises {
			rec.Exercises = append(rec.Exercises, ExerciseModelToProto(e))
		}

		for _, t := range model.Recommendation.Therapies {
			rec.Therapies = append(rec.Therapies, TherapyModelToProto(t))
		}

		result.Recommendation = rec
	}

	return result
}

// Image conversion
func ImageProtoToModel(proto *common.Image) models.Image {
	return models.Image{
		Type:               proto.GetType().String(),
		URL:                proto.GetUrl(),
		CapturedAt:         timestampPBToTime(proto.GetCapturedAt()),
		SignedURL:          proto.GetSignedUrl(),
		ThumbnailSignedURL: proto.GetThumbnailUrl(),
		Path:               proto.GetGcsPath(),
		ThumbnailPath:      proto.GetThumbnailPath(),
		ExpiresAt:          timestampPBToTime(proto.GetExpiresAt()),
	}
}

func ImageModelToProto(model models.Image) *common.Image {
	// Convert string type to ImageType enum
	imageType, exists := common.ImageType_value[model.Type]
	if !exists {
		// Default to unspecified if the type doesn't exist
		imageType = int32(common.ImageType_IMAGE_TYPE_UNSPECIFIED)
	}
	return &common.Image{
		Type:          common.ImageType(imageType),
		Url:           model.URL,
		CapturedAt:    timestamppb.New(model.CapturedAt),
		SignedUrl:     model.SignedURL,
		ThumbnailUrl:  model.ThumbnailSignedURL,
		GcsPath:       model.Path,
		ThumbnailPath: model.ThumbnailPath,
		ExpiresAt:     timestamppb.New(model.ExpiresAt),
	}
}

// Video conversion
func VideoProtoToModel(proto *common.Video) models.Video {
	return models.Video{
		Type:               proto.GetType().String(),
		URL:                proto.GetUrl(),
		Duration:           proto.GetDuration(),
		CapturedAt:         timestampPBToTime(proto.GetCapturedAt()),
		SignedURL:          proto.GetSignedUrl(),
		ThumbnailSignedURL: proto.GetThumbnailUrl(),
		Path:               proto.GetGcsPath(),
		ThumbnailPath:      proto.GetThumbnailPath(),
		ExpiresAt:          timestampPBToTime(proto.GetExpiresAt()),
	}
}

func VideoModelToProto(model models.Video) *common.Video {
	// Convert string type to VideoType enum
	videoType, exists := common.VideoType_value[model.Type]
	if !exists {
		// Default to unspecified if the type doesn't exist
		videoType = int32(common.VideoType_VIDEO_TYPE_UNSPECIFIED)
	}
	return &common.Video{
		Type:          common.VideoType(videoType),
		Url:           model.URL,
		Duration:      model.Duration,
		CapturedAt:    timestamppb.New(model.CapturedAt),
		SignedUrl:     model.SignedURL,
		ThumbnailUrl:  model.ThumbnailSignedURL,
		GcsPath:       model.Path,
		ThumbnailPath: model.ThumbnailPath,
		ExpiresAt:     timestamppb.New(model.ExpiresAt),
	}
}

// Product conversion
func ProductProtoToModel(proto *common.Product) models.Product {
	product := models.Product{
		ID:          proto.GetId(),
		Name:        proto.GetName(),
		Description: proto.GetDescription(),
		MediaURLs:   proto.GetMediaUrls(),
	}

	if proto.GetCategory() != nil {
		product.Category = &models.ProductCategory{
			ID:          proto.GetCategory().GetId(),
			Name:        proto.GetCategory().GetName(),
			Description: proto.GetCategory().GetDescription(),
			ImageURL:    proto.GetCategory().GetImageUrl(),
		}
	}

	for _, p := range proto.GetPrices() {
		product.Prices = append(product.Prices, models.ProductPrice{
			ProductID: p.GetProductId(),
			Price:     p.GetPrice(),
			Currency:  p.GetCurrency(),
		})
	}

	return product
}

func ProductModelToProto(model models.Product) *common.Product {
	product := &common.Product{
		Id:          model.ID,
		Name:        model.Name,
		Description: model.Description,
		MediaUrls:   model.MediaURLs,
	}

	if model.Category != nil {
		product.Category = &common.ProductCategory{
			Id:          model.Category.ID,
			Name:        model.Category.Name,
			Description: model.Category.Description,
			ImageUrl:    model.Category.ImageURL,
		}
	}

	for _, p := range model.Prices {
		product.Prices = append(product.Prices, &common.ProductPrice{
			ProductId: p.ProductID,
			Price:     p.Price,
			Currency:  p.Currency,
		})
	}

	return product
}

// Exercise conversion
func ExerciseProtoToModel(proto *common.Exercise) models.Exercise {
	return models.Exercise{
		ID:          proto.GetId(),
		Name:        proto.GetName(),
		Description: proto.GetDescription(),
		VideoURL:    proto.GetVideoUrl(),
	}
}

func ExerciseModelToProto(model models.Exercise) *common.Exercise {
	return &common.Exercise{
		Id:          model.ID,
		Name:        model.Name,
		Description: model.Description,
		VideoUrl:    model.VideoURL,
	}
}

// Therapy conversion
func TherapyProtoToModel(proto *common.Therapy) models.Therapy {
	return models.Therapy{
		ID:          proto.GetId(),
		Name:        proto.GetName(),
		Description: proto.GetDescription(),
		VideoURL:    proto.GetVideoUrl(),
	}
}

func TherapyModelToProto(model models.Therapy) *common.Therapy {
	return &common.Therapy{
		Id:          model.ID,
		Name:        model.Name,
		Description: model.Description,
		VideoUrl:    model.VideoURL,
	}
}

// Patient Proto <-> Model conversion
func PatientModelToProto(m models.Patient) *podoai.Patient {
	return &podoai.Patient{
		Id:            m.ID.Hex(),
		Name:          m.Name,
		PhoneNumber:   m.PhoneNumber,
		OwnerEntityId: m.OwnerEntityID,
		Age:           m.Age,
		Gender:        GenderStringToProto(m.Gender),
		FootSize:      m.FootSize,
		TotalScans:    m.TotalScans,
		LastScanDate:  timestamppb.New(m.LastScanDate),
		CreatedAt:     timestamppb.New(m.CreatedAt),
	}
}

// Proto Patient -> Model Patient
func PatientProtoToModel(p *podoai.Patient) models.Patient {
	var id primitive.ObjectID
	if p.Id != "" {
		oid, err := primitive.ObjectIDFromHex(p.Id)
		if err == nil {
			id = oid
		}
	}
	var lastScanDate, createdAt time.Time
	if p.LastScanDate != nil {
		lastScanDate = p.LastScanDate.AsTime()
	}
	if p.CreatedAt != nil {
		createdAt = p.CreatedAt.AsTime()
	}

	return models.Patient{
		ID:            id,
		Name:          p.Name,
		PhoneNumber:   p.PhoneNumber,
		OwnerEntityID: p.OwnerEntityId,
		Age:           p.Age,
		Gender:        GenderProtoToString(p.Gender),
		FootSize:      p.FootSize,
		TotalScans:    p.TotalScans,
		LastScanDate:  lastScanDate,
		CreatedAt:     createdAt,
	}
}

// Gender conversion helpers
func GenderStringToProto(gender string) podoai.Gender {
	switch gender {
	case "MALE":
		return podoai.Gender_MALE
	case "FEMALE":
		return podoai.Gender_FEMALE
	case "OTHER":
		return podoai.Gender_OTHER
	default:
		return podoai.Gender_GENDER_UNSPECIFIED
	}
}

func GenderProtoToString(gender podoai.Gender) string {
	switch gender {
	case podoai.Gender_MALE:
		return "MALE"
	case podoai.Gender_FEMALE:
		return "FEMALE"
	case podoai.Gender_OTHER:
		return "OTHER"
	default:
		return "GENDER_UNSPECIFIED"
	}
}

// RolesToStrings converts proto Roles to string slice using proto's built-in String() method
func RolesToStrings(roles []pb.Role) []string {
	var roleStrs []string
	for _, role := range roles {
		if role != pb.Role_ROLE_UNSPECIFIED {
			// Use proto's built-in String() method and convert to lowercase
			roleStrs = append(roleStrs, strings.ToLower(role.String()))
		}
	}
	return roleStrs
}

// StringsToRoles converts string slice to proto Roles using proto's built-in parsing
func StringsToRoles(roleStrs []string) []pb.Role {
	var roles []pb.Role
	for _, roleStr := range roleStrs {
		// Convert to uppercase since proto enum values are uppercase
		upperRoleStr := strings.ToUpper(roleStr)

		// Use proto's built-in parsing with Role_value map
		if roleValue, exists := pb.Role_value[upperRoleStr]; exists {
			role := pb.Role(roleValue)
			if role != pb.Role_ROLE_UNSPECIFIED {
				roles = append(roles, role)
			}
		}
	}
	return roles
}

// Helper for timestamp conversion
func timestampPBToTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}
