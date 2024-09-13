package gudu

import (
	"database/sql"
	"github.com/jackc/pgx/v5/pgxpool"
)

type initializedFoldersPath struct {
	currentRootPath string
	folderNames     []string
}

// packageConfigs for package default configurations
type packageConfigs struct {
	port             string
	renderer         string
	sessionStoreType string
	cookies          cookieConfig
	databaseConfigs  databaseConfig
	redis            redisConfig
}

// cookieConfig for session configurations
type cookieConfig struct {
	name     string
	lifetime string
	persist  string
	secure   string
	domain   string
}

type databaseConfig struct {
	dsn          string
	databaseType string
}

type DatabaseConn struct {
	DatabaseType string
	SqlConnPool  *sql.DB
	PgxConnPool  *pgxpool.Pool
}

type redisConfig struct {
	host     string
	password string
	prefix   string
}
