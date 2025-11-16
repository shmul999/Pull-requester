package migrations

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"strconv"
	"strings"
)

type Migration struct {
	Version int
	Up      string
	Down    string
}

func Run(db *sql.DB, migrationsFS fs.FS) error {
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	applied, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	migrations, err := loadMigrations(migrationsFS)
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Применяем миграции в порядке версий
	for _, migration := range migrations {
		if _, exists := applied[migration.Version]; !exists {
			log.Printf("Applying migration %d", migration.Version)
			
			tx, err := db.Begin()
			if err != nil {
				return fmt.Errorf("failed to begin transaction: %w", err)
			}

			if _, err := tx.Exec(migration.Up); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to execute migration %d: %w", migration.Version, err)
			}

			if err := markMigrationApplied(tx, migration.Version); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to mark migration %d as applied: %w", migration.Version, err)
			}

			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit migration %d: %w", migration.Version, err)
			}

			log.Printf("Migration %d applied successfully", migration.Version)
		}
	}

	return nil
}

func createMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

func getAppliedMigrations(db *sql.DB) (map[int]bool, error) {
	rows, err := db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}
	return applied, rows.Err()
}

func loadMigrations(migrationsFS fs.FS) ([]Migration, error) {
	entries, err := fs.ReadDir(migrationsFS, ".")
	if err != nil {
		return nil, err
	}

	var migrations []Migration
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			versionStr := strings.Split(entry.Name(), "_")[0]
			version, err := strconv.Atoi(versionStr)
			if err != nil {
				return nil, fmt.Errorf("invalid migration file name: %s", entry.Name())
			}

			upSQL, err := fs.ReadFile(migrationsFS, entry.Name())
			if err != nil {
				return nil, err
			}

			downFile := strings.Replace(entry.Name(), ".up.sql", ".down.sql", 1)
			downSQL, _ := fs.ReadFile(migrationsFS, downFile) // down migration optional

			migrations = append(migrations, Migration{
				Version: version,
				Up:      string(upSQL),
				Down:    string(downSQL),
			})
		}
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func markMigrationApplied(tx *sql.Tx, version int) error {
	_, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version)
	return err
}