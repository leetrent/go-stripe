package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(app.SessionLoad)

	mux.Get("/", app.Home)
	mux.Get("/ws", app.WsEndPoint)

	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(app.Auth)
		mux.Get("/virtual-terminal", app.VirtualTerminal)
		mux.Get("/all-sales", app.AllSales)
		mux.Get("/all-subscriptions", app.AllSubscriptions)
		mux.Get("/sales/{id}", app.ShowSale)
		mux.Get("/subscriptions/{id}", app.ShowSubscription)
		mux.Get("/all-users", app.AllUsers)
		mux.Get("/all-users/{id}", app.OneUser)
	})

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
	mux.Get("/forgot-password", app.ForgotPassword)
	mux.Get("/reset-password", app.ShowResetPassword)

	return mux

}
