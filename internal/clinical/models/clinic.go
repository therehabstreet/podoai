package models

import "time"

type Clinic struct {
	ID      string `bson:"_id"`
	Name    string `bson:"name"`
	Address string `bson:"address"`
}

type ClinicUser struct {
	ID        string    `bson:"_id"`
	Name      string    `bson:"name"`
	Roles     []string  `bson:"roles"`
	ClinicID  string    `bson:"clinic_id"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}
