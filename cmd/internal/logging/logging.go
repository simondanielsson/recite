package logging

import (
	"io"
	"log"
)

// Logger holds a out and err logger
type Logger struct {
	Out *log.Logger
	Err *log.Logger
}

func NewLogger(outWriter io.Writer, errWriter io.Writer) Logger {
	outLogger := log.Logger{}
	outLogger.SetOutput(outWriter)

	errLogger := log.Logger{}
	errLogger.SetOutput(errWriter)
	return Logger{
		Out: &outLogger,
		Err: &errLogger,
	}
}
