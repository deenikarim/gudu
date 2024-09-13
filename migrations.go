package gudu

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
)

// UpMigrate applying all up migrations.
func (g *Gudu) UpMigrate(dsn string) error {
	// Read migrations from /home/mattes/migrations and connect to a local postgres database.
	m, err := migrate.New("file://"+g.RootPath+"/migrations", dsn)
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
	// Read migrations from /home/mattes/migrations and connect to a local postgres database.
	m, err := migrate.New("file://"+g.RootPath+"/migrations", dsn)
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
	// Read migrations from /home/mattes/migrations and connect to a local postgres database.
	m, err := migrate.New("file://"+g.RootPath+"/migrations", dsn)
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
	// Read migrations from /home/mattes/migrations and connect to a local postgres database.
	m, err := migrate.New("file://"+g.RootPath+"/migrations", dsn)
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
