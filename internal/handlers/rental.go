package handlers

import (
	"database/sql"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/bikerental/api/internal/middleware"
)

type RentalHandler struct {
	db *sql.DB
}

func NewRentalHandler(db *sql.DB) *RentalHandler {
	return &RentalHandler{db: db}
}

// randomLocationWithin5km generates a random lat/lng within 5km of the given point
func randomLocationWithin5km(lat, lng float64) (float64, float64) {
	// Earth radius in km
	const earthRadius = 6371.0
	const maxDistKm = 5.0

	// Random distance and bearing
	dist := rand.Float64() * maxDistKm
	bearing := rand.Float64() * 2 * math.Pi

	latRad := lat * math.Pi / 180
	lngRad := lng * math.Pi / 180
	dr := dist / earthRadius

	newLat := math.Asin(math.Sin(latRad)*math.Cos(dr) + math.Cos(latRad)*math.Sin(dr)*math.Cos(bearing))
	newLng := lngRad + math.Atan2(math.Sin(bearing)*math.Sin(dr)*math.Cos(latRad), math.Cos(dr)-math.Sin(latRad)*math.Sin(newLat))

	return newLat * 180 / math.Pi, newLng * 180 / math.Pi
}

func (h *RentalHandler) Start(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var req struct {
		BikeID int64 `json:"bike_id"`
	}
	if err := decode(r, &req); err != nil || req.BikeID == 0 {
		respondError(w, http.StatusBadRequest, "bike_id is required")
		return
	}

	// Check user doesn't have an active rental
	var activeCount int
	h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM rentals WHERE user_id = ? AND status = 'running'`, userID).Scan(&activeCount)
	if activeCount > 0 {
		respondError(w, http.StatusConflict, "you already have an active rental")
		return
	}

	// Check bike exists and is available
	var lat, lng float64
	var isAvailable bool
	err := h.db.QueryRowContext(r.Context(),
		`SELECT is_available, latitude, longitude FROM bikes WHERE id = ?`, req.BikeID).
		Scan(&isAvailable, &lat, &lng)
	if err != nil {
		respondError(w, http.StatusNotFound, "bike not found")
		return
	}
	if !isAvailable {
		respondError(w, http.StatusConflict, "bike is not available")
		return
	}

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not start transaction")
		return
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	result, err := tx.ExecContext(r.Context(),
		`INSERT INTO rentals (user_id, bike_id, status, start_time, start_latitude, start_longitude) VALUES (?, ?, 'running', ?, ?, ?)`,
		userID, req.BikeID, now, lat, lng)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not create rental")
		return
	}

	_, err = tx.ExecContext(r.Context(), `UPDATE bikes SET is_available = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, req.BikeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not update bike availability")
		return
	}

	if err := tx.Commit(); err != nil {
		respondError(w, http.StatusInternalServerError, "could not commit transaction")
		return
	}

	id, _ := result.LastInsertId()
	respond(w, http.StatusCreated, map[string]any{
		"id":              id,
		"bike_id":         req.BikeID,
		"status":          "running",
		"start_time":      now,
		"start_latitude":  lat,
		"start_longitude": lng,
	})
}

func (h *RentalHandler) End(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	// Find active rental
	var rentalID, bikeID int64
	var startLat, startLng float64
	var startTime time.Time
	err := h.db.QueryRowContext(r.Context(),
		`SELECT id, bike_id, start_latitude, start_longitude, start_time FROM rentals WHERE user_id = ? AND status = 'running'`, userID).
		Scan(&rentalID, &bikeID, &startLat, &startLng, &startTime)
	if err != nil {
		respondError(w, http.StatusNotFound, "no active rental found")
		return
	}

	now := time.Now().UTC()
	endLat, endLng := randomLocationWithin5km(startLat, startLng)

	durationMins := int(math.Ceil(now.Sub(startTime).Minutes()))
	if durationMins < 1 {
		durationMins = 1
	}

	// Fetch price per minute
	var pricePerMin float64
	h.db.QueryRowContext(r.Context(), `SELECT price_per_minute FROM bikes WHERE id = ?`, bikeID).Scan(&pricePerMin)
	cost := float64(durationMins) * pricePerMin

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not start transaction")
		return
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(r.Context(),
		`UPDATE rentals SET status = 'ended', end_time = ?, end_latitude = ?, end_longitude = ?, duration_minutes = ?, cost = ? WHERE id = ?`,
		now, endLat, endLng, durationMins, cost, rentalID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not update rental")
		return
	}

	_, err = tx.ExecContext(r.Context(),
		`UPDATE bikes SET is_available = 1, latitude = ?, longitude = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		endLat, endLng, bikeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not update bike")
		return
	}

	if err := tx.Commit(); err != nil {
		respondError(w, http.StatusInternalServerError, "could not commit transaction")
		return
	}

	respond(w, http.StatusOK, map[string]any{
		"id":               rentalID,
		"bike_id":          bikeID,
		"status":           "ended",
		"start_time":       startTime,
		"end_time":         now,
		"start_latitude":   startLat,
		"start_longitude":  startLng,
		"end_latitude":     endLat,
		"end_longitude":    endLng,
		"duration_minutes": durationMins,
		"cost":             cost,
	})
}

func (h *RentalHandler) History(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, bike_id, status, start_time, end_time, start_latitude, start_longitude, end_latitude, end_longitude, duration_minutes, cost
		 FROM rentals WHERE user_id = ? ORDER BY start_time DESC`, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not fetch rental history")
		return
	}
	defer rows.Close()

	type rentalEntry struct {
		ID             int64    `json:"id"`
		BikeID         int64    `json:"bike_id"`
		Status         string   `json:"status"`
		StartTime      string   `json:"start_time"`
		EndTime        *string  `json:"end_time,omitempty"`
		StartLatitude  float64  `json:"start_latitude"`
		StartLongitude float64  `json:"start_longitude"`
		EndLatitude    *float64 `json:"end_latitude,omitempty"`
		EndLongitude   *float64 `json:"end_longitude,omitempty"`
		DurationMins   *int     `json:"duration_minutes,omitempty"`
		Cost           *float64 `json:"cost,omitempty"`
	}

	rentals := []rentalEntry{}
	for rows.Next() {
		var e rentalEntry
		var endTime sql.NullString
		var endLat, endLng sql.NullFloat64
		var durMins sql.NullInt64
		var cost sql.NullFloat64
		if err := rows.Scan(&e.ID, &e.BikeID, &e.Status, &e.StartTime, &endTime,
			&e.StartLatitude, &e.StartLongitude, &endLat, &endLng, &durMins, &cost); err == nil {
			if endTime.Valid {
				e.EndTime = &endTime.String
			}
			if endLat.Valid {
				e.EndLatitude = &endLat.Float64
			}
			if endLng.Valid {
				e.EndLongitude = &endLng.Float64
			}
			if durMins.Valid {
				d := int(durMins.Int64)
				e.DurationMins = &d
			}
			if cost.Valid {
				e.Cost = &cost.Float64
			}
			rentals = append(rentals, e)
		}
	}
	respond(w, http.StatusOK, rentals)
}
