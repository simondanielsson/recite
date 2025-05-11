package server

import (
	"log"
	"net/http"

	"github.com/simondanielsson/recite/cmd/internal/config"
	"github.com/simondanielsson/recite/cmd/internal/logging"
	"github.com/simondanielsson/recite/cmd/internal/routes"
)

func New(config config.Config, logger *log.Logger) http.Server {
	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, logger)

	var handler http.Handler = mux
	handler = logging.AddLoggingMiddleware(handler, logger)

	return http.Server{
		Addr:         ":" + config.Server.Addr,
		ReadTimeout:  config.Server.ReadTimeout,
		WriteTimeout: config.Server.WriteTimeout,
		Handler:      handler,
	}
}
