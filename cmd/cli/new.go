package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var appURL string

func doNew(appName string) {
	//Sanitize the Application Name: Ensures that the app name is in lowercase
	//and extracts the name if it's in a URL format.
	appName = strings.ToLower(appName)

	appURL = appName

	if strings.Contains(appName, "/") {
		// Extract the last element after "/"
		exploded := strings.SplitAfter(appName, "/")
		appName = exploded[(len(exploded) - 1)]
	}

	//log.Println("app name is", appName)

	// Clone the skeleton repository
	color.Green("\tcloning repository.....")
	// Clones the repository into the given dir, just as a normal git clone does
	_, err := git.PlainClone("./"+appName, false, &git.CloneOptions{
		URL:      "git@github.com:deenikarim/gudu-skeleton.git",
		Progress: os.Stdout,
		Depth:    1,
	})

	if err != nil {
		exitGracefully(err)
	}

	// remove the .git repository
	// when repository is cloned there will be a git repository indicating that all the codes
	// belong to a remote repository and that is wrong
	// remove .git directory
	color.Yellow("\tRemoving .git directory...")
	err = os.RemoveAll(fmt.Sprintf("./%s/.git", appName))
	if err != nil {
		exitGracefully(err)
	}

	//create a ready to use .env file
	color.Yellow("\tCreating .env file")
	d, err := templateFS.ReadFile("templates/env.txt")
	if err != nil {
		exitGracefully(err)
	}
	env := string(d)
	env = strings.ReplaceAll(env, "${APP_NAME}", appName)
	env = strings.ReplaceAll(env, "${KEY}", gud.GenerateRandomString(32))

	err = copyDataToFile([]byte(env), fmt.Sprintf("./%s/.env", appName))
	if err != nil {
		exitGracefully(err)
	}

	// OS-specific Makefile handling
	var makefileSource string
	if runtime.GOOS == "windows" {
		makefileSource = fmt.Sprintf("./%s/Makefile.windows", appName)
	} else {
		makefileSource = fmt.Sprintf("./%s/Makefile.mac", appName)
	}
	err = copyFile(makefileSource, fmt.Sprintf("./%s/Makefile", appName))
	if err != nil {
		exitGracefully(err)
	}

	// Clean up OS-specific files
	color.Yellow("\tCleaning up OS-specific Makefiles...")
	//_ = os.Remove("./" + appName + "/Makefile.windows")
	//_ = os.Remove("./" + appName + "/Makefile.mac")
	go func() {
		_ = os.Remove(fmt.Sprintf("./%s/Makefile.windows", appName))
	}()
	go func() {
		_ = os.Remove(fmt.Sprintf("./%s/Makefile.mac", appName))
	}()

	//update the go mod file
	// delete the go mod file that came with the cloning and create the appropriate mod file
	color.Yellow("\tCreating the go mod file....")
	err = os.Remove("./" + appName + "/go.mod")
	if err != nil {
		exitGracefully(err)
	}
	d, err = templateFS.ReadFile("templates/go.mod.txt")
	if err != nil {
		exitGracefully(err)
	}
	mod := string(d)
	mod = strings.ReplaceAll(mod, "${APP_NAME}", appURL)

	err = copyDataToFile([]byte(mod), fmt.Sprintf("./%s/go.mod", appName))
	if err != nil {
		exitGracefully(err)
	}

	//update the existing go files with the correct imports/name
	color.Yellow("\tupdate the existing go files with the correct imports names....")
	_ = os.Chdir("./" + appName)

	updateSource()

	//run go mod tidy in the project directory
	color.Yellow("\tRunning go mod tidy....")
	cmd := exec.Command("go", "mod", "tidy")
	err = cmd.Start()

	if err != nil {
		exitGracefully(err)
	}

	// final message to the user of the package
	color.Green("Done building " + appURL)
	color.Green("Good luck with project")
}
