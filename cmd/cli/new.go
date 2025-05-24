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

	// sanitize appName
	if strings.Contains(appName, "/") {
		exploded := strings.SplitAfter(appName, "/")
		appName = exploded[len(exploded)-1]
	}
	log.Println("App name is", appName)

	// git clone skeleton application
	color.Yellow("\tCloning repository...")
	_, err := git.PlainClone("./"+appName, false, &git.CloneOptions{
		URL:      "https://github.com/fouched/rapidus-app.git",
		Progress: os.Stdout,
		Depth:    1,
	})
	if err != nil {
		exitGracefully(err)
	}

	// remove .git directory
	err = os.RemoveAll(fmt.Sprintf("./%s/.git", appName))
	if err != nil {
		exitGracefully(err)
	}

	// create a ready to go .env file
	color.Yellow("\tCreating .env file...")
	data, err := templateFS.ReadFile("templates/env.txt")
	if err != nil {
		exitGracefully(err)
	}

	env := string(data)
	env = strings.ReplaceAll(env, "${APP_NAME}", appName)
	env = strings.ReplaceAll(env, "${KEY}", rap.RandomString(32))

	err = copyDataToFile([]byte(env), fmt.Sprintf("./%s/.env", appName))
	if err != nil {
		exitGracefully(err)
	}

	// create a makefile
	source, err := os.Open(fmt.Sprintf("./%s/Makefile.linux", appName))
	if runtime.GOOS == "windows" {
		source, err = os.Open(fmt.Sprintf("./%s/Makefile.windows", appName))
	}
	if err != nil {
		exitGracefully(err)
	}
	defer source.Close()

	destination, err := os.Create(fmt.Sprintf("./%s/Makefile", appName))
	if err != nil {
		exitGracefully(err)
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		exitGracefully(err)
	}

	_ = os.Remove(fmt.Sprintf("./%s/Makefile.linux", appName))
	_ = os.Remove(fmt.Sprintf("./%s/Makefile.windows", appName))

	// update go.mod file
	color.Yellow("\tCreating go.mod file...")
	_ = os.Remove(fmt.Sprintf("./%s/go.mod", appName))

	data, err = templateFS.ReadFile("templates/go.mod.txt")
	if err != nil {
		exitGracefully(err)
	}

	mod := string(data)
	mod = strings.ReplaceAll(mod, "${APP_NAME}", appURL)

	err = copyDataToFile([]byte(mod), "./"+appName+"/go.mod")

	// update existing .go files with correct name/imports
	color.Yellow("\tUpdating source files...")
	os.Chdir("./" + appName)
	updateSource()

	// run go mod tidy in project directory
	color.Yellow("\tRunning go mod tidy...")
	cmd := exec.Command("go", "mod", "tidy")
	err = cmd.Start()
	if err != nil {
		exitGracefully(err)
	}

	color.Green("Done building " + appURL)
	color.Green("Go build something awesome!")
}
