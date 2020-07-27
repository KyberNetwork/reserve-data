package testutil

import (
	"fmt"

	"github.com/KyberNetwork/reserve-data/lib/migration"
	_ "github.com/golang-migrate/migrate/v4/source/file" // go migrate
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // sql driver name: "postgres"
)

const (
	postgresHost     = "127.0.0.1"
	postgresPort     = 5432
	postgresUser     = "reserve_data"
	postgresPassword = "reserve_data"
)

// MustNewDevelopmentDB creates a new development DB.
// It also returns a function to teardown it after the test.
func MustNewDevelopmentDB(migrationPath string) (*sqlx.DB, func() error) {
	dbName := RandomString(8)

	ddlDBConnStr := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable",
		postgresHost,
		postgresPort,
		postgresUser,
		postgresPassword,
	)
	ddlDB := sqlx.MustConnect("postgres", ddlDBConnStr)
	ddlDB.MustExec(fmt.Sprintf(`CREATE DATABASE "%s"`, dbName))
	if err := ddlDB.Close(); err != nil {
		panic(err)
	}

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		postgresHost,
		postgresPort,
		postgresUser,
		postgresPassword,
		dbName,
	)
	db := sqlx.MustConnect("postgres", connStr)
	m, err := migration.RunMigrationUp(db.DB, migrationPath, dbName)
	if err != nil {
		panic(err)
	}
	return db, func() error {
		if _, err := m.Close(); err != nil {
			return err
		}
		ddlDB, err := sqlx.Connect("postgres", ddlDBConnStr)
		if err != nil {
			return err
		}
		if _, err = ddlDB.Exec(fmt.Sprintf(`DROP DATABASE "%s"`, dbName)); err != nil {
			return err
		}
		return ddlDB.Close()
	}
}
