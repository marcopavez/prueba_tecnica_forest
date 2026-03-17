package models

import "time"

type User struct {
	ID             int64     `json:"id"`
	Email          string    `json:"email"`
	HashedPassword string    `json:"-"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Bike struct {
	ID          int64     `json:"id"`
	IsAvailable bool      `json:"is_available"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	PricePerMin float64   `json:"price_per_minute"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Rental struct {
	ID             int64      `json:"id"`
	UserID         int64      `json:"user_id"`
	BikeID         int64      `json:"bike_id"`
	Status         string     `json:"status"` // running, ended
	StartTime      time.Time  `json:"start_time"`
	EndTime        *time.Time `json:"end_time,omitempty"`
	StartLatitude  float64    `json:"start_latitude"`
	StartLongitude float64    `json:"start_longitude"`
	EndLatitude    *float64   `json:"end_latitude,omitempty"`
	EndLongitude   *float64   `json:"end_longitude,omitempty"`
	DurationMins   *int       `json:"duration_minutes,omitempty"`
	Cost           *float64   `json:"cost,omitempty"`
}
