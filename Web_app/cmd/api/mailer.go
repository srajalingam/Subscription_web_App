package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strconv"
	"time"

	mail "github.com/xhit/go-simple-mail/v2"
)

//go:embed templates
var emailTemplateFs embed.FS

// sendEmail
func (app *application) sendEmail(from, to, subject, tmpl string, data interface{}) error {

	templateToRender := fmt.Sprintf("templates/%s.html.tmpl", tmpl)
	t, err := template.New("email-html").ParseFS(emailTemplateFs, templateToRender)

	if err != nil {
		app.errorLog.Println("Error parsing email template:", err)
		return err
	}

	var tpl bytes.Buffer

	if err := t.ExecuteTemplate(&tpl, "body", data); err != nil {
		app.errorLog.Println("Error executing email template:", err)
		return err
	}

	formattedMessage := tpl.String()

	templateToRender = fmt.Sprintf("templates/%s.plain.tmpl", tmpl)
	t, err = template.New("email-plain").ParseFS(emailTemplateFs, templateToRender)

	if err != nil {
		app.errorLog.Println("Error parsing email template:", err)
		return err
	}

	if err := t.ExecuteTemplate(&tpl, "body", data); err != nil {
		app.errorLog.Println("Error executing email template:", err)
		return err
	}

	plainMessage := tpl.String()

	app.infoLog.Println(formattedMessage, plainMessage)

	server := mail.NewSMTPClient()
	server.Host = app.config.smtp.host
	port, err := strconv.Atoi(app.config.smtp.port)
	if err != nil {
		app.errorLog.Println("Error parsing SMTP port:", err)
		return err
	}
	server.Port = port
	server.Username = app.config.smtp.username
	server.Password = app.config.smtp.password
	server.Encryption = mail.EncryptionSTARTTLS
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smtpClient, err := server.Connect()
	if err != nil {
		app.errorLog.Println("Error connecting to SMTP server:", err)
		return err
	}
	email := mail.NewMSG()
	email.SetFrom(from).AddTo(to).
		SetSubject(subject).
		SetBody(mail.TextHTML, formattedMessage)
	email.AddAlternative(mail.TextPlain, plainMessage)

	err = email.Send(smtpClient)
	if err != nil {
		app.errorLog.Println("Error sending email:", err)
		return err
	}
	app.infoLog.Println("Email sent successfully to", to)
	return nil
}
