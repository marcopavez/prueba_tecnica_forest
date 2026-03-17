package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type AdminHandler struct {
	db *sql.DB
}

func NewAdminHandler(db *sql.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

// --- Bike Management ---

func (h *AdminHandler) CreateBike(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
		PricePerMin float64 `json:"price_per_minute"`
	}
	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.PricePerMin == 0 {
		req.PricePerMin = 0.1
	}

	result, err := h.db.ExecContext(r.Context(),
		`INSERT INTO bikes (latitude, longitude, price_per_minute) VALUES (?, ?, ?)`,
		req.Latitude, req.Longitude, req.PricePerMin)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not create bike")
		return
	}
	id, _ := result.LastInsertId()
	respond(w, http.StatusCreated, map[string]any{
		"id":              id,
		"is_available":    true,
		"latitude":        req.Latitude,
		"longitude":       req.Longitude,
		"price_per_minute": req.PricePerMin,
	})
}

func (h *AdminHandler) UpdateBike(w http.ResponseWriter, r *http.Request) {
	bikeID, err := strconv.ParseInt(chi.URLParam(r, "bike_id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid bike_id")
		return
	}

	var req struct {
		IsAvailable *bool    `json:"is_available"`
		Latitude    *float64 `json:"latitude"`
		Longitude   *float64 `json:"longitude"`
		PricePerMin *float64 `json:"price_per_minute"`
	}
	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.IsAvailable != nil {
		avail := 0
		if *req.IsAvailable {
			avail = 1
		}
		h.db.ExecContext(r.Context(), `UPDATE bikes SET is_available = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, avail, bikeID)
	}
	if req.Latitude != nil {
		h.db.ExecContext(r.Context(), `UPDATE bikes SET latitude = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, *req.Latitude, bikeID)
	}
	if req.Longitude != nil {
		h.db.ExecContext(r.Context(), `UPDATE bikes SET longitude = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, *req.Longitude, bikeID)
	}
	if req.PricePerMin != nil {
		h.db.ExecContext(r.Context(), `UPDATE bikes SET price_per_minute = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, *req.PricePerMin, bikeID)
	}

	var b struct {
		ID          int64     `json:"id"`
		IsAvailable bool      `json:"is_available"`
		Latitude    float64   `json:"latitude"`
		Longitude   float64   `json:"longitude"`
		PricePerMin float64   `json:"price_per_minute"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}
	err = h.db.QueryRowContext(r.Context(),
		`SELECT id, is_available, latitude, longitude, price_per_minute, created_at, updated_at FROM bikes WHERE id = ?`, bikeID).
		Scan(&b.ID, &b.IsAvailable, &b.Latitude, &b.Longitude, &b.PricePerMin, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		respondError(w, http.StatusNotFound, "bike not found")
		return
	}
	respond(w, http.StatusOK, b)
}

func (h *AdminHandler) ListBikes(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, is_available, latitude, longitude, price_per_minute, created_at, updated_at FROM bikes`)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not fetch bikes")
		return
	}
	defer rows.Close()

	type bikeEntry struct {
		ID          int64     `json:"id"`
		IsAvailable bool      `json:"is_available"`
		Latitude    float64   `json:"latitude"`
		Longitude   float64   `json:"longitude"`
		PricePerMin float64   `json:"price_per_minute"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}
	bikes := []bikeEntry{}
	for rows.Next() {
		var b bikeEntry
		if err := rows.Scan(&b.ID, &b.IsAvailable, &b.Latitude, &b.Longitude, &b.PricePerMin, &b.CreatedAt, &b.UpdatedAt); err == nil {
			bikes = append(bikes, b)
		}
	}
	respond(w, http.StatusOK, bikes)
}

// --- User Management ---

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, email, first_name, last_name, created_at, updated_at FROM users`)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not fetch users")
		return
	}
	defer rows.Close()

	type userEntry struct {
		ID        int64     `json:"id"`
		Email     string    `json:"email"`
		FirstName string    `json:"first_name"`
		LastName  string    `json:"last_name"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	users := []userEntry{}
	for rows.Next() {
		var u userEntry
		if err := rows.Scan(&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.CreatedAt, &u.UpdatedAt); err == nil {
			users = append(users, u)
		}
	}
	respond(w, http.StatusOK, users)
}

func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "user_id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	var u struct {
		ID        int64     `json:"id"`
		Email     string    `json:"email"`
		FirstName string    `json:"first_name"`
		LastName  string    `json:"last_name"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	err = h.db.QueryRowContext(r.Context(),
		`SELECT id, email, first_name, last_name, created_at, updated_at FROM users WHERE id = ?`, userID).
		Scan(&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}
	respond(w, http.StatusOK, u)
}

