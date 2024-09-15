package main

// doMake build the make command
func doMake(arg3, arg4 string) error {
	switch arg3 {
	case "migration":
		err := doMigration(arg4)
		if err != nil {
			exitGracefully(err)
		}

	case "auth":
		err := doAuth()
		if err != nil {
			exitGracefully(err)
		}

	case "controllers":
		err := doControllers(arg4)
		if err != nil {
			exitGracefully(err)
		}
	case "model":
		err := doModels(arg4)
		if err != nil {
			exitGracefully(err)
		}
	}

	return nil
}
