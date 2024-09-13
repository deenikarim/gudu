package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var appURL string

func doNew(appName string) {

	appName = strings.ToLower(appName)

	appURL = appName

	// sanitize the application name: convert url to a single word name for application
	if strings.Contains(appName, "/") {
		exploded := strings.SplitAfter(appName, "/")
		appName = exploded[(len(exploded) - 1)]
	}

	log.Println("app name is", appName)

	// git clone the skeleton application
	// Clone the given repository to the given directory
	color.Green("\tcloning repository.....")
	// Clones the repository into the given dir, just as a normal git clone does
	_, err := git.PlainClone("./"+appName, false, &git.CloneOptions{
		URL:      "git@github.com:deenikarim/gudu-skeleton.git",
		Progress: os.Stdout,
		Depth:    1,
	})

	if err != nil {
		existGracefully(err)
	}

	//remove the .git repository
	// when repository is cloned there will be a git repository indicating that all the codes
	// belong to a remote repository and that is wrong
	err = os.RemoveAll(fmt.Sprintf("./%s/.git", appName))
	if err != nil {
		existGracefully(err)
	}

	//create a ready to use .env file
	color.Yellow("\tCreating .env file")
	d, err := templateFS.ReadFile("templates/env.txt")
	if err != nil {
		existGracefully(err)
	}
	env := string(d)
	env = strings.ReplaceAll(env, "${APP_NAME}", appName)
	env = strings.ReplaceAll(env, "${KEY}", gud.GenerateRandomString(32))

	err = copyDataToFile([]byte(env), fmt.Sprintf("./%s/.env", appName))
	if err != nil {
		existGracefully(err)
	}

	//create a makefile based on user's operating system
	if runtime.GOOS == "windows" {
		source, err := os.Open(fmt.Sprintf("./%s/Makefile.windows", appName))
		if err != nil {
			existGracefully(err)
		}
		defer func(source *os.File) {
			_ = source.Close()
		}(source)

		dest, err := os.Create(fmt.Sprintf("./%s/Makefile", appName))
		if err != nil {
			existGracefully(err)
		}
		defer func(dest *os.File) {
			_ = dest.Close()
		}(dest)

		_, err = io.Copy(dest, source)
		if err != nil {
			existGracefully(err)
		}

	} else {
		source, err := os.Open(fmt.Sprintf("./%s/Makefile.mac", appName))
		if err != nil {
			existGracefully(err)
		}
		defer func(source *os.File) {
			_ = source.Close()
		}(source)

		dest, err := os.Create(fmt.Sprintf("./%s/Makefile", appName))
		if err != nil {
			existGracefully(err)
		}
		defer func(dest *os.File) {
			_ = dest.Close()
		}(dest)

		_, err = io.Copy(dest, source)
		if err != nil {
			existGracefully(err)
		}
	}
	_ = os.Remove("./" + appName + "/Makefile.windows")
	_ = os.Remove("./" + appName + "/Makefile.mac")

	//update the go mod file
	// delete the go mod file that came with the cloning and create the appropriate mod file
	color.Yellow("\tCreating the go mod file....")
	_ = os.Remove("./" + appName + "/go.mod")
	d, err = templateFS.ReadFile("templates/go.mod.txt")
	if err != nil {
		existGracefully(err)
	}
	mod := string(d)
	env = strings.ReplaceAll(mod, "${APP_NAME}", appURL)

	err = copyDataToFile([]byte(mod), fmt.Sprintf("./%s/go.mod", appName))
	if err != nil {
		existGracefully(err)
	}

	//update the existing go files with the correct imports/name
	color.Yellow("\tupdate the existing go files with the correct imports names....")
	err = os.Chdir("./" + appName)
	if err != nil {
		_ = fmt.Errorf("error while changing directory %w", err)
		return
	}
	updateSource()

	//run go mod tidy in the project directory
	color.Yellow("\tRunning go mod tidy....")
	cmd := exec.Command("go", "mod", "tidy")
	err = cmd.Start()

	if err != nil {
		existGracefully(err)
	}

	// final message to the user of the package
	color.Green("Done building " + appURL)
	color.Green("Good luck with project")
}
