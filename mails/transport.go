package mails

import (
	"github.com/toorop/go-dkim"
	mailpkg "github.com/xhit/go-simple-mail/v2"
	"log"
)

// MailTransport defines an interface for sending emails
type MailTransport interface {
	Send(m *Message) error
	SendMultiple(emails []*Message) error
}

// SMTPMailTransport implements MailTransport using go-simple-mail
type SMTPMailTransport struct {
	server *mailpkg.SMTPServer
	client *mailpkg.SMTPClient
}

// NewSMTPMailTransport creates a new SimpleMailTransport with
// the given configuration
func NewSMTPMailTransport(config *MailerConfig) *SMTPMailTransport {
	server := mailpkg.NewSMTPClient()
	server.Host = config.Host
	server.Port = config.Port
	server.Username = config.Username
	server.Password = config.Password
	server.Encryption = config.Encryption
	server.KeepAlive = false // Default to false, managed in SendMultiple
	server.ConnectTimeout = config.ConnectTimeout
	server.SendTimeout = config.SendTimeout
	server.TLSConfig = config.TLSConfig

	client, err := server.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to SMTP server: %v", err)
	}

	return &SMTPMailTransport{
		server: server,
		client: client,
	}
}

// Send sends a single email message
func (s *SMTPMailTransport) Send(m *Message) error {
	email := mailpkg.NewMSG()
	email.SetFrom(m.From.Address).SetSubject(m.Subject)

	if m.ReplyTo.Address != "" {
		email.SetReplyTo(m.ReplyTo.Address)
	}

	if len(m.To) > 0 {
		for _, recipient := range m.To {
			email.AddTo(recipient.Address)
		}
	}

	if len(m.Cc) > 0 {
		for _, cc := range m.Cc {
			email.AddCc(cc.Address)
		}
	}

	if len(m.Bcc) > 0 {
		for _, bcc := range m.Bcc {
			email.AddBcc(bcc.Address)
		}
	}

	// Set the HTML and plain text bodies
	if m.HTMLBody != "" {
		email.SetBody(mailpkg.TextHTML, m.HTMLBody)
	}
	if m.Body != "" {
		email.AddAlternative(mailpkg.TextPlain, m.Body)
	}

	if len(m.Attachments) > 0 {
		for _, attachment := range m.Attachments {
			file := mailpkg.File{
				Name:     attachment.Name,
				MimeType: attachment.MimeType,
				Data:     attachment.Data,
				Inline:   attachment.Inline,
			}
			email.Attach(&file)
		}
	}

	// Add DKIM signature if provided
	if dkimOptions, ok := m.Headers["DkimOptions"]; ok && dkimOptions != "" {
		opts := dkim.NewSigOptions()
		opts.PrivateKey = []byte(dkimOptions)
		opts.Domain = "example.com"
		opts.Selector = "default"
		opts.SignatureExpireIn = 3600
		opts.AddSignatureTimestamp = true
		opts.Headers = []string{"from", "date", "mime-version", "received", "received"}
		opts.Canonicalization = "relaxed/relaxed"
	}

	if email.Error != nil {
		return email.Error
	}

	err := email.Send(s.client)
	if err != nil {
		return err
	}

	return nil
}

// SendMultiple sends multiple email messages using the same SMTP connection
func (s *SMTPMailTransport) SendMultiple(emails []*Message) error {
	// Keep the connection alive for sending multiple emails
	s.client.KeepAlive = true
	defer func(client *mailpkg.SMTPClient) {
		_ = client.Quit()
	}(s.client) // Ensure the connection is closed after sending all emails

	for _, m := range emails {
		err := s.Send(m)
		if err != nil {
			ErrorLogger.Printf("Failed to send email to %v: %v", m.To, err)
		} else {
			InfoLogger.Printf("Email sent successfully to %v", m.To)
		}
	}
	return nil
}
