package models

import "time"

// Scan model for common service - matches proto definition
type Scan struct {
	ID            string        `bson:"_id"`
	PatientID     string        `bson:"patient_id"`
	OwnerEntityID string        `bson:"owner_entity_id"`
	CreatedAt     time.Time     `bson:"created_at"`
	Images        []Image       `bson:"images"`
	Videos        []Video       `bson:"videos"`
	ScanAIResult  *ScanAIResult `bson:"scan_ai_result,omitempty"`
	ScannedByID   string        `bson:"scanned_by_id"`
	ReviewedByID  string        `bson:"reviewed_by_id"`
	ReviewedAt    *time.Time    `bson:"reviewed_at,omitempty"`
}

type ScanAIResult struct {
	ArchType       string              `bson:"arch_type"`
	Pronation      string              `bson:"pronation"`
	BalanceScore   float32             `bson:"balance_score"`
	LLMResult      *ScanLLMResult      `bson:"llm_result,omitempty"`
	Recommendation *ScanRecommendation `bson:"recommendation,omitempty"`
}

type ScanLLMResult struct {
	Summary string `bson:"summary"`
}

type ScanRecommendation struct {
	Products  []Product  `bson:"products"`
	Exercises []Exercise `bson:"exercises"`
	Therapies []Therapy  `bson:"therapies"`
}

type Image struct {
	Type       string    `bson:"type"`
	URL        string    `bson:"url"`
	CapturedAt time.Time `bson:"captured_at"`
}

type Video struct {
	Type       string    `bson:"type"`
	URL        string    `bson:"url"`
	Duration   int32     `bson:"duration"`
	CapturedAt time.Time `bson:"captured_at"`
}

// Product, Exercise, Therapy models to match proto
type Product struct {
	ID          string           `bson:"_id"`
	Name        string           `bson:"name"`
	Description string           `bson:"description"`
	MediaURLs   []string         `bson:"media_urls"`
	Category    *ProductCategory `bson:"category,omitempty"`
	Prices      []ProductPrice   `bson:"prices"`
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
