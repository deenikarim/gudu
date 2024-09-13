package mails

import (
	"bytes"
	"fmt"
	"github.com/vanng822/go-premailer/premailer"
	htmlTemplate "html/template"
	textTemplate "text/template"
)

// buildHTMLMessage creates the HTML version of the message
func (m *Mailer) buildHTMLMessage(templateName string, data interface{}) (string, error) {
	templateToRender := fmt.Sprintf("%s/%s.html.gohtml", m.Config.TemplatesDir, templateName)

	t, err := htmlTemplate.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", data); err != nil {
		return "", err
	}

	formattedMessage := tpl.String()
	formattedMessage, err = m.inlineCSS(formattedMessage)
	if err != nil {
		return "", err
	}

	return formattedMessage, nil
}

// buildPlainTextMessage creates the plain text version of the message
func (m *Mailer) buildPlainTextMessage(templateName string, data interface{}) (string, error) {
	templateToRender := fmt.Sprintf("%s/%s.plain.gohtml", m.Config.TemplatesDir, templateName)

	t, err := textTemplate.New("email-plain").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", data); err != nil {
		return "", err
	}

	return tpl.String(), nil
}

// inlineCSS takes HTML input as a string and inlines CSS where possible
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
