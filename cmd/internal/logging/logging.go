package logging

import (
	"io"
	"log"
)

const DateFormat string = "Mon Jan 2 15:04:05 MST 2006"

func NewLogger(w io.Writer) *log.Logger {
	logger := log.Logger{}
	logger.SetOutput(w)
	return &logger
}
