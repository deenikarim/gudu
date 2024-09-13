package mailer

import (
	"bytes"
	"fmt"
	"github.com/vanng822/go-premailer/premailer"
	"html/template"
)

func (m *Mailer) buildHTMLMessage(msg MailMessage) (string, error) {
	templateToRender := fmt.Sprintf("%s/%s.html.gohtml", m.Templates, msg.Template)

	tmpl, err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tplBuffer bytes.Buffer
	if err := tmpl.ExecuteTemplate(&tplBuffer, "body", msg.Data); err != nil {
		return "", err
	}

	formattedMessage := tplBuffer.String()
	formattedMessage, err = m.inlineCSS(formattedMessage)
	if err != nil {
		return "", err
	}

	return formattedMessage, nil
}

func (m *Mailer) buildPlainTextMessage(msg MailMessage) (string, error) {
	templateToRender := fmt.Sprintf("%s/%s.plain.gohtml", m.Templates, msg.Template)

	t, err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "body", msg.Data); err != nil {
		return "", err
	}

	plainMessage := buf.String()

	return plainMessage, nil
}

func (m *Mailer) inlineCSS(s string) (string, error) {
	options := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}

	prem, err := premailer.NewPremailerFromString(s, &options)
	if err != nil {
		return "", err
	}

	html, err := prem.Transform()
	if err != nil {
		return "", err
	}

	return html, nil
}
