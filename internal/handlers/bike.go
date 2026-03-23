package handlers

import (
	"database/sql"
	"net/http"
	"time"
)

type BikeHandler struct {
	db *sql.DB
}

func NewBikeHandler(db *sql.DB) *BikeHandler {
	return &BikeHandler{db: db}
}

type bikeRow struct {
	ID          int64     `json:"id"`
	IsAvailable bool      `json:"is_available"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	PricePerMin float64   `json:"price_per_minute"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (h *BikeHandler) ListAvailable(w http.ResponseWriter, r *http.Request) {

	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, is_available, latitude, longitude, price_per_minute, created_at, updated_at FROM bikes WHERE is_available = ?`)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not fetch bikes")
		return
	}
	defer rows.Close()

	bikes := []bikeRow{}
	for rows.Next() {
		var b bikeRow
		if err := rows.Scan(&b.ID, &b.IsAvailable, &b.Latitude, &b.Longitude, &b.PricePerMin, &b.CreatedAt, &b.UpdatedAt); err == nil {
			bikes = append(bikes, b)
		}
	}
	respond(w, http.StatusOK, bikes)
}
