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
func MustNewDevelopmentDB() (*sqlx.DB, func() error) {
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
	if err := migration.RunMigrationUp(db, "/Users/gin/go/src/github.com/KyberNetwork/reserve-data/cmd/migrations", dbName); err != nil {
		panic(err)
	}
	return db, func() error {
		if err := db.Close(); err != nil {
			return err
		}
		ddlDB, err := sqlx.Connect("postgres", ddlDBConnStr)
		if err != nil {
			return err
		}
		// close all connections to db
		query := `SELECT pg_terminate_backend(pg_stat_activity.pid)
    FROM pg_stat_activity
    WHERE pg_stat_activity.datname = $1
			AND pid <> pg_backend_pid();`
		if _, err = ddlDB.Exec(query, dbName); err != nil {
			return err
		}
		if _, err = ddlDB.Exec(fmt.Sprintf(`DROP DATABASE "%s"`, dbName)); err != nil {
			return err
		}
		return ddlDB.Close()
	}
}
