package router

import (
	"net/http"

	"github.com/bikerental/api/internal/handlers"
	"github.com/bikerental/api/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type NewRouter struct {
	ChiMux              *chi.Mux
	User                *handlers.UserHandler
	Admin               *handlers.AdminHandler
	Bike                *handlers.BikeHandler
	Rental              *handlers.RentalHandler
	UserAuthMiddleware  *middleware.UserAuth
	AdminAuthMiddleware *middleware.AdminAuth
}

func (router *NewRouter) Route() {
	router.userRoutes()
	router.adminRouter()
	router.healthStatusRouter()
}

func (router *NewRouter) userRoutes() {
	// User routes
	router.ChiMux.Post("/users/register", router.User.Register)
	router.ChiMux.Post("/users/login", router.User.Login)

	router.ChiMux.Group(func(r chi.Router) {
		r.Use(router.UserAuthMiddleware.Authenticate)
		r.Get("/users/profile", router.User.GetProfile)
		r.Patch("/users/profile", router.User.UpdateProfile)

		r.Get("/bikes/available", router.Bike.ListAvailable)

		r.Post("/rentals/start", router.Rental.Start)
		r.Post("/rentals/end", router.Rental.End)
		r.Get("/rentals/history", router.Rental.History)
	})
}

func (router *NewRouter) adminRouter() {
	// Admin routes
	router.ChiMux.Group(func(r chi.Router) {
		r.Use(router.AdminAuthMiddleware.Authenticate)
		r.Post("/admin/bikes", router.Admin.CreateBike)
		r.Patch("/admin/bikes/{bike_id}", router.Admin.UpdateBike)
		r.Get("/admin/bikes", router.Admin.ListBikes)

		r.Get("/admin/users", router.Admin.ListUsers)
		r.Get("/admin/users/{user_id}", router.Admin.GetUser)
		r.Patch("/admin/users/{user_id}", router.Admin.UpdateUser)

		r.Get("/admin/rentals", router.Admin.ListRentals)
		r.Get("/admin/rentals/{rental_id}", router.Admin.GetRental)
		r.Patch("/admin/rentals/{rental_id}", router.Admin.UpdateRental)
	})
}

func (router *NewRouter) healthStatusRouter() {
	router.ChiMux.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
}
