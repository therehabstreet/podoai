package models

import "time"

// Scan model for common service - matches proto definition
type Scan struct {
	ID            string        `bson:"_id"`
	PatientID     string        `bson:"patient_id"`
	OwnerEntityID string        `bson:"owner_entity_id"`
	Status        string        `bson:"status"`
	Images        []*ScanImage  `bson:"images"`
	Videos        []*ScanVideo  `bson:"videos"`
	AIResult      *ScanAIResult `bson:"ai_result,omitempty"`
	ScannedByID   string        `bson:"scanned_by_id"`
	ReviewedByID  string        `bson:"reviewed_by_id"`
	ReviewedAt    *time.Time    `bson:"reviewed_at,omitempty"`
	CreatedAt     time.Time     `bson:"created_at"`
	UpdatedAt     time.Time     `bson:"updated_at"`
}

type ScanAIResult struct {
	FootScore              float64             `bson:"foot_score"`
	GaitScore              float64             `bson:"gait_score"`
	LeftPronationAngle     float64             `bson:"left_pronation_angle"`
	RightPronationAngle    float64             `bson:"right_pronation_angle"`
	LeftArchHeightIndex    float64             `bson:"left_arch_height_index"`
	RightArchHeightIndex   float64             `bson:"right_arch_height_index"`
	LeftHalluxValgusAngle  float64             `bson:"left_hallux_valgus_angle"`
	RightHalluxValgusAngle float64             `bson:"right_hallux_valgus_angle"`
	LLMResult              *ScanLLMResult      `bson:"llm_result,omitempty"`
	Recommendation         *ScanRecommendation `bson:"recommendation,omitempty"`
}

type ScanLLMResult struct {
	Summary string `bson:"summary"`
}

type ScanRecommendation struct {
	Products  []*Product  `bson:"products"`
	Exercises []*Exercise `bson:"exercises"`
	Therapies []*Therapy  `bson:"therapies"`
}

type ScanImage struct {
	Type               string    `bson:"type"`
	CapturedAt         time.Time `bson:"captured_at"`
	SignedURL          string    `bson:"signed_url,omitempty"`           // Temporary signed URL
	ThumbnailSignedURL string    `bson:"thumbnail_signed_url,omitempty"` // Thumbnail signed URL
	Path               string    `bson:"path,omitempty"`                 // GCS path
	ThumbnailPath      string    `bson:"thumbnail_path,omitempty"`       // Thumbnail GCS path
	ExpiresAt          time.Time `bson:"expires_at,omitempty"`           // When signed URLs expire
}

type ScanVideo struct {
	Type               string    `bson:"type"`
	Duration           int32     `bson:"duration"`
	CapturedAt         time.Time `bson:"captured_at"`
	SignedURL          string    `bson:"signed_url,omitempty"`           // Temporary signed URL
	ThumbnailSignedURL string    `bson:"thumbnail_signed_url,omitempty"` // Video thumbnail signed URL
	Path               string    `bson:"path,omitempty"`                 // GCS path
	ThumbnailPath      string    `bson:"thumbnail_path,omitempty"`       // Thumbnail GCS path
	ExpiresAt          time.Time `bson:"expires_at,omitempty"`           // When signed URLs expire
}

type GeneralImage struct {
	Path          string    `bson:"path"`
	ThumbnailPath string    `bson:"thumbnail_path,omitempty"`
	CapturedAt    time.Time `bson:"captured_at"`
}

type GeneralVideo struct {
	Path          string    `bson:"path"`
	ThumbnailPath string    `bson:"thumbnail_path,omitempty"`
	Duration      int32     `bson:"duration"`
	CapturedAt    time.Time `bson:"captured_at"`
}

// Product, Exercise, Therapy models to match proto
type Product struct {
	ID          string           `bson:"_id"`
	Name        string           `bson:"name"`
	Description string           `bson:"description"`
	MediaURLs   []string         `bson:"media_urls"`
	Category    *ProductCategory `bson:"category,omitempty"`
	Prices      []*ProductPrice  `bson:"prices"`
}

type ProductCategory struct {
	ID          string `bson:"_id"`
	Name        string `bson:"name"`
	Description string `bson:"description"`
	ImageURL    string `bson:"image_url"`
}

type ProductPrice struct {
	ProductID string  `bson:"product_id"`
	Price     float64 `bson:"price"`
	Currency  string  `bson:"currency"`
}

type Exercise struct {
	ID          string `bson:"_id"`
	Name        string `bson:"name"`
	Description string `bson:"description"`
	VideoURL    string `bson:"video_url"`
}

type Therapy struct {
	ID          string `bson:"_id"`
	Name        string `bson:"name"`
	Description string `bson:"description"`
	VideoURL    string `bson:"video_url"`
}
