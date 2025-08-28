package logging

import (
	"log"
	"net/http"
	"time"

	constants "github.com/simondanielsson/recite/cmd/internal"
)

func AddLoggingMiddleware(handler http.Handler, logger *log.Logger) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			handler.ServeHTTP(w, r)
			status, ok := r.Context().Value(constants.StatusCodeKey).(int)
			if !ok {
				// TODO: write proper error, or assume it failed
				panic("could not get status")
			}

			logger.Printf("%s | %s %s - %d\n", time.Now().Format(DateFormat), r.Method, r.URL, status)
		},
	)
}
