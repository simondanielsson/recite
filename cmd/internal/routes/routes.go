package routes

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/simondanielsson/recite/cmd/internal/marshal"
	"github.com/simondanielsson/recite/cmd/internal/services"
)

func RegisterRoutes(mux *http.ServeMux, logger *log.Logger) {
	mux.Handle("GET /api/v1", rootGetHandler(logger))
	mux.Handle("GET /api/v1/health", rootGetHandler(logger))
	mux.Handle("POST /api/v1/recitals", recitalPostHandler(logger))
}

func rootGetHandler(logger *log.Logger) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("ok\n")); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		},
	)
}

func recitalPostHandler(logger *log.Logger) http.Handler {
	type request struct {
		Url string `json:"url"`
	}
	type response struct {
		Message string `json:"message"`
	}

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			req, err := marshal.Decode[request](r)
			if err != nil || req.Url == "" {
				res := response{
					Message: fmt.Sprintf("bad request, should contain url. Got %s", req.Url),
				}
				if err := marshal.Encode(w, r, http.StatusBadRequest, res); err != nil {
					writeErr(w, err, logger)
				}
				return
			}

			ctx := context.Background()
			if err := services.CreateRecital(ctx, req.Url); err != nil {
				writeErr(w, err, logger)
				return
			}

			res := response{
				Message: "created",
			}
			if err := marshal.Encode(w, r, http.StatusCreated, res); err != nil {
				writeErr(w, err, logger)
				return
			}
		},
	)
}

func writeErr(w http.ResponseWriter, err error, logger *log.Logger) {
	w.WriteHeader(http.StatusInternalServerError)
	if _, err := w.Write([]byte(err.Error())); err != nil {
		logger.Print("failed writing error")
	}
}
