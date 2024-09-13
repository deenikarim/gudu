package gudu

import (
	"fmt"
	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
	"github.com/deenikarim/gudu/cache"
	"github.com/deenikarim/gudu/mailer"
	"github.com/deenikarim/gudu/mails"
	"github.com/deenikarim/gudu/render"
	"github.com/deenikarim/gudu/sessions"
	"github.com/go-chi/chi/v5"
	"log"
	"os"
	"strconv"
)

const version = "1.0.0"

var myRedisCache *cache.RedisCache
var myBadgerCache *cache.BadgerCache

type Gudu struct {
	AppName       string
	DebugMode     bool
	Version       string
	InfoLog       *log.Logger
	ErrorLog      *log.Logger
	RootPath      string
	Response      *Response
	config        packageConfigs
	DBConnection  DatabaseConn // database connection
	Router        *chi.Mux
	Render        *render.Render      // render engine
	Sessions      *scs.SessionManager // session manager
	JetViewsSetUp *jet.Set            // jet template engine
	EncryptionKey string
	Cache         cache.Cache
	Mailer        mailer.Mailer
	MailerMail    *mails.Mailer
}

// New is the main project setup
func (g *Gudu) New(currentRootPath string) error {
	// populate with values
	populateInitializedFoldersPath := initializedFoldersPath{
		currentRootPath: currentRootPath,
		folderNames: []string{
			"controllers", "migrations", "views", "data", "public", "tmp", "log",
			"middleware", "mails",
		},
	}
	// initialize empty folders during project setup if they don't exist
	err := g.InitFolders(populateInitializedFoldersPath)
	if err != nil {
		return err
	}

	twoFolder := initializedFoldersPath{
		currentRootPath: currentRootPath + "/views/",
		folderNames:     []string{"layouts", "pages"},
	}
	err = g.InitFolders(twoFolder)
	if err != nil {
		return err
	}

	// checking if a .env file exists and if not, create it
	err = g.checkDotEnvFile(currentRootPath)
	if err != nil {
		return err
	}

	// if the .env file exists then load and read its content
	err = g.LoadEnv(currentRootPath + "/.env")
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	// called the createLoggers method to create the customized logs
	infoLogger, errorLogger := g.createLoggers()
	g.RootPath = currentRootPath

	// load the mail config and initialize the mailer type
	config := mails.LoadConfig(g.RootPath)
	g.MailerMail = g.NewMailer(config)

	// initialize the response type in the gudu struct
	g.Response = g.NewResponse()

	// todo  connect to database
	// Build DSN based on environment variables
	dsn, err := g.BuildDSN()
	if err != nil {
		errorLogger.Println("can not build DSN: ", err)
		return err
	}

	dbDriverType := os.Getenv("DATABASE_TYPE")

	if dbDriverType != "" {
		sqlDb, pgxPool, err := g.OpenDBConnectionPool(dbDriverType, dsn)
		if err != nil {
			infoLogger.Println("can not connect to database:", err)
			os.Exit(1)
		}
		// populate database in the gudu structure
		g.DBConnection = DatabaseConn{
			DatabaseType: dbDriverType,
			SqlConnPool:  sqlDb,
			PgxConnPool:  pgxPool,
		}
	}

	// todo connect to redis server
	if os.Getenv("CACHE") == "redis" || os.Getenv("SESSION_TYPE") == "redis" {
		myRedisCache = g.initializeClientRedisCache()
		g.Cache = myRedisCache
	}

	// todo connect to badger database
	if os.Getenv("CACHE") == "badger" {
		myBadgerCache = g.initializeClientBadgerCache()
		g.Cache = myBadgerCache
		// set periodic garbage collection once a day
		_, err = g.MailerMail.Scheduler.C.AddFunc("@daily", func() {
			_ = myBadgerCache.Conn.RunValueLogGC(0.7)
		})
		if err != nil {
			return err
		}
	}

	g.InfoLog = infoLogger
	g.ErrorLog = errorLogger

	// populate fields in the Gudu struct type
	g.DebugMode, _ = strconv.ParseBool(os.Getenv("DEBUG"))
	g.Version = version
	g.Router = g.defaultRouter().(*chi.Mux)
	g.Mailer = g.createMailer()

	// configuration settings for the package
	g.config = packageConfigs{
		port:     os.Getenv("PORT"),
		renderer: os.Getenv("RENDERER"),
		cookies: cookieConfig{
			name:     os.Getenv("COOKIE_NAME"),
			lifetime: os.Getenv("COOKIE_LIFETIME"),
			persist:  os.Getenv("COOKIE_PERSIST"),
			secure:   os.Getenv("COOKIE_SECURE"),
			domain:   os.Getenv("COOKIE_DOMAIN"),
		},
		sessionStoreType: os.Getenv("SESSION_TYPE"),
		databaseConfigs: databaseConfig{
			dsn:          dsn,
			databaseType: dbDriverType,
		},
		redis: redisConfig{
			host:     os.Getenv("REDIS_HOST"),
			password: os.Getenv("REDIS_PASSWORD"),
			prefix:   os.Getenv("REDIS_PREFIX"),
		},
	}

	// session management initialisation
	populateSessionManager := sessions.Session{
		CookieName:       g.config.cookies.name,
		CookieLifeTime:   g.config.cookies.lifetime,
		CookiePersistent: g.config.cookies.persist,
		CookieDomain:     g.config.cookies.domain,
		CookieSecure:     g.config.cookies.secure,
		SessionStore:     g.config.sessionStoreType,
	}
	// populate the session store type
	switch g.config.sessionStoreType {
	case "redis":
		populateSessionManager.RedisConnPool = myRedisCache.Conn
	case "mariadb", "mysql", "postgres", "postgresql":
		populateSessionManager.DBConnPool = g.DBConnection.SqlConnPool
	}

	// initialized and store the session in Gudu type
	g.Sessions = populateSessionManager.InitSession()
	g.EncryptionKey = os.Getenv("KEY")

	//****************** jet render setup

	if g.DebugMode {
		var jetSet = jet.NewSet(
			jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", currentRootPath)),
			jet.InDevelopmentMode(),
		)
		g.JetViewsSetUp = jetSet
	} else {
		var jetSet = jet.NewSet(jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", currentRootPath)))
		g.JetViewsSetUp = jetSet
	}

	// populate the render struct type with field values
	g.createRenderer()

	// start the mail channel to listen for mails
	go g.Mailer.ListenForMails()

	// Listen for incoming emails on the emailQueue channel
	go g.MailerMail.ListenForEmails()

	return nil
}
