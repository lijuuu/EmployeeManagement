package database

import (
	"time"

	"github.com/google/uuid"
)

type Employee struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Position  string    `json:"position"`
	Salary    float64   `json:"salary"`
	HiredDate time.Time `json:"hired_date"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Credentials struct {
	Email    string `json:"email" example:"admin@gmail.com"`
	Password string `json:"password" example:"password"`
}

type TokenResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}
