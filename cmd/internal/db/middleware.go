package db

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	constants "github.com/simondanielsson/recite/cmd/internal"
	"github.com/simondanielsson/recite/cmd/internal/logging"
	"github.com/simondanielsson/recite/cmd/internal/queries"
)

func AddDatabaseMiddleware(handler http.Handler, pool *pgxpool.Pool, logger logging.Logger) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			repository, commit, rollback, err := NewRepositoryWithTx(ctx, pool)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			// Attach transaction-scoped repository, and pool for transactions to be launched outside request context
			ctx = context.WithValue(ctx, constants.RepositoryKey, repository)
			ctx = context.WithValue(ctx, constants.DBConnPool, pool)
			r = r.WithContext(ctx)

			handler.ServeHTTP(w, r)

			status, ok := r.Context().Value(constants.StatusCodeKey).(int)
			if !ok {
				logger.Err.Println("No status code found in context, rolling back tx")
				_ = rollback(ctx)
				return
			}

			if status >= http.StatusBadRequest {
				_ = rollback(ctx)
			} else if err := commit(ctx); err != nil {
				// Response has already been sent - just log
				logger.Err.Printf("tx commit failed: %v", err)
			}
		},
	)
}

func NewRepositoryWithTx(ctx context.Context, pool *pgxpool.Pool) (q *queries.Queries, commit func(context.Context) error, rollback func(context.Context) error, err error) {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	commit = func(ctx context.Context) error { return tx.Commit(ctx) }
	rollback = func(ctx context.Context) error { return tx.Rollback(ctx) }
	return queries.New(pool).WithTx(tx), commit, rollback, nil
}
