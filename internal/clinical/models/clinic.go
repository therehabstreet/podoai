package models

import (
	"time"

	pb "github.com/therehabstreet/podoai/proto/common"
)

type Clinic struct {
	ID      string `bson:"_id"`
	Name    string `bson:"name"`
	Address string `bson:"address"`
}

type ClinicUser struct {
	ID          string    `bson:"_id"`
	Name        string    `bson:"name"`
	PhoneNumber string    `bson:"phone_number"`
	Email       string    `bson:"email,omitempty"`
	Roles       []string  `bson:"roles"`
	ClinicID    string    `bson:"clinic_id"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
}

// DefaultClinicUserRoles returns default roles for clinic users as strings
func DefaultClinicUserRoles() []string {
	return []string{"clinic_staff"} // Use proto enum string value
}

// DefaultClinicUserRolesProto returns default proto roles for clinic users
func DefaultClinicUserRolesProto() []pb.Role {
	return []pb.Role{pb.Role_CLINIC_STAFF}
}
