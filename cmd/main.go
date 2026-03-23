package main

import (
	"log"
	"net/http"
	"os"

	"github.com/bikerental/api/internal/auth"
	"github.com/bikerental/api/internal/database"
	"github.com/bikerental/api/internal/handlers"
	"github.com/bikerental/api/internal/middleware"
	"github.com/bikerental/api/internal/router"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {

	// set and get environment variables
	godotenv.Load()

	// db connection
	database := database.Initialize()
	defer database.Close()

	// configuration of authentication methods
	jwtAuth := auth.NewJWTAuth()
	basicAuth := auth.NewBasicAuth()

	// initialization of handlers
	userHandler := handlers.NewUserHandler(database, jwtAuth)
	bikeHandler := handlers.NewBikeHandler(database)
	rentalHandler := handlers.NewRentalHandler(database)
	adminHandler := handlers.NewAdminHandler(database)

	// configuration of middlewares
	userMiddleware := middleware.NewUserAuth(jwtAuth)
	adminMiddleware := middleware.NewAdminAuth(basicAuth)

	// configuration of endpoints/routes
	chiRouter := chi.NewRouter()
	chiRouter.Use(chimiddleware.Logger)
	chiRouter.Use(chimiddleware.Recoverer)
	chiRouter.Use(chimiddleware.RequestID)

	router := router.NewRouter{
		ChiMux:              chiRouter,
		User:                userHandler,
		Bike:                bikeHandler,
		Rental:              rentalHandler,
		Admin:               adminHandler,
		AdminAuthMiddleware: adminMiddleware,
		UserAuthMiddleware:  userMiddleware,
	}

	router.Route()

	port := os.Getenv("PORT")
	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, chiRouter); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
