package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(app.SessionLoad)

	mux.Get("/", app.Home)

	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(app.Auth)
		mux.Get("/virtual-terminal", app.VirtualTerminal)
	})

	// mux.Post("/virtual-terminal-payment-succeeded", app.VirtualTerminalPaymentSucceeded)
	// mux.Get("/virtual-terminal-receipt", app.VirtualTerminalReceipt)

	mux.Get("/receipt", app.Receipt)
	mux.Post("/payment-succeeded", app.PaymentSucceeded)

	//mux.Get("/charge-once", app.ChargeOnce)
	mux.Get("/widget/{id}", app.ChargeOnce)

	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	// Subscrption plans
	mux.Get("/plans/bronze", app.BronzePlan)
	mux.Get("/receipt/bronze", app.BronzePlanReceipt)

	// Authentication
	mux.Get("/login", app.LoginPage)
	mux.Post("/login", app.PostLoginPage)
	mux.Get("/logout", app.Logout)

	return mux

}
