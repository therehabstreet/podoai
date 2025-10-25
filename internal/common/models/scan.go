package models

import "time"

// Scan model for common service - matches proto definition
type Scan struct {
	ID            string        `bson:"_id"`
	PatientID     string        `bson:"patient_id"`
	OwnerEntityID string        `bson:"owner_entity_id"`
	CreatedAt     time.Time     `bson:"created_at"`
	UpdatedAt     time.Time     `bson:"updated_at"`
	Images        []*Image      `bson:"images"`
	Videos        []*Video      `bson:"videos"`
	ScanAIResult  *ScanAIResult `bson:"scan_ai_result,omitempty"`
	ScannedByID   string        `bson:"scanned_by_id"`
	ReviewedByID  string        `bson:"reviewed_by_id"`
	ReviewedAt    *time.Time    `bson:"reviewed_at,omitempty"`
	Status        string        `bson:"status"`
}

type ScanAIResult struct {
	FootScore              float32             `bson:"foot_score"`
	GaitScore              float32             `bson:"gait_score"`
	LeftPronationAngle     float32             `bson:"left_pronation_angle"`
	RightPronationAngle    float32             `bson:"right_pronation_angle"`
	LeftArchHeightIndex    float32             `bson:"left_arch_height_index"`
	RightArchHeightIndex   float32             `bson:"right_arch_height_index"`
	LeftHalluxValgusAngle  float32             `bson:"left_hallux_valgus_angle"`
	RightHalluxValgusAngle float32             `bson:"right_hallux_valgus_angle"`
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

type Image struct {
	Type               string    `bson:"type"`
	CapturedAt         time.Time `bson:"captured_at"`
	SignedURL          string    `bson:"signed_url,omitempty"`           // Temporary signed URL
	ThumbnailSignedURL string    `bson:"thumbnail_signed_url,omitempty"` // Thumbnail signed URL
	Path               string    `bson:"path,omitempty"`                 // GCS path
	ThumbnailPath      string    `bson:"thumbnail_path,omitempty"`       // Thumbnail GCS path
	ExpiresAt          time.Time `bson:"expires_at,omitempty"`           // When signed URLs expire
}

type Video struct {
	Type               string    `bson:"type"`
	Duration           int32     `bson:"duration"`
	CapturedAt         time.Time `bson:"captured_at"`
	SignedURL          string    `bson:"signed_url,omitempty"`           // Temporary signed URL
	ThumbnailSignedURL string    `bson:"thumbnail_signed_url,omitempty"` // Video thumbnail signed URL
	Path               string    `bson:"path,omitempty"`                 // GCS path
	ThumbnailPath      string    `bson:"thumbnail_path,omitempty"`       // Thumbnail GCS path
	ExpiresAt          time.Time `bson:"expires_at,omitempty"`           // When signed URLs expire
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
	Price     float32 `bson:"price"`
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
