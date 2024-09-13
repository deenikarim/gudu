package gudu

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/dgraph-io/badger"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"os"
	"time"
)

// OpenDBConnectionPool opens a database connection pool using pgx and the standard sql package.
func (g *Gudu) OpenDBConnectionPool(dbDriverType, connStr string) (*sql.DB, *pgxpool.Pool, error) {
	switch dbDriverType {
	case "postgresql", "postgres":
		dbDriverType = "pgx"
	case "mariadb", "mysql":
		dbDriverType = "mysql"
	}

	// driver configuration and database connection pool creation
	if dbDriverType == "pgx" {
		// Configure pgx pool with connection string
		poolConfig, err := pgxpool.ParseConfig(connStr)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse config: %w", err)
		}
		// Set additional pool configuration if needed
		poolConfig.MaxConnLifetime = time.Minute * 30
		poolConfig.MaxConnIdleTime = time.Minute * 10
		poolConfig.MaxConns = 10
		poolConfig.HealthCheckPeriod = time.Minute * 3

		// Open a connection pool
		connPool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create connection pool: %w", err)
		}

		// Register the pgx ConnConfig with the stdlib package
		stdlib.RegisterConnConfig(poolConfig.ConnConfig)

		// Create a *sql.DB instance using stdlib.OpenDB with pgx.ConnConfig
		// Wrap the pool in a sql.DB instance
		db := stdlib.OpenDB(*poolConfig.ConnConfig)

		// Optionally test the connection
		if err := db.Ping(); err != nil {
			_ = db.Close()
			return nil, nil, fmt.Errorf("failed to ping database with default: %w", err)
		}
		// Optionally test the connection
		if err := connPool.Ping(context.Background()); err != nil {
			connPool.Close()
			return nil, nil, fmt.Errorf("failed to ping database: %w", err)
		}
		return db, connPool, nil

	} else if dbDriverType == "mysql" {
		// Create a *sql.DB instance for MySQL
		db, err := sql.Open("mysql", connStr)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open MySQL database: %w", err)
		}

		// Optionally set additional pool configuration for MySQL
		db.SetConnMaxLifetime(time.Minute * 30)
		db.SetMaxIdleConns(10)
		db.SetMaxOpenConns(10)

		// Optionally test the connection
		if err := db.Ping(); err != nil {
			_ = db.Close()
			return nil, nil, fmt.Errorf("failed to ping MySQL database: %w", err)
		}
		return db, nil, nil
	}

	return nil, nil, fmt.Errorf("unsupported database driver type: %s", dbDriverType)
}

// BuildDSN build a connection string to connect to a database
func (g *Gudu) BuildDSN() (string, error) {
	// dsn holds the connection string
	var dsn string

	// Retrieve environment variables
	host := os.Getenv("DATABASE_HOST")
	port := os.Getenv("DATABASE_PORT")
	user := os.Getenv("DATABASE_USER")
	password := os.Getenv("DATABASE_PASS")
	dbname := os.Getenv("DATABASE_NAME")
	sslMode := os.Getenv("DATABASE_SSL_MODE")
	dbDriverType := os.Getenv("DATABASE_TYPE")

	// Check mandatory environment variables
	if host == "" || port == "" || user == "" || dbname == "" || dbDriverType == "" {
		return "", fmt.Errorf("missing mandatory environment variables")
	}

	// check database type and build a connection string
	switch dbDriverType {
	case "postgresql", "postgres":
		// Set default SSL mode for Postgres if not provided
		if sslMode == "" {
			sslMode = "disable" // Default to no SSL for Postgres
		}

		// Build Postgres DSN
		dsn = fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s timezone=UTC connect_timeout=40",
			host,
			port,
			user,
			dbname,
			sslMode)

		// Append password if provided
		if password != "" {
			dsn = fmt.Sprintf("%s password=%s", dsn, password)
		}
	case "mysql", "mariadb":
		// Set default SSL mode for MySQL if not provided
		if sslMode == "" {
			sslMode = "false" // Default to no SSL for MySQL
		}

		// Build MySQL DSN
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=%s",
			user,
			password,
			host,
			port,
			dbname,
			sslMode) // Add sslMode directly

	default:
		// Unsupported database type
		return "", fmt.Errorf("unsupported database type: %s", dbDriverType)
	}

	return dsn, nil
}

// NewRedisCache initializes a new redisCache instance with connection pooling.
func (g *Gudu) NewRedisCache() *redis.Pool {
	return &redis.Pool{
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", g.config.redis.host,
				redis.DialPassword(g.config.redis.password))
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
		MaxIdle:     50,
		MaxActive:   10000,
		IdleTimeout: 5 * time.Minute,
		TestOnBorrow: func(c redis.Conn, lastUsed time.Time) error {
			_, err := c.Do("PING")
			if err != nil {
				return err
			}
			return nil
		},
	}
}

// NewBadgerCache initializes a new badgerCache instance with a Badger database.
func (g *Gudu) NewBadgerCache(dir string) *badger.DB {
	return nil
}
