package db

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	constants "github.com/simondanielsson/recite/cmd/internal"
	"github.com/simondanielsson/recite/cmd/internal/queries"
)

func AddDatabaseMiddleware(handler http.Handler, pool *pgxpool.Pool, logger *log.Logger) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			repository, tx, err := NewRepositoryWithTx(ctx, pool)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			// Attach transaction-scoped repository, and pool for transactions to be launched outside request context
			ctx = context.WithValue(ctx, constants.RepositoryKey, repository)
			ctx = context.WithValue(ctx, constants.DBConnPool, pool)

			handler.ServeHTTP(w, r.WithContext(ctx))

			status, ok := ctx.Value(constants.StatusCodeKey).(int)
			if !ok {
				_ = tx.Rollback(ctx)
				return
			}

			if status >= http.StatusBadRequest {
				_ = tx.Rollback(ctx)
			} else if err := tx.Commit(ctx); err != nil {
				// Response has already been sent - just log
				logger.Printf("tx commit failed: %v", err)
			}
		},
	)
}

func NewRepositoryWithTx(ctx context.Context, pool *pgxpool.Pool) (*queries.Queries, pgx.Tx, error) {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	return queries.New(pool).WithTx(tx), tx, nil
}
