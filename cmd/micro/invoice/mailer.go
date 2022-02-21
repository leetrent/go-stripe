package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"time"

	mail "github.com/xhit/go-simple-mail/v2"
)

//go:embed email-templates
var emailTemplateFS embed.FS

func (app *application) SendMail(from, to, subject, tmpl string, attachments []string, data interface{}) error {

	logSnippet := "[api][mailer] =>"

	////////////////////////////////////////////////////////////////////////////////
	// HTML
	////////////////////////////////////////////////////////////////////////////////
	templateToRender := fmt.Sprintf("email-templates/%s.html.tmpl", tmpl)

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
	//app.infoLog.Printf("%s (htmlMessage): ", logSnippet, htmlMessage)

	////////////////////////////////////////////////////////////////////////////////
	// PLAIN TEXT
	////////////////////////////////////////////////////////////////////////////////
	templateToRender = fmt.Sprintf("email-templates/%s.plain.tmpl", tmpl)

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
	//app.infoLog.Printf("%s (plainMessage): ", logSnippet, plainMessage)

	////////////////////////////////////////////////////////////////////////////////
	// CONFIGURE SMTP SERVER
	////////////////////////////////////////////////////////////////////////////////
	server := mail.NewSMTPClient()
	server.Host = app.config.smtp.host
	server.Port = app.config.smtp.port
	server.Username = app.config.smtp.username
	server.Password = app.config.smtp.password
	server.Encryption = mail.EncryptionTLS
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smtpClient, err := server.Connect()
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	email := mail.NewMSG()
	email.SetFrom(from).AddTo(to).SetSubject(subject)
	email.SetBody(mail.TextHTML, htmlMessage)
	email.AddAlternative(mail.TextPlain, plainMessage)

	if len(attachments) > 0 {
		for _, path := range attachments {
			email.AddAttachment(path)
		}
	}

	err = email.Send(smtpClient)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	app.infoLog.Printf("%s (email successfully sent): ", logSnippet)

	return nil
}
