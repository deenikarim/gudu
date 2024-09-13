package mails

// EmailAddress represents an email address with a name
type EmailAddress struct {
	Address string
	Name    string
}

// ContentType represents the type of the email content
type ContentType int

const (
	// TextPlain represents plain text email content
	TextPlain ContentType = iota

	// TextHTML represents HTML email content
	TextHTML
)

// Message represents an email message
type Message struct {
	From        EmailAddress
	ReplyTo     EmailAddress
	To          []EmailAddress
	Cc          []EmailAddress
	Bcc         []EmailAddress
	Subject     string
	Body        string
	HTMLBody    string
	ContentType ContentType
	Attachments []Attachment
	Headers     map[string]string
	Metadata    map[string]string
}

// AddRecipient adds a recipient to the email
func (m *Message) AddRecipient(email, name string) {
	m.To = append(m.To, EmailAddress{email, name})
}

// AddCc adds a CC recipient to the email
func (m *Message) AddCc(email, name string) {
	m.Cc = append(m.Cc, EmailAddress{email, name})
}

// AddBcc adds a BCC recipient to the email
func (m *Message) AddBcc(email, name string) {
	m.Bcc = append(m.Bcc, EmailAddress{email, name})
}

// AddAttachment adds an attachment to the email
func (m *Message) AddAttachment(name string, data []byte, mimeType string, inline bool) {
	m.Attachments = append(m.Attachments, Attachment{Name: name, Data: data, MimeType: mimeType, Inline: inline})
}

// AddAttachmentFromFile adds an attachment from a file path to the email
func (m *Message) AddAttachmentFromFile(filePath string, inline bool) error {
	attachment, err := NewAttachmentFromFile(filePath, inline)
	if err != nil {
		return err
	}
	m.Attachments = append(m.Attachments, *attachment)
	return nil
}

// AddAttachmentFromBase64 adds an attachment from a base64 string to the email
func (m *Message) AddAttachmentFromBase64(name, b64Data, mimeType string, inline bool) error {
	attachment, err := NewAttachmentFromBase64(name, b64Data, mimeType, inline)
	if err != nil {
		return err
	}
	m.Attachments = append(m.Attachments, *attachment)
	return nil
}

// AddAttachmentFromBytes adds an attachment from a byte slice to the email
func (m *Message) AddAttachmentFromBytes(name string, data []byte, mimeType string, inline bool) error {
	attachment, err := NewAttachmentFromBytes(name, data, mimeType, inline)
	if err != nil {
		return err
	}
	m.Attachments = append(m.Attachments, *attachment)
	return nil
}
