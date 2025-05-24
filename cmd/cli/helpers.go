package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"os"
	"path/filepath"
	"strings"
)

func setup(arg1, arg2 string) {
	if arg1 != "new" && arg1 != "version" && arg1 != "help" {
		err := godotenv.Load()
		if err != nil {
			exitGracefully(err)
		}

		path, err := os.Getwd()
		if err != nil {
			exitGracefully(err)
		}

		rap.RootPath = path
		rap.DB.Type = os.Getenv("DATABASE_TYPE")
	}
}

func getDSN() string {
	dbType := rap.DB.Type

	//we use pgx, but golang migrate uses different driver
	//so convert dsn to work with golang migrate
	if dbType == "pgx" {
		dbType = "postgres"
	}

	if dbType == "postgres" {
		var dsn string
		if os.Getenv("DATABASE_PASS") != "" {
			dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
				os.Getenv("DATABASE_USER"),
				os.Getenv("DATABASE_PASS"),
				os.Getenv("DATABASE_HOST"),
				os.Getenv("DATABASE_PORT"),
				os.Getenv("DATABASE_NAME"),
				os.Getenv("DATABASE_SSL_MODE"))
		} else {
			dsn = fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=%s",
				os.Getenv("DATABASE_USER"),
				os.Getenv("DATABASE_HOST"),
				os.Getenv("DATABASE_PORT"),
				os.Getenv("DATABASE_NAME"),
				os.Getenv("DATABASE_SSL_MODE"))
		}
		return dsn
	}
	return "mysql://" + rap.BuildDSN()
}

func showHelp() {
	color.Yellow(`Available commands:

    help                     - show help
    version                  - print version
    make auth                - creates authentication tables, models and middleware
    make handler <name>      - creates a stub handler in the handlers directory
    make key                 - creates a random 32 character encryption key
    make mail <name>         - creates starter templates for text and html emails in the mail directory
    make model <name>        - creates a new model in the data directory
    make session             - creates a new table as a session store
    
    make migration <name>    - creates new up and down migrations
    migrate                  - runs all up migrations
    migrate down             - reverses most recent migration
    migrate reset            - runs all down migrations, then all up migrations
    `)
}

func updateSource() {
	// walk entire project folder
	err := filepath.Walk(".", updateSourceFiles)
	if err != nil {
		exitGracefully(err)
	}
}

func updateSourceFiles(path string, fi os.FileInfo, err error) error {
	// check for error
	if err != nil {
		return err
	}

	// check if current file is directory and ignore it
	if fi.IsDir() {
		return nil
	}

	// only check go files
	matched, err := filepath.Match("*.go", fi.Name())
	if err != nil {
		return err
	}

	// have matching file
	if matched {
		// read file contents
		read, err := os.ReadFile(path)
		if err != nil {
			exitGracefully(err)
		}

		newContents := strings.Replace(string(read), "myapp", appURL, -1)

		// write changed file
		err = os.WriteFile(path, []byte(newContents), 0) // 0 don't change permissions
		if err != nil {
			exitGracefully(err)
		}
	}
	return nil
}
