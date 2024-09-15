package main

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/gertd/go-pluralize"
	"os"
	"strings"
	"time"
)

// doAuth build the subcommand of authentication for make command
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

	err = copyDataToFile([]byte("drop table if exists users cascade; drop table if exists tokens cascade; drop table if exists remember_tokens;"), targetDownFilePath)
	if err != nil {
		exitGracefully(err)
	}

	//run up migration by adding migrate command directly
	err = doMigrate("up", "")
	if err != nil {
		exitGracefully(err)
	}

	// copy the data models and its methods and also make sure existing files don't overwrite
	err = copyFilesFromTemplate("templates/data/user.go.txt", gud.RootPath+"/data/user.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFilesFromTemplate("templates/data/token.go.txt", gud.RootPath+"/data/token.go")
	if err != nil {
		exitGracefully(err)
	}

	//copy the  middleware
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

// doMigration build the subcommand of migration for make command that create two files for up and down
// migrations
func doMigration(arg4 string) error {
	dbType := gud.DBConnection.DatabaseType
	// checking for migration name
	if arg4 == "" {
		exitGracefully(errors.New("must give the migration a name"))
	}

	migrationFileName := fmt.Sprintf("%d_%s", time.Now().UnixMicro(), arg4)

	// path the up and down migration folders
	targetUpFilePath := gud.RootPath + "/migrations/" + migrationFileName + "." + dbType + ".up.sql"
	targetDownFilePath := gud.RootPath + "/migrations/" + migrationFileName + "." + dbType + ".down.sql"

	// templates for the migration (existing contents embed to be copied to the target folders
	err := copyFilesFromTemplate("templates/migrations/migration."+dbType+".up.sql", targetUpFilePath)
	if err != nil {
		exitGracefully(err)
	}

	err = copyFilesFromTemplate("templates/migrations/migration."+dbType+".down.sql", targetDownFilePath)
	if err != nil {
		exitGracefully(err)
	}

	return nil
}

// doControllers build the subcommand of handlers for make command
func doControllers(arg4 string) error {
	// checking for controller name
	if arg4 == "" {
		exitGracefully(errors.New("must give the controller a name"))
	}

	controllerFileName := gud.RootPath + "/controllers/" + strings.ToLower(arg4) + ".go"
	if fileExists(controllerFileName) {
		exitGracefully(errors.New(controllerFileName + "file already exists"))
	}

	data, err := templateFS.ReadFile("templates/controllers/controller.go.txt")
	if err != nil {
		exitGracefully(err)
	}
	controller := string(data)
	controller = strings.ReplaceAll(controller, "$CONTROLLERNAME$", toCamelCase(arg4))

	err = os.WriteFile(controllerFileName, []byte(controller), 0644)
	if err != nil {
		exitGracefully(err)
	}

	return nil
}

// doModels build the subcommand of models for make command
func doModels(arg4 string) error {
	// checking for model name
	if arg4 == "" {
		exitGracefully(errors.New("must give the model a name"))
	}

	data, err := templateFS.ReadFile("templates/data/model.go.txt")
	if err != nil {
		exitGracefully(err)
	}
	model := string(data)

	plur := pluralize.NewClient()

	var modelName = arg4
	var tableName = arg4

	if plur.IsPlural(arg4) {
		modelName = plur.Singular(arg4)
		tableName = strings.ToLower(tableName)
	}

	// target file
	targetFile := gud.RootPath + "/data/" + strings.ToLower(modelName) + ".go"

	// final version of data going to the target file
	model = strings.ReplaceAll(model, "$MODELNAME$", toCamelCase(modelName))
	model = strings.ReplaceAll(model, "$TABLENAME$", tableName)

	// copy data to the files
	err = copyDataToFile([]byte(model), targetFile)
	if err != nil {
		exitGracefully(err)
	}

	return nil
}