func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "user_id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	var req struct {
		Email     *string `json:"email"`
		FirstName *string `json:"first_name"`
		LastName  *string `json:"last_name"`
	}
	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email != nil {
		h.db.ExecContext(r.Context(), `UPDATE users SET email = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, *req.Email, userID)
	}
	if req.FirstName != nil {
		h.db.ExecContext(r.Context(), `UPDATE users SET first_name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, *req.FirstName, userID)
	}
	if req.LastName != nil {
		h.db.ExecContext(r.Context(), `UPDATE users SET last_name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, *req.LastName, userID)
	}

	// Return updated user
	r2 := r.Clone(r.Context())
	// reuse GetUser via a fake URL param; easier to just re-query
	var u struct {
		ID        int64     `json:"id"`
		Email     string    `json:"email"`
		FirstName string    `json:"first_name"`
		LastName  string    `json:"last_name"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	_ = r2
	h.db.QueryRowContext(r.Context(),
		`SELECT id, email, first_name, last_name, created_at, updated_at FROM users WHERE id = ?`, userID).
		Scan(&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.CreatedAt, &u.UpdatedAt)
	respond(w, http.StatusOK, u)
}

// --- Rental Management ---

type adminRentalEntry struct {
	ID             int64    `json:"id"`
	UserID         int64    `json:"user_id"`
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

func scanRental(rows *sql.Rows) (adminRentalEntry, error) {
	var e adminRentalEntry
	var endTime sql.NullString
	var endLat, endLng sql.NullFloat64
	var durMins sql.NullInt64
	var cost sql.NullFloat64
	err := rows.Scan(&e.ID, &e.UserID, &e.BikeID, &e.Status, &e.StartTime, &endTime,
		&e.StartLatitude, &e.StartLongitude, &endLat, &endLng, &durMins, &cost)
	if err == nil {
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
	}
	return e, err
}

func (h *AdminHandler) ListRentals(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, user_id, bike_id, status, start_time, end_time, start_latitude, start_longitude, end_latitude, end_longitude, duration_minutes, cost FROM rentals ORDER BY start_time DESC`)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not fetch rentals")
		return
	}
	defer rows.Close()

	rentals := []adminRentalEntry{}
	for rows.Next() {
		if e, err := scanRental(rows); err == nil {
			rentals = append(rentals, e)
		}
	}
	respond(w, http.StatusOK, rentals)
}

func (h *AdminHandler) GetRental(w http.ResponseWriter, r *http.Request) {
	rentalID, err := strconv.ParseInt(chi.URLParam(r, "rental_id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid rental_id")
		return
	}

	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, user_id, bike_id, status, start_time, end_time, start_latitude, start_longitude, end_latitude, end_longitude, duration_minutes, cost FROM rentals WHERE id = ?`, rentalID)
	if err != nil || !rows.Next() {
		respondError(w, http.StatusNotFound, "rental not found")
		return
	}
	defer rows.Close()
	e, err := scanRental(rows)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not scan rental")
		return
	}
	respond(w, http.StatusOK, e)
}

func (h *AdminHandler) UpdateRental(w http.ResponseWriter, r *http.Request) {
	rentalID, err := strconv.ParseInt(chi.URLParam(r, "rental_id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid rental_id")
		return
	}

	var req struct {
		Status *string `json:"status"`
	}
	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Status != nil {
		h.db.ExecContext(r.Context(), `UPDATE rentals SET status = ? WHERE id = ?`, *req.Status, rentalID)
	}

	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, user_id, bike_id, status, start_time, end_time, start_latitude, start_longitude, end_latitude, end_longitude, duration_minutes, cost FROM rentals WHERE id = ?`, rentalID)
	if err != nil || !rows.Next() {
		respondError(w, http.StatusNotFound, "rental not found")
		return
	}
	defer rows.Close()
	e, _ := scanRental(rows)
	respond(w, http.StatusOK, e)
}
