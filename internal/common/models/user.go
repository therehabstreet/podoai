package models

import (
	"time"

	pb "github.com/therehabstreet/podoai/proto/common"
)

// User represents a user in the system (common for both consumer and clinical)
type User struct {
	ID          string    `bson:"_id,omitempty"`
	Name        string    `bson:"name"`
	PhoneNumber string    `bson:"phone_number"`
	Email       string    `bson:"email,omitempty"`
	Roles       []string  `bson:"roles"` // Store as strings in MongoDB for simplicity
	Age         string    `bson:"age,omitempty"`
	Gender      string    `bson:"gender,omitempty"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
}

// DefaultConsumerRoles returns default roles for consumer users as strings
func DefaultConsumerRoles() []string {
	return []string{"consumer"} // Use proto enum string value
}

// DefaultClinicalRoles returns default roles for clinical users as strings
func DefaultClinicalRoles() []string {
	return []string{"clinic_staff"} // Use proto enum string value
}

// DefaultConsumerRolesProto returns default proto roles for consumer users
func DefaultConsumerRolesProto() []pb.Role {
	return []pb.Role{pb.Role_CONSUMER}
}

// DefaultClinicalRolesProto returns default proto roles for clinical users
func DefaultClinicalRolesProto() []pb.Role {
	return []pb.Role{pb.Role_CLINIC_STAFF}
}
