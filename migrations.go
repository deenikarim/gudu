package gudu

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// formatMigrationPath adjusts the migration path based on the user's operating system
// and ensures it is an absolute path.
func formatMigrationPath(rootPath string) (string, error) {
	// Ensure rootPath is an absolute path
	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		return "", err
	}

	// Log the absolute path for debugging
	//log.Println("Absolute migration path: ", absPath)

	// Check if the directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return "", err // Return error if the path does not exist
	}

	if runtime.GOOS == "windows" {
		// On Windows, prepend 'file://' and convert backslashes to forward slashes
		return "file://" + strings.ReplaceAll(absPath, "\\", "/"), nil
	} else {
		// On Unix-based systems (macOS/Linux), prepend 'file://'
		return "file://" + filepath.ToSlash(absPath), nil
	}
}

// UpMigrate applying all up migrations.
func (g *Gudu) UpMigrate(dsn string) error {
	// Format the migration path based on the OS and check if it's valid
	migrationPath, err := formatMigrationPath(g.RootPath + "/migrations")
	if err != nil {
		return err
	}
	m, err := migrate.New(migrationPath, dsn)
	if err != nil {
		return err
	}
	defer func(m *migrate.Migrate) {
		_, _ = m.Close()
	}(m)

	// Migrate all the way up ...
	if err := m.Up(); err != nil {
		log.Println("error running up migrations")
		return err
	}
	return nil
}

// DownMigrate applying all down migrations.
func (g *Gudu) DownMigrate(dsn string) error {
	// Format the migration path based on the OS and check if it's valid
	migrationPath, err := formatMigrationPath(g.RootPath + "/migrations")
	if err != nil {
		return err
	}
	m, err := migrate.New(migrationPath, dsn)
	if err != nil {
		return err
	}
	defer func(m *migrate.Migrate) {
		_, _ = m.Close()
	}(m)

	// Migrate all the way down ...
	if err := m.Down(); err != nil {
		log.Println("error running down migrations")
		return err
	}
	return nil
}

// StepsMigrate It will migrate up if n > 0, and down if n < 0.
func (g *Gudu) StepsMigrate(n int, dsn string) error {
	// Format the migration path based on the OS and check if it's valid
	migrationPath, err := formatMigrationPath(g.RootPath + "/migrations")
	if err != nil {
		return err
	}
	m, err := migrate.New(migrationPath, dsn)
	if err != nil {
		return err
	}
	defer func(m *migrate.Migrate) {
		_, _ = m.Close()
	}(m)

	//  It will migrate up if n > 0, and down if n < 0. ...
	if err := m.Steps(n); err != nil {
		log.Println("error running steps migrations")
		return err
	}
	return nil
}

// ForceMigrate sets a migration version. It does not check any currently active version in database.
// It resets the dirty state to false.
func (g *Gudu) ForceMigrate(dsn string) error {
	// Format the migration path based on the OS and check if it's valid
	migrationPath, err := formatMigrationPath(g.RootPath + "/migrations")
	if err != nil {
		return err
	}
	m, err := migrate.New(migrationPath, dsn)
	if err != nil {
		return err
	}
	defer func(m *migrate.Migrate) {
		_, _ = m.Close()
	}(m)

	//  get rid of the last migration run ...
	if err := m.Force(-1); err != nil {
		log.Println("error forcing migrations")
		return err
	}
	return nil
}
