package main

import (
	"log"
	"net/http"
	"os"

	"github.com/bikerental/api/internal/auth"
	"github.com/bikerental/api/internal/db"
	"github.com/bikerental/api/internal/handlers"
	"github.com/bikerental/api/internal/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Load config from env
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-for-testing"
	}
	adminCredentials := os.Getenv("ADMIN_CREDENTIALS")
	if adminCredentials == "" {
		// default: admin:password
		adminCredentials = "YWRtaW46cGFzc3dvcmQ="
	}
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./bike_rental.db"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Init DB
	database, err := db.New(dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// Auth helpers
	jwtAuth := auth.NewJWTAuth(jwtSecret)
	basicAuth := auth.NewBasicAuth(adminCredentials)

	// Handlers
	userHandler := handlers.NewUserHandler(database, jwtAuth)
	bikeHandler := handlers.NewBikeHandler(database)
	rentalHandler := handlers.NewRentalHandler(database)
	adminHandler := handlers.NewAdminHandler(database)

	// Middleware
	userMiddleware := middleware.NewUserAuth(jwtAuth)
	adminMiddleware := middleware.NewAdminAuth(basicAuth)

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)

	// Utility
	r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// User routes
	r.Post("/users/register", userHandler.Register)
	r.Post("/users/login", userHandler.Login)

	r.Group(func(r chi.Router) {
		r.Use(userMiddleware.Authenticate)
		r.Get("/users/profile", userHandler.GetProfile)
		r.Patch("/users/profile", userHandler.UpdateProfile)

		r.Get("/bikes/available", bikeHandler.ListAvailable)

		r.Post("/rentals/start", rentalHandler.Start)
		r.Post("/rentals/end", rentalHandler.End)
		r.Get("/rentals/history", rentalHandler.History)
	})

	// Admin routes
	r.Group(func(r chi.Router) {
		r.Use(adminMiddleware.Authenticate)
		r.Post("/admin/bikes", adminHandler.CreateBike)
		r.Patch("/admin/bikes/{bike_id}", adminHandler.UpdateBike)
		r.Get("/admin/bikes", adminHandler.ListBikes)

		r.Get("/admin/users", adminHandler.ListUsers)
		r.Get("/admin/users/{user_id}", adminHandler.GetUser)
		r.Patch("/admin/users/{user_id}", adminHandler.UpdateUser)

		r.Get("/admin/rentals", adminHandler.ListRentals)
		r.Get("/admin/rentals/{rental_id}", adminHandler.GetRental)
		r.Patch("/admin/rentals/{rental_id}", adminHandler.UpdateRental)
	})

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
