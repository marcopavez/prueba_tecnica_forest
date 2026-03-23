package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bikerental/api/internal/auth"
	"github.com/bikerental/api/internal/handlers"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { db.Close() })
	return db
}

func TestRegisterAndLogin(t *testing.T) {
	database := setupTestDB(t)
	jwtAuth := auth.NewJWTAuth()
	h := handlers.NewUserHandler(database, jwtAuth)

	// Register
	body, _ := json.Marshal(map[string]string{
		"email":      "nuevotest123@gmail.com",
		"password":   "password123",
		"first_name": "Test",
		"last_name":  "User",
	})
	req := httptest.NewRequest(http.MethodPost, "/users/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.Register(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// Login
	body, _ = json.Marshal(map[string]string{
		"email":    "test123@example.com",
		"password": "password123",
	})
	req = httptest.NewRequest(http.MethodPost, "/users/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	h.Login(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["token"] == "" {
		t.Error("expected non-empty token in login response")
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	db := setupTestDB(t)
	jwtAuth := auth.NewJWTAuth()
	h := handlers.NewUserHandler(db, jwtAuth)

	registerUser := func() *httptest.ResponseRecorder {
		body, _ := json.Marshal(map[string]string{
			"email":      "dzxczxcup@example.com",
			"password":   "password123",
			"first_name": "John",
			"last_name":  "Doe",
		})
		req := httptest.NewRequest(http.MethodPost, "/users/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		h.Register(w, req)
		return w
	}

	// primer registro — debe ser exitoso
	firstResponse := registerUser()
	if firstResponse.Code != http.StatusCreated {
		t.Fatalf("first registration: expected 201, got %d: %s", firstResponse.Code, firstResponse.Body.String())
	}

	// segundo registro con el mismo email — debe fallar
	secondResponse := registerUser()
	if secondResponse.Code != http.StatusConflict {
		t.Fatalf("duplicate registration: expected 409, got %d: %s", secondResponse.Code, secondResponse.Body.String())
	}
}

func TestLoginWrongPassword(t *testing.T) {
	database := setupTestDB(t)
	jwtAuth := auth.NewJWTAuth()
	h := handlers.NewUserHandler(database, jwtAuth)

	// Register first
	body, _ := json.Marshal(map[string]string{
		"email": "u@example.com", "password": "correct", "first_name": "U", "last_name": "S",
	})
	req := httptest.NewRequest(http.MethodPost, "/users/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.Register(httptest.NewRecorder(), req)

	// Login with wrong password
	body, _ = json.Marshal(map[string]string{"email": "u@example.com", "password": "wrong"})
	req = httptest.NewRequest(http.MethodPost, "/users/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.Login(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
