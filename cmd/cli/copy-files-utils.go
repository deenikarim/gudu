package main

import (
	"embed"
	"errors"
)

//go:embed templates
var templateFS embed.FS

func copyFilesFromTemplate(existTemplatePath, targetFile string) error {
	//check if the destination I am copying the files to, they already exists
	if fileExists(targetFile) {
		return errors.New(targetFile + "already exists")
	}

	//read data from the template
	contentOfFile, err := templateFS.ReadFile(existTemplatePath)
	if err != nil {
		exitGracefully(err)
	}
	// copy the data to the target file
	err = copyDataToFile(contentOfFile, targetFile)
	if err != nil {
		exitGracefully(err)
	}

	return nil
}
