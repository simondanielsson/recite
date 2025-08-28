package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/simondanielsson/recite/cmd/internal/config"
)

func New(ctx context.Context, config config.Config) (*pgxpool.Pool, error) {
	connString := fmt.Sprintf("user=%s dbname=%s password=%s port=%s host=%s sslmode=disable", config.DB.User, config.DB.Name, config.DB.Password, config.DB.Port, config.DB.Host)
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("could not connect to db: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed pinging database: %w", err)
	}

	return pool, nil
}
