package main

import (
	"fmt"
	"github.com/fatih/color"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"path/filepath"
	"strings"

	"os"
)

func setUp(arg2, arg3 string) {
	// do not need a .env file when running the new, help and version cmd so don't initialize the .env
	if arg2 != "new" && arg2 != "help" && arg2 != "version" {
		path, err := os.Getwd()
		if err != nil {
			exitGracefully(err)
		}

		err = gud.LoadEnv(path + "/.env")
		if err != nil {
			exitGracefully(err)
		}

		gud.RootPath = path
		gud.DBConnection.DatabaseType = os.Getenv("DATABASE_TYPE")
	}
}

func getDSN() (string, error) {
	dbType := gud.DBConnection.DatabaseType

	// dsn holds the connection string
	var dsn string

	// Retrieve environment variables
	host := os.Getenv("DATABASE_HOST")
	port := os.Getenv("DATABASE_PORT")
	user := os.Getenv("DATABASE_USER")
	password := os.Getenv("DATABASE_PASS")
	dbname := os.Getenv("DATABASE_NAME")
	sslMode := os.Getenv("DATABASE_SSL_MODE")

	if dbType == "pgx" {
		dbType = "postgres"
	}

	// Use default ssl mode if not set
	if sslMode == "" {
		sslMode = "disable"
	}

	// check database type and build a connection string
	switch dbType {
	case "postgresql", "postgres":
		// with password configuration
		if password != "" {
			dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, dbname, sslMode)
		} else {
			// without password configuration
			dsn = fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=%s", user, host, port, dbname, sslMode)

		}
	case "mysql", "mariadb":
		// Build MySQL DSN
		dsn = fmt.Sprintf("%s:%s@/%s?parseTime=True&loc=Local", user, password, dbname)

	default:
		// Unsupported database type
		return "", fmt.Errorf("unsupported database type: %s", dbType)
	}

	return dsn, nil

}

func showHelp() {
	color.Yellow(`Available commands:

	help                    -show the help command
	version                 -show the version command
	migrate                 -run all up migration that have not been previously run
	migrate down            -reverse the most recently run migration
	migrate reset           -run all down migration in reverse order then run run all up migration
	make migration <name>   -create two files, one for up migration and the other for down migration

`)
}

// exitGracefully Helper function to handle errors gracefully
func exitGracefully(err error, msg ...string) {
	message := ""
	if len(msg) > 0 {
		message = msg[0]
	}
	if err != nil {
		color.Red("Error: %v\n", err)
	}
	if len(message) > 0 {
		color.Yellow(message)
	} else {
		color.Green("finished!")
	}
	os.Exit(0)
}

// copyFile Helper function to copy files
func copyFile(sourcePath, destPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer func(source *os.File) {
		_ = source.Close()
	}(source)

	dest, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer func(dest *os.File) {
		_ = dest.Close()
	}(dest)

	_, err = io.Copy(dest, source)
	return err
}

func walkFuncUpdateSourceFiles(path string, fi os.FileInfo, err error) error {
	// check for error before doing anything
	if err != nil {
		return err
	}

	// check if the current file is directory
	if fi.IsDir() {
		return nil
	}

	// only check go files
	matchedGoFiles, err := filepath.Match("*.go", fi.Name())
	if err != nil {
		return err
	}

	// now we have a matching go files, so read its content
	if matchedGoFiles {
		//read file content
		r, err := os.ReadFile(path)
		if err != nil {
			exitGracefully(err)
		}

		newContent := strings.Replace(string(r), "myapp", appURL, -1)

		// write the changed file
		err = os.WriteFile(path, []byte(newContent), 0)
		if err != nil {
			exitGracefully(err)
		}
	}
	return nil
}

func updateSource() {
	// walk through the entire project including folder directories and subfolders
	err := filepath.Walk(".", walkFuncUpdateSourceFiles)
	if err != nil {
		exitGracefully(err)
	}
}
