package db

import (
	"context"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/simondanielsson/recite/cmd/internal/queries"
)

type RepositoryKeyType string

const RepositoryKey RepositoryKeyType = "repo_key"

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func AddDatabaseMiddleware(handler http.Handler, pool *pgxpool.Pool, logger *log.Logger) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			tx, err := pool.BeginTx(ctx, pgx.TxOptions{
				IsoLevel:   pgx.Serializable,
				AccessMode: pgx.ReadWrite,
			})
			if err != nil {
				http.Error(w, "failed to start transaction", http.StatusInternalServerError)
			}
			repository := queries.New(pool).WithTx(tx)

			ctx = context.WithValue(ctx, RepositoryKey, repository)

			rec := statusRecorder{ResponseWriter: w, status: http.StatusOK}
			handler.ServeHTTP(rec, r.WithContext(ctx))

			if rec.status >= http.StatusBadRequest {
				_ = tx.Rollback(ctx)
			} else if err := tx.Commit(ctx); err != nil {
				// Response has already been sent - just log
				logger.Printf("commit tx failed: %v", err)
			}
		},
	)
}
