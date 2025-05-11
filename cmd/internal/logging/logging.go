package logging

import (
	"io"
	"log"
	"net/http"
	"time"
)

func NewLogger(w io.Writer) *log.Logger {
	logger := log.Logger{}
	logger.SetOutput(w)
	return &logger
}

func AddLoggingMiddleware(handler http.Handler, logger *log.Logger) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			handler.ServeHTTP(w, r)
			logger.Printf("%s | %s %s\n", time.Now().Format("Mon Jan 2 15:04:05 MST 2006"), r.Method, r.URL)
		},
	)
}
