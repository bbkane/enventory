package app

import (
	"context"
	"database/sql"
	"fmt"

	"go.bbkane.com/enventory/app/sqliteconnect"
	"go.bbkane.com/enventory/models"
)

type EnvService struct {
	dbtx models.DBTX
}

func NewEnvService(ctx context.Context, dsn string) (models.EnvService, error) {
	db, err := sqliteconnect.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("could not init db: %w", err)
	}
	return &EnvService{
		dbtx: db,
	}, nil
}

func (e *EnvService) WithTx(ctx context.Context, fn func(ctx context.Context, es models.EnvService) error) error {
	// TODO: make a WithReadOnlyTx too. Check error handling.
	// I could use https://www.reddit.com/r/golang/comments/16xswxd/concurrency_when_writing_data_into_sqlite/
	// but I really like just having a single file. I only expect one user at a time anyway.
	db, ok := e.dbtx.(*sql.DB)
	if !ok {
		panic("EnvService.dbtx is not a *sql.DB, cannot begin transaction. Maybe you tried to next WithTx?")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	es := &EnvService{dbtx: tx}
	if err := fn(ctx, es); err != nil {
		return fmt.Errorf("error in transaction: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %w", err)
	}
	return nil
}
