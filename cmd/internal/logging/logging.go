package logging

import (
	"io"
	"log"
)

const DateFormat string = "Mon Jan 2 15:04:05 MST 2006"

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
