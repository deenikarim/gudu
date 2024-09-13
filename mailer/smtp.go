package mailer

import (
	mail "github.com/xhit/go-simple-mail/v2"
	"time"
)

func (m *Mailer) SendSMTP(msg MailMessage) error {
	formattedMessage, err := m.buildHTMLMessage(msg)
	if err != nil {
		return err
	}

	plainMessage, err := m.buildPlainTextMessage(msg)
	if err != nil {
		return err
	}

	// SMTP Server
	server := mail.NewSMTPClient()
	server.Host = m.HostName
	server.Port = m.Port
	server.Username = m.UserName
	server.Encryption = m.getEncryption(m.Encryption)
	server.Password = m.Password
	// Variable to keep alive connection
	server.KeepAlive = false
	// Timeout for connect to SMTP Server
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	// SMTP client
	smtpClient, err := server.Connect()
	if err != nil {
		return err
	}

	// New email simple html with inline and CC
	email := mail.NewMSG()
	email.SetFrom(msg.From).AddTo(msg.To)
	// copy another person
	if len(msg.Cc) > 0 {
		for _, c := range msg.Cc {
			email.AddCc(c)
		}
	}
	email.SetSubject(msg.Subject)

	email.SetBody(mail.TextHTML, formattedMessage)
	email.AddAlternative(mail.TextPlain, plainMessage)
	// set attachments for mails
	if len(msg.Attachments) > 0 {
		for _, a := range msg.Attachments {
			email.AddAttachment(a)
		}
	}

	// always check error before send
	if email.Error != nil {
		return email.Error
	}

	// Call Send and pass the client
	err = email.Send(smtpClient)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mailer) getEncryption(enc string) mail.Encryption {
	switch enc {
	case "tls":
		return mail.EncryptionSTARTTLS
	case "ssl":
		return mail.EncryptionSSL
	case "none":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSTARTTLS
	}
}
