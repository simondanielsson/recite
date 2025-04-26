package completions

import (
	"context"
	"io"
)

// Completion provides functionality to generate text or speech.
type Completion interface {
	NewSpeech(context.Context, string, string) (string, error)
	NewText(context.Context, string, string) (io.ReadCloser, func(), error)
}
