package rapidus

import (
	"github.com/golang-migrate/migrate/v4"
	"log"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func (r *Rapidus) MigrateUp(dsn string) error {
	rootPath := filepath.ToSlash(r.RootPath)
	m, err := migrate.New("file://"+rootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		log.Println("error running migration: ", err)
		return err
	}

	return nil
}

func (r *Rapidus) MigrateDownAll(dsn string) error {
	rootPath := filepath.ToSlash(r.RootPath)
	m, err := migrate.New("file://"+rootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Down(); err != nil {
		return err
	}

	return nil
}

func (r *Rapidus) Steps(n int, dsn string) error {
	rootPath := filepath.ToSlash(r.RootPath)
	m, err := migrate.New("file://"+rootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Steps(n); err != nil {
		return err
	}

	return nil
}

func (r *Rapidus) MigrateForce(dsn string) error {
	rootPath := filepath.ToSlash(r.RootPath)
	m, err := migrate.New("file://"+rootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Force(-1); err != nil {
		return err
	}

	return nil
}
