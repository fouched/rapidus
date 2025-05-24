package main

func doMigrate(arg2, arg3 string) error {
	dsn := getDSN()

	switch arg2 {
	case "up":
		err := rap.MigrateUp(dsn)
		if err != nil {
			return err
		}
	case "down":
		if arg3 == "all" {
			err := rap.MigrateDownAll(dsn)
			if err != nil {
				return err
			}
		} else {
			err := rap.Steps(-1, dsn)
			if err != nil {
				return err
			}
		}
	case "reset":
		err := rap.MigrateDownAll(dsn)
		if err != nil {
			return err
		}
		err = rap.MigrateUp(dsn)
		if err != nil {
			return err
		}
	default:
		showHelp()
	}

	return nil
}
