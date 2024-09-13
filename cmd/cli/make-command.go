package main

import (
	"errors"
	"fmt"
	"time"
)

// doMake build the make command
func doMake(arg3, arg4 string) error {
	switch arg3 {
	case "migration":
		dbType := gud.DBConnection.DatabaseType
		// checking for migration name
		if arg4 == "" {
			existGracefully(errors.New("must give the migration a name"))
		}

		migrationFileName := fmt.Sprintf("%d_%s", time.Now().UnixMicro(), arg4)

		targetUpFilePath := gud.RootPath + "/migrations/" + migrationFileName + "." + dbType + ".up.sql"

		targetDownFilePath := gud.RootPath + "/migrations/" + migrationFileName + "." + dbType + ".down.sql"

		err := copyFilesFromTemplate("templates/migrations/migration."+dbType+".up.sql", targetUpFilePath)
		if err != nil {
			existGracefully(err)
		}

		err = copyFilesFromTemplate("templates/migrations/migration."+dbType+".up.sql", targetDownFilePath)
		if err != nil {
			existGracefully(err)
		}

	case "auth":
		err := doAuth()
		if err != nil {
			existGracefully(err)
		}

	}

	return nil
}
