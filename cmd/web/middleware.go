package main

import "net/http"

func (app *application) SessionLoad(next http.Handler) http.Handler {
	return app.Session.LoadAndSave(next)
}
