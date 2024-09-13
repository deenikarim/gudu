package mailer

type Mailer struct {
	WebDomain   string
	Templates   string
	Port        int
	HostName    string
	UserName    string
	Password    string
	Encryption  string
	FromAddress string
	FromName    string
	Jobs        chan MailMessage
	Results     chan MailResult
	WhichAPI    string
	APIKey      string
	APIUrl      string
}

type MailMessage struct {
	From        string
	FromName    string
	To          string
	Cc          []string
	Subject     string
	Template    string
	Attachments []string
	Data        interface{}
}

type MailResult struct {
	Success bool
	Error   error
}

func (m *Mailer) ListenForMails() {
	for {
		msg := <-m.Jobs
		err := m.Send(msg)
		if err != nil {
			m.Results <- MailResult{
				Success: false,
				Error:   err,
			}
		} else {
			m.Results <- MailResult{
				Success: true,
				Error:   nil,
			}
		}
	}
}

func (m *Mailer) Send(msg MailMessage) error {
	if len(m.WhichAPI) > 0 && len(m.APIKey) > 0 && len(m.APIUrl) > 0 && m.WhichAPI != "smtp" {
		err := m.ChooseAPI(msg)
		if err != nil {
			return err
		}
	}

	return m.SendSMTP(msg)
}