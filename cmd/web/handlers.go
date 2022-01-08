package main

import "net/http"

func (app *application) VirtualTerminal(w http.ResponseWriter, r *http.Request) {
	app.infoLog.Println("[application][VirtualTerminal] => ()")
	if err := app.renderTemplate(w, r, "terminal", nil); err != nil {
		app.errorLog.Println("[application][VirtualTerminal] => (error encounterd):")
		app.errorLog.Println(err)
	}
}
