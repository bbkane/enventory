package app

import (
	"context"
	"database/sql"
	"fmt"

	db "go.bbkane.com/enventory/db"
	"go.bbkane.com/enventory/models"
)

type EnvService struct {
	dbtx models.DBTX
}

func NewEnvService(ctx context.Context, dsn string) (models.Service, error) {
	dbtx, err := db.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("could not init db: %w", err)
	}
	return &EnvService{
		dbtx: models.NewTracedDBTX(models.Tracer, dbtx),
	}, nil
}

func (e *EnvService) WithTx(ctx context.Context, txFunc func(ctx context.Context, es models.Service) error) error {
	// TODO: make a WithReadOnlyTx too. Check error handling.
	// I could use https://www.reddit.com/r/golang/comments/16xswxd/concurrency_when_writing_data_into_sqlite/
	// but I really like just having a single file. I only expect one user at a time anyway.

	// This really breaks the encapsulation of the EnvService, but I need to begin a transaction. Maybe I'll find a better way later
	var dbtx models.DBTX
	if tx, ok := e.dbtx.(*models.TracedDBTX); ok {
		dbtx = tx.Unwrap()
	} else {
		dbtx = e.dbtx
	}
	db, ok := dbtx.(*sql.DB)
	if !ok {
		panic("EnvService.dbtx is not a *sql.DB, cannot begin transaction. Maybe you tried to nest WithTx?")
	}
	// this is slightly copied from sqliteconnect.Connect
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}

	es := &EnvService{dbtx: models.NewTracedDBTX(models.Tracer, tx)}
	err = txFunc(ctx, es)
	if err != nil {
		err = fmt.Errorf("error in transaction: %w", err)
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
