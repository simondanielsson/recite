package routes

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/simondanielsson/recite/cmd/internal/db"
	"github.com/simondanielsson/recite/cmd/internal/marshal"
	"github.com/simondanielsson/recite/cmd/internal/queries"
	"github.com/simondanielsson/recite/cmd/internal/services"
)

type messageResponse struct {
	Message string `json:"message"`
}

func RegisterRoutes(mux *http.ServeMux, logger *log.Logger) {
	mux.Handle("GET /api/v1", rootGetHandler(logger))
	mux.Handle("GET /api/v1/health", rootGetHandler(logger))
	mux.Handle("POST /api/v1/recitals", recitalPostHandler(logger))
	mux.Handle("GET /api/v1/recitals/{id}", recitalGetHandler(logger))
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
		Id int `json:"id"`
	}

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			req, err := marshal.Decode[request](r)
			if err != nil || req.Url == "" {
				res := messageResponse{
					Message: fmt.Sprintf("bad request, should contain url. Got %s", req.Url),
				}
				if err := marshal.Encode(w, r, http.StatusBadRequest, res); err != nil {
					writeErrHeader(w, err, logger)
				}
				return
			}

			// TODO:
			// 2. Persist a raw row in the generations table ot get id, let path to generation be empty
			//    Once recital has been generated in the background, update the path in the db
			// 3. Create endpoint for streaming the adio from the file - perhaps WS?
			// 4. Build a simple voice streamer in a basic htmx web page

			ctx := r.Context()
			repository, ok := ctx.Value(db.RepositoryKey).(*queries.Queries)
			if !ok {
				res := messageResponse{Message: "Something went wrong."}
				logger.Print(err)
				if err := marshal.Encode(w, r, int(http.StatusInternalServerError), res); err != nil {
					writeErrHeader(w, err, logger)
					return
				}
			}

			id, err := services.CreateRecital(ctx, req.Url, repository, logger)
			if err != nil {
				res := messageResponse{Message: "Something went wrong."}
				logger.Print(err)
				if err := marshal.Encode(w, r, int(http.StatusInternalServerError), res); err != nil {
					writeErrHeader(w, err, logger)
					return
				}
			}

			res := response{
				Id: id,
			}
			if err := marshal.Encode(w, r, http.StatusCreated, res); err != nil {
				writeErrHeader(w, err, logger)
				return
			}
		},
	)
}

// TODO: move
type GenerationStatus string

const (
	InProgress GenerationStatus = "in progress"
	Completed  GenerationStatus = "completed"
	Failed     GenerationStatus = "failed"
)

func recitalGetHandler(logger *log.Logger) http.Handler {
	type response struct {
		Status GenerationStatus
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			res := messageResponse{Message: "invalid id"}
			if err := marshal.Encode(w, r, http.StatusBadRequest, res); err != nil {
				writeErrHeader(w, err, logger)
				return
			}
		}

		_ = id
		// TODO: unmock
		res := response{Status: Completed}
		if err := marshal.Encode(w, r, http.StatusOK, res); err != nil {
			res := messageResponse{"failed to encode response"}
			if err := marshal.Encode(w, r, http.StatusInternalServerError, res); err != nil {
				writeErrHeader(w, err, logger)
			}
		}

		return
	})
}

func writeErrHeader(w http.ResponseWriter, err error, logger *log.Logger) {
	w.WriteHeader(http.StatusInternalServerError)
	if _, err := w.Write([]byte(err.Error())); err != nil {
		logger.Print("failed writing error")
	}
}
