package sessions

import (
	"database/sql"
	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gomodule/redigo/redis"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Session struct {
	CookieName       string
	CookieLifeTime   string
	CookiePersistent string
	CookieDomain     string
	CookieSecure     string
	SessionStore     string
	DBConnPool       *sql.DB
	RedisConnPool    *redis.Pool
}

func (s *Session) InitSession() *scs.SessionManager {
	var secure, persist bool

	// how long should the session lasts
	lifetimeMinutes, err := strconv.Atoi(s.CookieLifeTime)
	if err != nil || lifetimeMinutes <= 0 {
		lifetimeMinutes = 60 // Default to 60 minutes if invalid or missing
	}

	// should cookies persist
	if strings.ToLower(s.CookiePersistent) == "true" {
		persist = true
	}

	// must cookies secure
	if strings.ToLower(s.CookieSecure) == "true" {
		secure = true
	}

	// session setup
	sessionConfig := scs.New()
	sessionConfig.Lifetime = time.Duration(lifetimeMinutes) * time.Minute
	sessionConfig.Cookie.Name = s.CookieName
	sessionConfig.Cookie.Persist = persist
	sessionConfig.Cookie.Secure = secure
	sessionConfig.Cookie.Domain = s.CookieDomain
	sessionConfig.Cookie.SameSite = http.SameSiteStrictMode

	// which session store
	switch strings.ToLower(s.SessionStore) {
	case "redis":
		// Configure session to use Redis store
		sessionConfig.Store = redisstore.New(s.RedisConnPool)
	case "mysql", "mariadb":
		// Configure session to use MySQL/MariaDB store
		sessionConfig.Store = mysqlstore.New(s.DBConnPool)
	case "postgres", "postgresql":
		// Configure session to use PostgresSQL store
		sessionConfig.Store = postgresstore.New(s.DBConnPool)
	default:
		// No external store specified, default to cookie-based session
	}

	return sessionConfig
}
