package gudu

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"github.com/deenikarim/gudu/cache"
	"github.com/deenikarim/gudu/mailer"
	"github.com/deenikarim/gudu/mails"
	"github.com/deenikarim/gudu/render"
	"github.com/dgraph-io/badger"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func (g *Gudu) InitFolders(p initializedFoldersPath) error {
	currentRoot := p.currentRootPath

	for _, folderName := range p.folderNames {
		// create folders if they don't exist
		err := g.CreateDirIfNotExists(currentRoot + "/" + folderName)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *Gudu) checkDotEnvFile(path string) error {
	err := g.CreateFileIfNotExist(fmt.Sprintf("%s/.env", path))
	if err != nil {
		return err
	}
	return nil
}

// LoadEnv loads the environment variables from the .env file.
func (g *Gudu) LoadEnv(filePath ...string) error {
	// Open the .env file
	file, err := os.Open(filePath[0])
	if err != nil {
		return err
	}
	// Ensure the file is closed when the function exits
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)
	// Read the file line by line
	for scanner.Scan() {
		line := scanner.Text()
		// Remove any leading and trailing whitespace
		line = strings.TrimSpace(line)

		// Skip empty lines and lines starting with '#'
		if line == "" || strings.HasPrefix(line, "#") {
			// Skips the current iteration of the loop and moves to the next line.
			continue
		}

		// Split the line into key and value at the first '=' character
		parts := strings.SplitN(line, "=", 2)
		// Checks if the line was successfully split into exactly two parts
		if len(parts) != 2 {
			// Ignore lines that do not have exactly one '='
			continue
		}

		// Trim whitespace from key and value
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Set the environment variable
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}

	// Check for errors that may have occurred during scanning
	err = scanner.Err()
	if err != nil {
		return err
	}

	return nil
}

// createLoggers creates a customized loggers
func (g *Gudu) createLoggers() (*log.Logger, *log.Logger) {
	var errorLogger *log.Logger
	var infoLogger *log.Logger

	errorLogger = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	infoLogger = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	return infoLogger, errorLogger
}

// ListenAndServe creates a web server listening on the given port and serving
func (g *Gudu) ListenAndServe() {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", os.Getenv("PORT")),
		Handler:      g.Router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 600 * time.Second,
		IdleTimeout:  30 * time.Second,
		ErrorLog:     g.ErrorLog,
	}

	defer func(SqlConnPool *sql.DB) {
		_ = SqlConnPool.Close()
	}(g.DBConnection.SqlConnPool)

	defer g.DBConnection.PgxConnPool.Close()

	g.InfoLog.Printf("Listening on port %s", os.Getenv("PORT"))

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		g.ErrorLog.Fatalf("Could not listen on :%s: %v\n", os.Getenv("PORT"), err)
	}

}

func (g *Gudu) createRenderer() {
	myRender := &render.Render{
		RendererEngine:    g.config.renderer,
		TemplatesRootPath: g.RootPath,
		Port:              g.config.port,
		JetViews:          g.JetViewsSetUp,
		DevelopmentMode:   g.DebugMode,
		Session:           g.Sessions,
	}

	g.Render = myRender
}

// initializeClientRedisCache create a cache redis client by initializing the
// redisCache struct type
func (g *Gudu) initializeClientRedisCache() *cache.RedisCache {
	return &cache.RedisCache{
		Conn:   g.NewRedisCache(),
		Prefix: g.config.redis.prefix,
	}
}

// initializeClientBadgerCache create a cache redis client by initializing the
// redisCache struct type
func (g *Gudu) initializeClientBadgerCache() *cache.BadgerCache {
	db, err := badger.Open(badger.DefaultOptions(g.RootPath + "tmp/badger"))
	if err != nil {
		return nil
	}
	return &cache.BadgerCache{
		Conn:   db,
		Prefix: g.config.redis.prefix,
	}
}

func (g *Gudu) createConnToBadger() *badger.DB {
	db, err := badger.Open(badger.DefaultOptions(g.RootPath + "tmp/badger"))
	if err != nil {
		return nil
	}
	return db
}

func (g *Gudu) createMailer() mailer.Mailer {
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))

	myMailer := mailer.Mailer{
		WebDomain:   os.Getenv("MAIL_DOMAIN"),
		Templates:   g.RootPath + "/mails",
		Port:        port,
		HostName:    os.Getenv("SMTP_HOST"),
		UserName:    os.Getenv("SMTP_USERNAME"),
		Password:    os.Getenv("SMTP_PASSWORD"),
		Encryption:  os.Getenv("SMTP_ENCRYPTION"),
		FromAddress: os.Getenv("FROM_ADDRESS"),
		FromName:    os.Getenv("FROM_NAME"),
		Jobs:        make(chan mailer.MailMessage, 20),
		Results:     make(chan mailer.MailResult, 20),
		WhichAPI:    os.Getenv("API_SERVER"),
		APIKey:      os.Getenv("API_KEY"),
		APIUrl:      os.Getenv("API_URL"),
	}
	return myMailer
}

// NewMailer creates a new Mailer
func (g *Gudu) NewMailer(config *mails.MailerConfig) *mails.Mailer {
	transport := mails.NewSMTPMailTransport(config)
	scheduler := mails.NewScheduler(transport)

	return &mails.Mailer{
		Config:     config,
		Transport:  transport,
		Scheduler:  scheduler,
		EmailQueue: make(chan *mails.Message, 100), // Channel to listen for incoming emails

	}
}
