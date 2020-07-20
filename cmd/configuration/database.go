package configuration

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // sql driver name: "postgres"
	"github.com/urfave/cli"
)

const (
	postgresHostFlag    = "postgres-host"
	defaultPostgresHost = "127.0.0.1"

	postgresPortFlag    = "postgres-port"
	defaultPostgresPort = 5432

	postgresUserFlag    = "postgres-user"
	defaultPostgresUser = "reserve_data"

	postgresPasswordFlag    = "postgres-password"
	defaultPostgresPassword = "reserve_data"

	postgresDatabaseFlag = "postgres-database"
)

// NewPostgreSQLFlags creates new cli flags for PostgreSQL client.
func NewPostgreSQLFlags(defaultDB string) []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   postgresHostFlag,
			Usage:  "PostgreSQL host to connect",
			EnvVar: "POSTGRES_HOST",
			Value:  defaultPostgresHost,
		},
		cli.IntFlag{
			Name:   postgresPortFlag,
			Usage:  "PostgreSQL port to connect",
			EnvVar: "POSTGRES_PORT",
			Value:  defaultPostgresPort,
		},
		cli.StringFlag{
			Name:   postgresUserFlag,
			Usage:  "PostgreSQL user to connect",
			EnvVar: "POSTGRES_USER",
			Value:  defaultPostgresUser,
		},
		cli.StringFlag{
			Name:   postgresPasswordFlag,
			Usage:  "PostgreSQL password to connect",
			EnvVar: "POSTGRES_PASSWORD",
			Value:  defaultPostgresPassword,
		},
		cli.StringFlag{
			Name:   postgresDatabaseFlag,
			Usage:  "Postgres database to connect",
			EnvVar: "POSTGRES_DATABASE",
			Value:  defaultDB,
		},
	}
}

// NewDBFromContext creates a DB instance from cli flags configuration.
func NewDBFromContext(c *cli.Context) (*sqlx.DB, error) {
	const driverName = "postgres"
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.String(postgresHostFlag),
		c.Int(postgresPortFlag),
		c.String(postgresUserFlag),
		c.String(postgresPasswordFlag),
		c.String(postgresDatabaseFlag),
	)
	return sqlx.Connect(driverName, connStr)
}

// DatabaseNameFromContext return database name
func DatabaseNameFromContext(c *cli.Context) string {
	return c.String(postgresDatabaseFlag)
}
