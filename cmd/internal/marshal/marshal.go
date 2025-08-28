package marshal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	constants "github.com/simondanielsson/recite/cmd/internal"
)

func Encode[T any](ctx context.Context, w http.ResponseWriter, r *http.Request, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// A bit hacky way to write override the existing context
	ctx = context.WithValue(ctx, constants.StatusCodeKey, status)
	*r = *(r.WithContext(ctx))

	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

func Decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}
