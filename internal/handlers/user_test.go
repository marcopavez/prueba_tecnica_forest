package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/bikerental/api/internal/auth"
	"github.com/bikerental/api/internal/database"
	"github.com/bikerental/api/internal/handlers"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	f, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })

	dbPath := ""
	database := database.Initialize(dbPath)

	t.Cleanup(func() { database.Close() })
	return database
}

func TestRegisterAndLogin(t *testing.T) {
	database := setupTestDB(t)
	jwtAuth := auth.NewJWTAuth("test-secret")
	h := handlers.NewUserHandler(database, jwtAuth)

	// Register
	body, _ := json.Marshal(map[string]string{
		"email":      "test@example.com",
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
		"email":    "test@example.com",
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
	database := setupTestDB(t)
	jwtAuth := auth.NewJWTAuth("test-secret")
	h := handlers.NewUserHandler(database, jwtAuth)

	body, _ := json.Marshal(map[string]string{
		"email": "dup@example.com", "password": "pass", "first_name": "A", "last_name": "B",
	})

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/users/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		h.Register(w, req)
		if i == 1 && w.Code != http.StatusConflict {
			t.Fatalf("expected 409 on duplicate, got %d", w.Code)
		}
	}
}

func TestLoginWrongPassword(t *testing.T) {
	database := setupTestDB(t)
	jwtAuth := auth.NewJWTAuth("test-secret")
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
