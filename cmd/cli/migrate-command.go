package main

import "errors"

// doMigrate build the migrate command
func doMigrate(arg3, arg4 string) error {
	dsn, err := getDSN()
	if err != nil {
		exitGracefully(errors.New("error getting dsn"))

	}
	switch arg3 {
	case "up":
		err := gud.UpMigrate(dsn)
		if err != nil {
			return err
		}
	case "down":
		if arg4 == "all" {
			err := gud.DownMigrate(dsn)
			if err != nil {
				return err
			}
		} else {
			err := gud.StepsMigrate(-1, dsn)
			if err != nil {
				return err
			}
		}
	case "reset":
		err := gud.DownMigrate(dsn)
		if err != nil {
			return err
		}
		err = gud.UpMigrate(dsn)
		if err != nil {
			return err
		}
	default:
		showHelp()
	}

	return nil
}
