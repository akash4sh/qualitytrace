package testdb

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/intelops/qualitytrace/server/migrations"
)

type postgresDB struct {
	db *sql.DB
}

func Postgres(options ...PostgresOption) (*postgresDB, error) {
	ps := &postgresDB{}
	for _, option := range options {
		err := option(ps)
		if err != nil {
			return nil, err
		}
	}

	err := ps.ensureLatestMigration()
	if err != nil {
		return nil, fmt.Errorf("could not execute migrations: %w", err)
	}

	return ps, nil
}

func (p *postgresDB) ensureLatestMigration() error {
	driver, err := postgres.WithInstance(p.db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not get driver from postgres connection: %w", err)
	}
	sourceDriver, err := iofs.New(migrations.Migrations, ".")
	if err != nil {
		log.Fatal(err)
	}

	migrateClient, err := migrate.NewWithInstance("iofs", sourceDriver, "qualitytrace", driver)
	if err != nil {
		return fmt.Errorf("could not get migration client: %w", err)
	}

	err = migrateClient.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run migrations: %w", err)
	}

	return nil
}

func (td *postgresDB) Drop() error {
	return dropTables(
		td,
		"test_suite_run_steps",
		"test_suite_runs",
		"test_suite_steps",
		"test_suites",
		"test_runs",
		"tests",
		"environments",
		"data_stores",
		"server",
		"schema_migrations",
	)
}

func dropTables(td *postgresDB, tables ...string) error {
	for _, table := range tables {
		_, err := td.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s;", table))
		if err != nil {
			return err
		}
	}

	return nil
}

func (td *postgresDB) Close() error {
	return td.db.Close()
}
