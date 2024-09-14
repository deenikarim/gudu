package main

import (
	"fmt"
	"github.com/fatih/color"
	"time"
)

// doAuth build the make command
func doAuth() error {
	// make migration
	dbType := gud.DBConnection.DatabaseType
	fileName := fmt.Sprintf("%d_create_auth_table", time.Now().UnixMicro())

	targetUpFilePath := gud.RootPath + "/migrations/" + fileName + "." + dbType + ".up.sql"

	targetDownFilePath := gud.RootPath + "/migrations/" + fileName + "." + dbType + ".down.sql"

	err := copyFilesFromTemplate("templates/migrations/auth_table."+dbType+".sql", targetUpFilePath)
	if err != nil {
		exitGracefully(err)
	}

	err = copyDataToFile([]byte("drop table if exists users cascade"), targetDownFilePath)
	if err != nil {
		exitGracefully(err)
	}

	/*run up migration
	err = doMigrate("up", "")
	if err != nil {
		exitGracefully(err)
	}*/

	// copy the data models
	err = copyFilesFromTemplate("templates/data/user.go.txt", gud.RootPath+"/data/user.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFilesFromTemplate("templates/data/token.go.txt", gud.RootPath+"/data/token.go")
	if err != nil {
		exitGracefully(err)
	}

	// copy the  middleware
	err = copyFilesFromTemplate("templates/middleware/auth-web.go.txt", gud.RootPath+"/middleware/auth-web.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFilesFromTemplate("templates/middleware/auth-token.go.txt", gud.RootPath+"/middleware/auth-api.go")
	if err != nil {
		exitGracefully(err)
	}

	//display message feedback to end users
	color.Yellow("   -users, tokens and remember_tokens migration created and executed")
	color.Yellow("   -user and token models created!!")
	color.Yellow("   -auth middleware created!!")
	color.Yellow("")
	color.Red("   -dont forget to add user and token models in data/models.go" +
		"and add appropriate middleware to your routes")

	return nil
}
