package models

import "time"

// User model for consumer service
type User struct {
	ID        string    `bson:"_id"`
	Name      string    `bson:"name"`
	Phone     string    `bson:"phone"`
	Email     string    `bson:"email"`
	Age       string    `bson:"age"`
	Gender    string    `bson:"gender"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}
