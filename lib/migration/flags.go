package migration

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/urfave/cli"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
)

const (
	migrationPathFlag = "migration-path"
)

//NewMigrationFolderPathFlag return new flag for migration folder
func NewMigrationFolderPathFlag() cli.Flag {
	return cli.StringFlag{
		Name:   migrationPathFlag,
		Usage:  "path for migration files",
		EnvVar: "MIGRATION_PATH",
		Value:  "migrations",
	}
}

// NewMigrationPathFromContext return migration folder path
func NewMigrationPathFromContext(c *cli.Context) string {
	return c.String(migrationPathFlag)
}

// RunMigrationUp ...
func RunMigrationUp(db *sqlx.DB, migrationFolderPath, databaseName string) error {
	driver, err := migratepostgres.WithInstance(db.DB, &migratepostgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationFolderPath),
		databaseName, driver,
	)
	if err != nil {
		return err
	}
	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

// RunMigrationDown ...
func RunMigrationDown(db *sqlx.DB, migrationFolderPath, databaseName string) error {
	driver, err := migratepostgres.WithInstance(db.DB, &migratepostgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationFolderPath),
		databaseName, driver,
	)
	if err != nil {
		return err
	}
	if err = m.Down(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
