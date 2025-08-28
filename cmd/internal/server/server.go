package server

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/simondanielsson/recite/cmd/internal/config"
	"github.com/simondanielsson/recite/cmd/internal/db"
	"github.com/simondanielsson/recite/cmd/internal/logging"
	"github.com/simondanielsson/recite/cmd/internal/routes"
)

type App struct {
	Server http.Server
	DB     *pgxpool.Pool
}

func New(config config.Config, DB *pgxpool.Pool, logger logging.Logger) App {
	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, logger)

	var handler http.Handler = mux
	handler = logging.AddLoggingMiddleware(handler, logger)
	handler = db.AddDatabaseMiddleware(handler, DB, logger)

	return App{
		Server: http.Server{
			Addr:         ":" + config.Server.Addr,
			ReadTimeout:  config.Server.ReadTimeout,
			WriteTimeout: config.Server.WriteTimeout,
			Handler:      handler,
		},
		DB: DB,
	}
}

func (a *App) Shutdown(ctx context.Context) error {
	if err := a.Server.Shutdown(ctx); err != nil {
		return err
	}

	a.DB.Close()
	return nil
}
