package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},

		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	mux.Post("/api/payment_intent", app.GetPaymentIntent)
	mux.Get("/api/widget/{id}", app.GetWidgetByID)
	mux.Post("/api/create-customer-and-subscribe-to-plan", app.CreateCustomerAndSubscribeToPlan)

	mux.Post("/api/authenticate", app.CreateAuthToken)

	mux.Post("/api/is-authenticated", app.checkAuthentication)
	mux.Post("/api/forgot-password", app.SendPasswordResetEmail)
	mux.Post("/api/reset-password", app.ResetPassword)
	// Protected routes
	mux.Route("/api/admin", func(mux chi.Router) {
		mux.Use(app.authenticate)

		mux.Post("/virtual-terminal-succeeded", app.VirtualTerminalPaymentSucceeded)
		mux.Post("/all-sales", app.AllSales)
		mux.Get("/get-sale/{id}", app.GetSaleByID)
		mux.Get("/all-subscriptions", app.AllSubscriptions)
		mux.Post("/refund", app.RefundCharge)
		mux.Post("/cancel-subscription", app.CancelSubscription)
	})
	return mux
}
