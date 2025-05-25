package database

import (
	"time"
)

type Employee struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Position  string    `json:"position"`
	Salary    int       `json:"salary"`
	HiredDate string    `json:"hired_date"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}
