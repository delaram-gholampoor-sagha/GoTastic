package boot

import (
	"database/sql"
	"fmt"

	"github.com/delaram/GoTastic/pkg/logger"

	"github.com/golang-migrate/migrate/v4"
	mysql "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(db *sql.DB, dbName string, log logger.Logger) error {
	driver, err := mysql.WithInstance(db, &mysql.Config{
		MigrationsTable: "schema_migrations",
		DatabaseName:    dbName,
		// NoLock:       false, // (optional) set to true if you want to skip GET_LOCK
	})
	if err != nil {
		return fmt.Errorf("migrate mysql driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", dbName, driver)
	if err != nil {
		return fmt.Errorf("migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}
	log.Info("DB migrations applied (or already up-to-date)")
	return nil
}
