package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WithTx runs fn inside a database transaction.
//   - fn returns nil   -> COMMIT
//   - fn returns error -> ROLLBACK, and the error is passed back to the caller
//
// The deferred Rollback is a harmless no-op after a successful Commit, so every
// early `return err` inside fn safely undoes all earlier writes.
func WithTx(ctx context.Context, pool *pgxpool.Pool, fn func(tx pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
