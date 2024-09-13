package mailer

import (
	"fmt"
	"github.com/ainsleyclark/go-mail/drivers"
	mailapi "github.com/ainsleyclark/go-mail/mail"
	"os"
	"path/filepath"
)

func (m *Mailer) ChooseAPI(msg MailMessage) error {
	switch m.WhichAPI {
	case "mailgun", "sparkpost", "sendgrid":
		return m.SendAPI(msg, m.WhichAPI)
	default:
		return fmt.Errorf("unknown API %s: only mailgun, sendgrid and sparkpost are supported", m.WhichAPI)
	}
}

func (m *Mailer) SendAPI(msg MailMessage, transport string) error {
	if msg.From == "" {
		msg.From = m.FromAddress
	}

	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

	config := mailapi.Config{
		URL:         m.APIUrl,
		APIKey:      m.APIKey,
		Domain:      m.WebDomain,
		FromAddress: msg.From,
		FromName:    msg.FromName,
	}

	mailerDriver, err := drivers.NewMailgun(config)
	if err != nil {
		return err
	}

	formattedMessage, err := m.buildHTMLMessage(msg)
	if err != nil {
		return err
	}

	plainMessage, err := m.buildPlainTextMessage(msg)
	if err != nil {
		return err
	}

	// Sending Data:
	tx := &mailapi.Transmission{
		Recipients: []string{msg.To},
		CC:         msg.Cc,
		Subject:    msg.Subject,
		HTML:       formattedMessage,
		PlainText:  plainMessage,
	}

	// add attachments
	err = m.addAPIAttachments(msg, tx)

	_, err = mailerDriver.Send(tx)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mailer) addAPIAttachments(msg MailMessage, tx *mailapi.Transmission) error {
	if len(msg.Attachments) > 0 {
		var attachments []mailapi.Attachment

		for _, a := range msg.Attachments {
			var attach mailapi.Attachment
			content, err := os.ReadFile(a)
			if err != nil {
				return err
			}

			fileName := filepath.Base(a)
			attach.Bytes = content
			attach.Filename = fileName
			attachments = append(attachments, attach)
		}

		tx.Attachments = attachments
	}
	return nil
}
