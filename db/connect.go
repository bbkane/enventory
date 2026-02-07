package db

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"time"

	_ "modernc.org/sqlite"
)

// NOTE: most of this is heavily inspired by https://github.com/benbjohnson/wtf/blob/05bc90c940d5f9e2490fc93cf467d9e8aa48ad63/sqlite/sqlite.go

//go:embed migrations/*.sql
var migrationFS embed.FS

func Connect(ctx context.Context, dsn string) (*sql.DB, error) {
	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("db open error: %s: %w", dsn, err)
	}

	err = sqlDB.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("db ping error: %w", err)
	}

	if _, err := sqlDB.ExecContext(ctx, `PRAGMA foreign_keys = ON;`); err != nil {
		return nil, fmt.Errorf("foreign keys pragma: %w", err)
	}

	if err := migrate(ctx, sqlDB, migrationFS, "*/*.sql"); err != nil {
		return nil, fmt.Errorf("migrate connect error: %w", err)
	}

	return sqlDB, nil
}

// migrate sets up migration tracking and executes pending migration files.
//
// Migration files are embedded in the sqlite/migration folder and are executed
// in lexigraphical order.
//
// Once a migration is run, its name is stored in the 'migrations' table so it
// is not re-executed. Migrations run in a transaction to prevent partial
// migrations.
func migrate(ctx context.Context, sqlDB *sql.DB, migrationFS fs.ReadFileFS, migrationsGlobPattern string) error {

	// Create or update the migration table. Wouldn't need this if I was starting from scratch :D
	if err := withTx(ctx, sqlDB, migrateMigrationTable); err != nil {
		return fmt.Errorf("could not migrate migration table: %w", err)
	}

	// Read migration files from our embedded file system.
	// This uses Go 1.16's 'embed' package.
	names, err := fs.Glob(migrationFS, migrationsGlobPattern)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		return fmt.Errorf("no files found at: %v", migrationsGlobPattern)
	}
	sort.Strings(names)
	// fmt.Printf("migrations: %v\n", names)

	// Loop over all migration files and execute them in order.
	for _, name := range names {
		if err := migrateFile(ctx, sqlDB, migrationFS, name); err != nil {
			return fmt.Errorf("migration error: name=%q err=%w", name, err)
		}
	}
	return nil
}

// migrateMigrationTable creates the migrations table if it does not exist, and updates it if we have an old version of the table. This is a special table that tracks updates to the rest of the db, so we have to update it separately
func migrateMigrationTable(ctx context.Context, tx *sql.Tx) error {

	// Check if the migrations table exists
	migrationsTableCount := false
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='migrations';`).Scan(&migrationsTableCount); err != nil {
		return fmt.Errorf("check if migrations table exists err: %w", err)
	}

	migrationV2TableCount := false
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='migration_v2';`).Scan(&migrationV2TableCount); err != nil {
		return fmt.Errorf("check if migration_v2 table exists err: %w", err)
	}

	migrationV2Create := `
CREATE TABLE migration_v2 (
	migration_v2_id INTEGER PRIMARY KEY,
	file_name TEXT NOT NULL,
	migrate_time TEXT NOT NULL,
	UNIQUE(file_name)
) STRICT;
`

	switch {
	case migrationsTableCount && migrationV2TableCount:
		// this should never happen
		return errors.New("both migrations and migration_v2 tables exist")

	case migrationsTableCount && !migrationV2TableCount:
		if _, err := tx.ExecContext(ctx, migrationV2Create); err != nil {
			return fmt.Errorf("cannot create migration_v2 table: %w", err)
		}
		// NOTE: this relies on the embedded file system using '/' as the path separator. I think that's ok?
		insertIntoMigrationV2 := `
INSERT INTO migration_v2 (file_name, migrate_time)
SELECT replace(name, 'embedded_migrations/', '') AS file_name, ?
FROM migrations
ORDER BY name;
`
		now := time.Now().Round(0).UTC().Format(time.RFC3339)
		if _, err := tx.ExecContext(ctx, insertIntoMigrationV2, now); err != nil {
			return fmt.Errorf("cannot copy data from migrations to migration_v2: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `DROP TABLE migrations;`); err != nil {
			return fmt.Errorf("cannot drop migrations table: %w", err)
		}

		return nil

	case !migrationsTableCount && migrationV2TableCount:
		return nil

	case !migrationsTableCount && !migrationV2TableCount:
		if _, err := tx.ExecContext(ctx, migrationV2Create); err != nil {
			return fmt.Errorf("cannot create migration_v2 table: %w", err)
		}

		return nil
	}

	return errors.New("unreachable")
}

// migrate runs a single migration file within a transaction. On success, the
// migration file name is saved to the "migrations" table to prevent re-running.
func migrateFile(ctx context.Context, db *sql.DB, migrationFS fs.ReadFileFS, name string) error {
	err := withTx(
		ctx,
		db,
		func(ctx context.Context, tx *sql.Tx) error {

			fileName := filepath.Base(name)
			// Ensure migration has not already been run.
			var n int
			if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM migration_v2 WHERE file_name = ?`, fileName).Scan(&n); err != nil {
				return err
			} else if n != 0 {
				return nil // already run migration, skip
			}

			// Read and execute migration file.
			if buf, err := fs.ReadFile(migrationFS, name); err != nil {
				return err
			} else if _, err := tx.ExecContext(ctx, string(buf)); err != nil {
				return err
			}

			now := time.Now().Round(0).UTC().Format(time.RFC3339)
			// Insert record into migrations to prevent re-running migration.
			if _, err := tx.ExecContext(ctx, `INSERT INTO migration_v2 (file_name, migrate_time) VALUES (?, ?)`, fileName, now); err != nil {
				return err
			}
			return nil
		},
	)

	if err != nil {
		return fmt.Errorf("migrate file err: %w", err)
	}
	return nil
}

// withTx makes transactions easy!!
func withTx(ctx context.Context, sqlDB *sql.DB, txFunc func(ctx context.Context, tx *sql.Tx) error) error {

	tx, err := sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)

	}

	err = txFunc(ctx, tx)
	if err != nil {
		err = fmt.Errorf("err inside transaction: %w", err)
		goto rollback
	}

	err = tx.Commit()
	if err != nil {
		err = fmt.Errorf("could not commit transaction: %w", err)
		goto rollback
	}
	return nil

rollback:
	rollbackErr := tx.Rollback()
	if rollbackErr != nil {
		return fmt.Errorf("could not rollback transaction: %w after previous err: %w", rollbackErr, err)
	}
	return fmt.Errorf("transaction rolled back after err: %w", err)
}
