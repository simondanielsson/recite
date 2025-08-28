package routes

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	constants "github.com/simondanielsson/recite/cmd/internal"
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
				logger.Println(err)
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
				if err := marshal.Encode(r.Context(), w, r, http.StatusBadRequest, res); err != nil {
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
			repository, ok := ctx.Value(constants.RepositoryKey).(*queries.Queries)
			if !ok {
				logGenericInternalServiceError(ctx, w, r, fmt.Errorf("failed loading repository"), logger)
				return
			}
			pool, ok := ctx.Value(constants.DBConnPool).(*pgxpool.Pool)
			if !ok {
				logGenericInternalServiceError(ctx, w, r, fmt.Errorf("failed loading db connection pool"), logger)
				return
			}

			id, err := services.CreateRecital(ctx, req.Url, repository, pool, logger)
			if err != nil {
				logGenericInternalServiceError(ctx, w, r, err, logger)
				return
			}

			res := response{
				Id: id,
			}
			if err := marshal.Encode(ctx, w, r, http.StatusCreated, res); err != nil {
				writeErrHeader(w, err, logger)
				return
			}
		},
	)
}

func recitalGetHandler(logger *log.Logger) http.Handler {
	type response struct {
		queries.Recital
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			res := messageResponse{Message: "invalid id"}
			if err := marshal.Encode(r.Context(), w, r, http.StatusBadRequest, res); err != nil {
				writeErrHeader(w, err, logger)
			}
			return
		}

		ctx := r.Context()
		repository, ok := ctx.Value(constants.RepositoryKey).(*queries.Queries)
		if !ok {
			logGenericInternalServiceError(ctx, w, r, fmt.Errorf("failed loading repository"), logger)
		}

		recital, err := services.GetRecital(ctx, int32(id), repository, logger)
		if err != nil {
			res := messageResponse{Message: fmt.Sprintf("Could not find recital with id %d", id)}
			if err := marshal.Encode(ctx, w, r, http.StatusNotFound, res); err != nil {
				logFailedEncodingResponse(ctx, w, r, err, logger)
			}
			return
		}

		res := response{
			Recital: recital,
		}
		if err := marshal.Encode(ctx, w, r, http.StatusOK, res); err != nil {
			logFailedEncodingResponse(ctx, w, r, err, logger)
		}
	})
}

func writeErrHeader(w http.ResponseWriter, err error, logger *log.Logger) {
	w.WriteHeader(http.StatusInternalServerError)
	if _, err := w.Write([]byte(err.Error())); err != nil {
		logger.Print("failed writing error")
	}
}

func logGenericInternalServiceError(ctx context.Context, w http.ResponseWriter, r *http.Request, err error, logger *log.Logger) {
	res := messageResponse{Message: "Something went wrong."}
	logger.Print(err)
	if err := marshal.Encode(ctx, w, r, int(http.StatusInternalServerError), res); err != nil {
		writeErrHeader(w, err, logger)
		return
	}
}

func logFailedEncodingResponse(ctx context.Context, w http.ResponseWriter, r *http.Request, err error, logger *log.Logger) {
	logger.Println(err)
	res := messageResponse{"failed to encode response"}
	if err := marshal.Encode(ctx, w, r, http.StatusInternalServerError, res); err != nil {
		writeErrHeader(w, err, logger)
	}
}
