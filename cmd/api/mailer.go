package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
)

//go:embed templates
var emailTemplateFS embed.FS

func (app *application) SendMail(from, to, subject, tmpl string, data interface{}) error {

	////////////////////////////////////////////////////////////////////////////////
	// HTML
	////////////////////////////////////////////////////////////////////////////////
	templateToRender := fmt.Sprintf("templates/%s.html.tmpl")

	t, err := template.New("email-html").ParseFS(emailTemplateFS, templateToRender)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", data); err != nil {
		app.errorLog.Println(err)
		return err
	}

	htmlMessage := tpl.String()

	////////////////////////////////////////////////////////////////////////////////
	// PLAIN TEXT
	////////////////////////////////////////////////////////////////////////////////
	templateToRender = fmt.Sprintf("templates/%s.plain.tmpl", tmpl)

	t, err = template.New("email-plain").ParseFS(emailTemplateFS, templateToRender)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	if err = t.ExecuteTemplate(&tpl, "body", data); err != nil {
		app.errorLog.Println(err)
		return err
	}

	plainMessage := tpl.String()

	return nil
}
