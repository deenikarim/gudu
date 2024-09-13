package mails

import (
	"fmt"
	"sync"
	"time"
)

type Mailer struct {
	Config     *MailerConfig
	Transport  MailTransport
	Scheduler  *Scheduler
	InitOnce   sync.Once //
	EmailQueue chan *Message
}

// Init initializes the Mailer
func (m *Mailer) Init() {
	m.InitOnce.Do(func() {
		InitLogger()
		m.Scheduler.Start()
	})
}

// SendEmail sends a single email
func (m *Mailer) SendEmail(message *Message) error {
	m.Init()
	return m.sendWithRetry(message)
}

// ListenForEmails listens for incoming emails on the emailQueue channel and
// sends them
func (m *Mailer) ListenForEmails() {
	m.Init()
	go func() {
		for msg := range m.EmailQueue {
			if err := m.SendEmail(msg); err != nil {
				ErrorLogger.Printf("Failed to send email: %v", err)
			} else {
				InfoLogger.Printf("Email sent successfully to %v", msg.To)
			}
		}
	}()
}

// QueueEmail queues an email to be sent
func (m *Mailer) QueueEmail(message *Message) {
	m.EmailQueue <- message
}

// SendMultipleEmails sends multiple emails using the same SMTP connection
func (m *Mailer) SendMultipleEmails(messages []*Message) error {
	m.Init()
	return m.Transport.SendMultiple(messages)
}

// sendWithRetry sends an email with retry logic
func (m *Mailer) sendWithRetry(message *Message) error {
	const maxRetries = 3
	for i := 0; i < maxRetries; i++ {
		err := m.Transport.Send(message)
		if err == nil {
			return nil
		}
		ErrorLogger.Printf("Failed to send email, attempt %d/%d: %v", i+1, maxRetries, err)
		time.Sleep(2 * time.Second)

	}
	return fmt.Errorf("failed to send email after %d attempts", maxRetries)
}

// ScheduleEmail schedules an email to be sent at a specific time
func (m *Mailer) ScheduleEmail(message *Message, sendTime time.Time) error {
	m.Init()
	_, err := m.Scheduler.ScheduleEmail(message, sendTime)
	return err
}

// SetBodyFromTemplate sets the email body from a template
func (m *Mailer) SetBodyFromTemplate(message *Message, templateName string, data interface{}) error {
	body, err := m.buildPlainTextMessage(templateName, data)
	if err != nil {
		return err
	}
	message.Body = body
	return nil
}

// SetHTMLBodyFromTemplate sets the HTML email body from a template
func (m *Mailer) SetHTMLBodyFromTemplate(message *Message, templateName string, data interface{}) error {
	htmlBody, err := m.buildHTMLMessage(templateName, data)
	if err != nil {
		return err
	}
	message.HTMLBody = htmlBody
	return nil
}
