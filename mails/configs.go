package mails

import (
	"crypto/tls"
	mailpkg "github.com/xhit/go-simple-mail/v2"
	"os"
	"strconv"
	"time"
)

// MailerConfig holds configuration for the SMTP server
type MailerConfig struct {
	Host           string
	Port           int
	Username       string
	Password       string
	Encryption     mailpkg.Encryption
	From           EmailAddress
	KeepAlive      bool
	ConnectTimeout time.Duration
	SendTimeout    time.Duration
	TLSConfig      *tls.Config
	TemplatesDir   string
}

// LoadConfig loads the SMTP configuration from environment variables
func LoadConfig(currRoot string) *MailerConfig {
	port, err := strconv.Atoi(os.Getenv("MAIL_PORT"))
	if err != nil {
		port = 587
	}

	encryption := mailpkg.EncryptionSTARTTLS
	switch os.Getenv("MAIL_ENCRYPTION") {
	case "ssl":
		encryption = mailpkg.EncryptionSSLTLS
	case "tls":
		encryption = mailpkg.EncryptionSTARTTLS
	case "none", "":
		encryption = mailpkg.EncryptionNone

	}

	config := &MailerConfig{
		Host:       getEnv("MAIL_HOST", "smtp.example.com"),
		Port:       port,
		Username:   getEnv("MAIL_USERNAME", ""),
		Password:   getEnv("MAIL_PASSWORD", ""),
		Encryption: encryption,
		From: EmailAddress{
			Address: getEnv("MAIL_FROM_ADDRESS", "no-reply@example.com"),
			Name:    getEnv("MAIL_FROM_NAME", "Example"),
		},
		// KeepAlive:      getEnv("MAIL_KEEP_ALIVE", "false") == "true",
		ConnectTimeout: 10 * time.Second,
		SendTimeout:    10 * time.Second,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
		TemplatesDir: currRoot + "/mails",
	}

	/*if config.Username == "" || config.Password == "" {
		log.Fatalf("MAIL_USERNAME and MAIL_PASSWORD must be set")
	}*/

	return config
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
