package main

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/iancoleman/strcase"
	"os"
	"strings"
	"time"
)

func doMake(arg2, arg3 string) error {
	switch arg2 {
	case "auth":
		err := doAuth()
		if err != nil {
			exitGracefully(err)
		}
	case "handler":
		if arg3 == "" {
			exitGracefully(errors.New("you must give the handler a name"))
		}

		fileName := rap.RootPath + "/handlers/" + strings.ToLower(arg3) + ".go"
		if fileExists(fileName) {
			exitGracefully(errors.New("file " + fileName + " already exists"))
		}

		data, err := templateFS.ReadFile("templates/handlers/handler.go.txt")
		if err != nil {
			exitGracefully(err)
		}

		handler := string(data)
		handler = strings.ReplaceAll(handler, "$HANDLERNAME$", strcase.ToCamel(arg3))

		err = os.WriteFile(fileName, []byte(handler), 0644)
		if err != nil {
			exitGracefully(err)
		}
	case "key":
		rnd := rap.RandomString(32)
		color.Yellow("32 character encryption key: %s", rnd)
	case "mail":
		if arg3 == "" {
			exitGracefully(errors.New("you must give the mail template a name"))
		}
		htmlMail := rap.RootPath + "/mail/" + strings.ToLower(arg3) + ".html.tmpl"
		textMail := rap.RootPath + "/mail/" + strings.ToLower(arg3) + ".text.tmpl"
		if fileExists(htmlMail) {
			exitGracefully(errors.New("file " + htmlMail + " already exists"))
		}
		if fileExists(textMail) {
			exitGracefully(errors.New("file " + textMail + " already exists"))
		}

		err := copyFileFromTemplate("templates/mailer/mail.html.tmpl", htmlMail)
		if err != nil {
			exitGracefully(err)
		}

		err = copyFileFromTemplate("templates/mailer/mail.text.tmpl", textMail)
		if err != nil {
			exitGracefully(err)
		}
	case "migration":
		dbType := rap.DB.Type
		if arg3 == "" {
			exitGracefully(errors.New("you must give the migration a name"))
		}

		fileName := fmt.Sprintf("%d_%s", time.Now().UnixMicro(), arg3)
		upFile := rap.RootPath + "/migrations/" + fileName + "." + dbType + ".up.sql"
		downFile := rap.RootPath + "/migrations/" + fileName + "." + dbType + ".down.sql"

		err := copyFileFromTemplate("templates/migrations/migration."+dbType+".up.sql", upFile)
		if err != nil {
			exitGracefully(err)
		}

		err = copyFileFromTemplate("templates/migrations/migration."+dbType+".down.sql", downFile)
		if err != nil {
			exitGracefully(err)
		}
	case "model":
		if arg3 == "" {
			exitGracefully(errors.New("you must give the model a name"))
		}

		data, err := templateFS.ReadFile("templates/data/model.go.txt")
		if err != nil {
			exitGracefully(err)
		}

		model := string(data)
		var modelName = arg3
		var tableName = arg3

		fileName := rap.RootPath + "/data/" + strings.ToLower(modelName) + ".go"
		if fileExists(fileName) {
			exitGracefully(errors.New("file " + fileName + " already exists"))
		}

		model = strings.ReplaceAll(model, "$MODELNAME$", strcase.ToCamel(modelName))
		model = strings.ReplaceAll(model, "$TABLENAME$", tableName)

		err = copyDataToFile([]byte(model), fileName)
		if err != nil {
			exitGracefully(err)
		}
	case "session":
		err := doSessionTable()
		if err != nil {
			exitGracefully(err)
		}

	}

	return nil
}
