package logging

import (
	"net/http"
	"time"

	constants "github.com/simondanielsson/recite/cmd/internal"
)

func AddLoggingMiddleware(handler http.Handler, logger Logger) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			handler.ServeHTTP(w, r)
			status, ok := r.Context().Value(constants.StatusCodeKey).(int)
			if !ok {
				status = -1
				logger.Err.Println("Could not get status code from context")
			}

			logger.Out.Printf("%s | %s %s - %d\n", time.Now().Format(constants.DateFormat), r.Method, r.URL, status)
		},
	)
}
