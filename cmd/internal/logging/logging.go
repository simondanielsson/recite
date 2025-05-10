package logging

import (
	"io"
	"log"
	"net/http"
)

func NewLogger(w io.Writer) *log.Logger {
	logger := log.Logger{}
	logger.SetOutput(w)
	logger.SetPrefix("HELLO PREFIX | ")
	return &logger
}

func AddLoggingMiddleware(handler http.Handler, logger *log.Logger) http.Handler {
	return handler
}
