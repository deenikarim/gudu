package main

import (
	"errors"
	"github.com/deenikarim/gudu"
	"github.com/fatih/color"
	"os"
)

const version = "1.0.0"

var gud gudu.Gudu

// Main entry point for the command line tool
func main() {
	var message string
	// arg 1 = gudu: load the command line arguments
	arg2, arg3, arg4, err := validateInputs()
	if err != nil {
		existGracefully(err)
	}

	// load setup
	setUp(arg2, arg3)

	switch arg2 {
	case "help":
		showHelp()
	case "new":
		if arg3 == "" {
			existGracefully(errors.New("new require an application name"))
		}
		doNew(arg3)
	case "version":
		color.Yellow("Application version: " + version)
	case "make":
		if arg3 == "" {
			existGracefully(errors.New("make required a subcommand: (migration|handlers)"))
		}
		err = doMake(arg3, arg4)
		if err != nil {
			existGracefully(err)
		}
	case "migrate":
		if arg3 == "" {
			arg3 = "up"
		}
		err = doMigrate(arg3, arg4)
		if err != nil {
			existGracefully(err)
		}
		message = "migrations complete!"
	case "auth":

	default:
		showHelp()
	}
	existGracefully(nil, message)
}

func validateInputs() (string, string, string, error) {
	var arg2, arg3, arg4 string

	if len(os.Args) > 1 {
		arg2 = os.Args[1]

		if len(os.Args) >= 3 {
			arg3 = os.Args[2]

		}
		if len(os.Args) >= 4 {
			arg4 = os.Args[3]

		}
	} else {
		// first argument in the command line
		color.Red("Error: command required")
		showHelp()
		return "", "", "", errors.New("command required")

	}
	return arg2, arg3, arg4, nil
}
