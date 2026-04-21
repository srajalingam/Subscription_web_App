package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(SessionLoad)

	mux.Get("/", app.Home)
	mux.Get("/virtual-terminal", app.VirtualTerminal)
	mux.Post("/virtual-terminal-payment-succeeded", app.VirtualTerminalPaymentSucceeded)
	mux.Get("/virtual-terminal-reciept", app.VirtualTerminalReciept)

	mux.Post("/payment-succeeded", app.PaymentSucceeded)
	mux.Get("/reciept", app.Reciept)
	mux.Get("/widget/{id}", app.ChargeOnce)

	mux.Get("/plans/bronze", app.BronzePlan)

	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}
