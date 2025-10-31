package helpers

import (
	"strings"
	"time"

	"github.com/therehabstreet/podoai/internal/common/models"
	"github.com/therehabstreet/podoai/proto/common"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Scan Proto <-> Model conversion
func ScanProtoToModel(proto *common.Scan) *models.Scan {
	scan := &models.Scan{
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
		scan.Images = append(scan.Images, ScanImageProtoToModel(img))
	}

	// Convert videos
	for _, vid := range proto.GetVideos() {
		scan.Videos = append(scan.Videos, ScanVideoProtoToModel(vid))
	}

	// Convert ScanAIResult
	if proto.GetScanAiResult() != nil {
		scan.AIResult = ScanAIResultProtoToModel(proto.GetScanAiResult())
	}

	// Convert ReviewedAt
	if proto.GetReviewedAt() != nil {
		reviewedAt := timestampPBToTime(proto.GetReviewedAt())
		scan.ReviewedAt = &reviewedAt
	}

	return scan
}

func ScanModelToProto(model *models.Scan) *common.Scan {
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
		scan.Images = append(scan.Images, ScanImageModelToProto(img))
	}

	// Convert videos
	for _, vid := range model.Videos {
		scan.Videos = append(scan.Videos, ScanVideoModelToProto(vid))
	}

	// Convert ScanAIResult
	if model.AIResult != nil {
		scan.ScanAiResult = ScanAIResultModelToProto(model.AIResult)
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
		FootScore:              proto.GetFootScore(),
		GaitScore:              proto.GetGaitScore(),
		LeftPronationAngle:     proto.GetLeftPronationAngle(),
		RightPronationAngle:    proto.GetRightPronationAngle(),
		LeftArchHeightIndex:    proto.GetLeftArchHeightIndex(),
		RightArchHeightIndex:   proto.GetRightArchHeightIndex(),
		LeftHalluxValgusAngle:  proto.GetLeftHalluxValgusAngle(),
		RightHalluxValgusAngle: proto.GetRightHalluxValgusAngle(),
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
		FootScore:              model.FootScore,
		GaitScore:              model.GaitScore,
		LeftPronationAngle:     model.LeftPronationAngle,
		RightPronationAngle:    model.RightPronationAngle,
		LeftArchHeightIndex:    model.LeftArchHeightIndex,
		RightArchHeightIndex:   model.RightArchHeightIndex,
		LeftHalluxValgusAngle:  model.LeftHalluxValgusAngle,
		RightHalluxValgusAngle: model.RightHalluxValgusAngle,
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
func ScanImageProtoToModel(proto *common.ScanImage) *models.ScanImage {
	return &models.ScanImage{
		Type:               proto.GetType().String(),
		CapturedAt:         timestampPBToTime(proto.GetCapturedAt()),
		SignedURL:          proto.GetSignedUrl(),
		ThumbnailSignedURL: proto.GetThumbnailSignedUrl(),
		Path:               proto.GetPath(),
		ThumbnailPath:      proto.GetThumbnailPath(),
		ExpiresAt:          timestampPBToTime(proto.GetExpiresAt()),
	}
}

func ScanImageModelToProto(model *models.ScanImage) *common.ScanImage {
	// Convert string type to ImageType enum
	imageType, exists := common.ImageType_value[model.Type]
	if !exists {
		// Default to unspecified if the type doesn't exist
		imageType = int32(common.ImageType_IMAGE_TYPE_UNSPECIFIED)
	}
	return &common.ScanImage{
		Type:               common.ImageType(imageType),
		CapturedAt:         timestamppb.New(model.CapturedAt),
		SignedUrl:          model.SignedURL,
		ThumbnailSignedUrl: model.ThumbnailSignedURL,
		Path:               model.Path,
		ThumbnailPath:      model.ThumbnailPath,
		ExpiresAt:          timestamppb.New(model.ExpiresAt),
	}
}

// Video conversion
func ScanVideoProtoToModel(proto *common.ScanVideo) *models.ScanVideo {
	return &models.ScanVideo{
		Type:               proto.GetType().String(),
		Duration:           proto.GetDuration(),
		CapturedAt:         timestampPBToTime(proto.GetCapturedAt()),
		SignedURL:          proto.GetSignedUrl(),
		ThumbnailSignedURL: proto.GetThumbnailSignedUrl(),
		Path:               proto.GetPath(),
		ThumbnailPath:      proto.GetThumbnailPath(),
		ExpiresAt:          timestampPBToTime(proto.GetExpiresAt()),
	}
}

func ScanVideoModelToProto(model *models.ScanVideo) *common.ScanVideo {
	// Convert string type to VideoType enum
	videoType, exists := common.VideoType_value[model.Type]
	if !exists {
		// Default to unspecified if the type doesn't exist
		videoType = int32(common.VideoType_VIDEO_TYPE_UNSPECIFIED)
	}
	return &common.ScanVideo{
		Type:               common.VideoType(videoType),
		Duration:           model.Duration,
		CapturedAt:         timestamppb.New(model.CapturedAt),
		SignedUrl:          model.SignedURL,
		ThumbnailSignedUrl: model.ThumbnailSignedURL,
		Path:               model.Path,
		ThumbnailPath:      model.ThumbnailPath,
		ExpiresAt:          timestamppb.New(model.ExpiresAt),
	}
}

// GeneralImage conversion
func GeneralImageProtoToModel(proto *common.GeneralImage) *models.GeneralImage {
	return &models.GeneralImage{
		Path:          proto.GetPath(),
		ThumbnailPath: proto.GetThumbnailPath(),
		CapturedAt:    timestampPBToTime(proto.GetCapturedAt()),
	}
}

func GeneralImageModelToProto(model *models.GeneralImage) *common.GeneralImage {
	return &common.GeneralImage{
		Path:          model.Path,
		ThumbnailPath: model.ThumbnailPath,
		CapturedAt:    timestamppb.New(model.CapturedAt),
	}
}

// GeneralVideo conversion
func GeneralVideoProtoToModel(proto *common.GeneralVideo) *models.GeneralVideo {
	return &models.GeneralVideo{
		Path:          proto.GetPath(),
		ThumbnailPath: proto.GetThumbnailPath(),
		Duration:      proto.GetDuration(),
		CapturedAt:    timestampPBToTime(proto.GetCapturedAt()),
	}
}

func GeneralVideoModelToProto(model *models.GeneralVideo) *common.GeneralVideo {
	return &common.GeneralVideo{
		Path:          model.Path,
		ThumbnailPath: model.ThumbnailPath,
		Duration:      model.Duration,
		CapturedAt:    timestamppb.New(model.CapturedAt),
	}
}

// Product conversion
func ProductProtoToModel(proto *common.Product) *models.Product {
	product := &models.Product{
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
		product.Prices = append(product.Prices, &models.ProductPrice{
			ProductID: p.GetProductId(),
			Price:     p.GetPrice(),
			Currency:  p.GetCurrency(),
		})
	}

	return product
}

func ProductModelToProto(model *models.Product) *common.Product {
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
func ExerciseProtoToModel(proto *common.Exercise) *models.Exercise {
	return &models.Exercise{
		ID:          proto.GetId(),
		Name:        proto.GetName(),
		Description: proto.GetDescription(),
		VideoURL:    proto.GetVideoUrl(),
	}
}

func ExerciseModelToProto(model *models.Exercise) *common.Exercise {
	return &common.Exercise{
		Id:          model.ID,
		Name:        model.Name,
		Description: model.Description,
		VideoUrl:    model.VideoURL,
	}
}

// Therapy conversion
func TherapyProtoToModel(proto *common.Therapy) *models.Therapy {
	return &models.Therapy{
		ID:          proto.GetId(),
		Name:        proto.GetName(),
		Description: proto.GetDescription(),
		VideoURL:    proto.GetVideoUrl(),
	}
}

func TherapyModelToProto(model *models.Therapy) *common.Therapy {
	return &common.Therapy{
		Id:          model.ID,
		Name:        model.Name,
		Description: model.Description,
		VideoUrl:    model.VideoURL,
	}
}

// Patient Proto <-> Model conversion
func PatientModelToProto(m *models.Patient) *common.Patient {
	return &common.Patient{
		Id:            m.ID,
		Name:          m.Name,
		PhoneNumber:   m.PhoneNumber,
		OwnerEntityId: m.OwnerEntityID,
		Age:           m.Age,
		Gender:        GenderStringToProto(m.Gender),
		FootSize:      m.FootSize,
		TotalScans:    m.TotalScans,
		LastScanDate:  timestamppb.New(m.LastScanDate),
		CreatedAt:     timestamppb.New(m.CreatedAt),
		Weight:        m.Weight,
	}
}

// Proto Patient -> Model Patient
func PatientProtoToModel(p *common.Patient) *models.Patient {
	var lastScanDate, createdAt time.Time
	if p.LastScanDate != nil {
		lastScanDate = p.LastScanDate.AsTime()
	}
	if p.CreatedAt != nil {
		createdAt = p.CreatedAt.AsTime()
	}

	return &models.Patient{
		ID:            p.Id,
		Name:          p.Name,
		PhoneNumber:   p.PhoneNumber,
		OwnerEntityID: p.OwnerEntityId,
		Age:           p.Age,
		Gender:        GenderProtoToString(p.Gender),
		FootSize:      p.FootSize,
		TotalScans:    p.TotalScans,
		LastScanDate:  lastScanDate,
		CreatedAt:     createdAt,
		Weight:        p.Weight,
	}
}

// Gender conversion helpers
func GenderStringToProto(gender string) common.Gender {
	switch gender {
	case "MALE":
		return common.Gender_MALE
	case "FEMALE":
		return common.Gender_FEMALE
	case "OTHER":
		return common.Gender_OTHER
	default:
		return common.Gender_GENDER_UNSPECIFIED
	}
}

func GenderProtoToString(gender common.Gender) string {
	switch gender {
	case common.Gender_MALE:
		return "MALE"
	case common.Gender_FEMALE:
		return "FEMALE"
	case common.Gender_OTHER:
		return "OTHER"
	default:
		return "GENDER_UNSPECIFIED"
	}
}

// RolesToStrings converts proto Roles to string slice using proto's built-in String() method
func RolesToStrings(roles []common.Role) []string {
	var roleStrs []string
	for _, role := range roles {
		if role != common.Role_ROLE_UNSPECIFIED {
			// Use proto's built-in String() method and convert to lowercase
			roleStrs = append(roleStrs, strings.ToLower(role.String()))
		}
	}
	return roleStrs
}

// StringsToRoles converts string slice to proto Roles using proto's built-in parsing
func StringsToRoles(roleStrs []string) []common.Role {
	var roles []common.Role
	for _, roleStr := range roleStrs {
		// Convert to uppercase since proto enum values are uppercase
		upperRoleStr := strings.ToUpper(roleStr)

		// Use proto's built-in parsing with Role_value map
		if roleValue, exists := common.Role_value[upperRoleStr]; exists {
			role := common.Role(roleValue)
			if role != common.Role_ROLE_UNSPECIFIED {
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
