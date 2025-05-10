package server

import (
	"log"
	"net/http"

	"github.com/simondanielsson/recite/cmd/internal/config"
	"github.com/simondanielsson/recite/cmd/internal/logging"
)

type Server struct {
	server http.Server
}

func New(config config.Config, logger *log.Logger) http.Server {
	mux := http.NewServeMux()
	// Register routes
	mux.Handle("/", rootMux(logger))

	var handler http.Handler = mux
	handler = logging.AddLoggingMiddleware(handler, logger)

	// TODO: might not need this wrapping
	return http.Server{
		Addr:         ":" + config.Server.Addr,
		ReadTimeout:  config.Server.ReadTimeout,
		WriteTimeout: config.Server.WriteTimeout,
		Handler:      handler,
	}
}

func rootMux(logger *log.Logger) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			logger.Println("CALLED /")
		},
	)
}
