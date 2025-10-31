package models

import (
	"time"
)

// Patient represents a patient in the system
type Patient struct {
	ID            string    `bson:"_id,omitempty"`
	Name          string    `bson:"name"`
	PhoneNumber   string    `bson:"phone_number"`
	OwnerEntityID string    `bson:"owner_entity_id"`
	Age           int32     `bson:"age"`
	Gender        string    `bson:"gender"`
	FootSize      int32     `bson:"foot_size"`
	TotalScans    int32     `bson:"total_scans"`
	LastScanDate  time.Time `bson:"last_scan_date"`
	CreatedAt     time.Time `bson:"created_at"`
	Weight        float64   `bson:"weight"` // Weight in kg
}
