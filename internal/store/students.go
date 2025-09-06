package store

import (
	"database/sql"
	"time"
)

type Student struct {
	ID                int64     `json:"id"`
	FirstName         string    `json:"first_name"`
	LastName          string    `json:"last_name"`
	Email             string    `json:"email"`
	Password          string    `json:"-"`
	PhoneNumber       *string   `json:"phone_number"`
	Class             string    `json:"class"`
	BirthDate         time.Time `json:"birth_date"`
	Address           string    `json:"address"`
	ParentName        string    `json:"parent_name"`
	ParentPhoneNumber string    `json:"parent_phone_number"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type StudentStore struct {
	db *sql.DB
}
