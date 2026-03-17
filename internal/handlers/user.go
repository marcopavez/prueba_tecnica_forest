package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/bikerental/api/internal/auth"
	"github.com/bikerental/api/internal/middleware"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	db  *sql.DB
	jwt *auth.JWTAuth
}

func NewUserHandler(db *sql.DB, j *auth.JWTAuth) *UserHandler {
	return &UserHandler{db: db, jwt: j}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	if err := decode(r, &req); err != nil || req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not hash password")
		return
	}

	result, err := h.db.ExecContext(r.Context(),
		`INSERT INTO users (email, hashed_password, first_name, last_name) VALUES (?, ?, ?, ?)`,
		req.Email, string(hashed), req.FirstName, req.LastName)
	if err != nil {
		respondError(w, http.StatusConflict, "email already registered")
		return
	}

	id, _ := result.LastInsertId()
	respond(w, http.StatusCreated, map[string]any{
		"id":         id,
		"email":      req.Email,
		"first_name": req.FirstName,
		"last_name":  req.LastName,
	})
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var id int64
	var hashed, firstName, lastName string
	err := h.db.QueryRowContext(r.Context(),
		`SELECT id, hashed_password, first_name, last_name FROM users WHERE email = ?`, req.Email).
		Scan(&id, &hashed, &firstName, &lastName)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(req.Password)); err != nil {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := h.jwt.Generate(id, req.Email, firstName, lastName)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not generate token")
		return
	}

	respond(w, http.StatusOK, map[string]string{"token": token})
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	var user struct {
		ID        int64     `json:"id"`
		Email     string    `json:"email"`
		FirstName string    `json:"first_name"`
		LastName  string    `json:"last_name"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	err := h.db.QueryRowContext(r.Context(),
		`SELECT id, email, first_name, last_name, created_at, updated_at FROM users WHERE id = ?`, userID).
		Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}
	respond(w, http.StatusOK, user)
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	var req struct {
		FirstName *string `json:"first_name"`
		LastName  *string `json:"last_name"`
		Password  *string `json:"password"`
	}
	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.FirstName != nil {
		h.db.ExecContext(r.Context(), `UPDATE users SET first_name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, *req.FirstName, userID)
	}
	if req.LastName != nil {
		h.db.ExecContext(r.Context(), `UPDATE users SET last_name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, *req.LastName, userID)
	}
	if req.Password != nil {
		hashed, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err == nil {
			h.db.ExecContext(r.Context(), `UPDATE users SET hashed_password = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, string(hashed), userID)
		}
	}

	h.GetProfile(w, r)
}
