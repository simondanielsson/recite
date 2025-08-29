package routes

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	constants "github.com/simondanielsson/recite/cmd/internal"
	"github.com/simondanielsson/recite/cmd/internal/logging"
	"github.com/simondanielsson/recite/cmd/internal/marshal"
	"github.com/simondanielsson/recite/cmd/internal/queries"
	"github.com/simondanielsson/recite/cmd/internal/services"
)

type messageResponse struct {
	Message string `json:"message"`
}

func RegisterRoutes(mux *http.ServeMux, logger logging.Logger) {
	mux.Handle("GET /api/v1", healthcheckHandler(logger))
	mux.Handle("GET /api/v1/health", healthcheckHandler(logger))
	mux.Handle("POST /api/v1/recitals", createRecitalHandler(logger))
	mux.Handle("GET /api/v1/recitals", listRecitalsHandler(logger))
	mux.Handle("GET /api/v1/recitals/{id}", getRecitalHandler(logger))
	mux.Handle("DELETE /api/v1/recitals/{id}", deleteRecitalHandler(logger))
	mux.Handle("GET /api/v1/recitals/{id}/listen", mainpageHandler(logger))
	mux.Handle("GET /api/v1/recitals/{id}/audio", streamRecitalAudioHandler(logger))
	mux.Handle("POST /api/v1/users", createUserHandler(logger))
	mux.Handle("GET /api/v1/auth", loginUserHandler(logger))
	mux.Handle("GET /api/v1/openapi.json", swaggerHandler(logger))
}

func mainpageHandler(logger logging.Logger) http.Handler {
	page := template.Must(template.New("listen").Parse(`
	<!doctype html>
	<meta charset="utf-8" />
	<title>Listen</title>
	<body style="font-family: system-ui; padding: 2rem;">
		<h1>Listen</h1>
		<p>Now playing: {{.ID}}</p>
		<audio controls preload="auto" style="width: 100%;" src="/api/v1/recitals/{{.ID}}/audio"></audio>
	</body>
	`))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			res := messageResponse{Message: "invalid id"}
			if err := marshal.Encode(r.Context(), w, r, http.StatusBadRequest, res); err != nil {
				writeErrHeader(w, err, logger)
			}
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if err := page.Execute(w, struct{ ID int }{ID: id}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func healthcheckHandler(logger logging.Logger) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("ok\n")); err != nil {
				logger.Err.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
			}
		},
	)
}

func createRecitalHandler(logger logging.Logger) http.Handler {
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
				message := fmt.Sprintf("bad request, should contain url. Got %s", req.Url)
				repondWithErrorMessage(w, r, message, http.StatusBadRequest, logger)
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
				respondWithOpaqueMessage(ctx, w, r, fmt.Errorf("failed loading repository"), logger)
				return
			}
			pool, ok := ctx.Value(constants.DBConnPool).(*pgxpool.Pool)
			if !ok {
				respondWithOpaqueMessage(ctx, w, r, fmt.Errorf("failed loading db connection pool"), logger)
				return
			}

			id, err := services.CreateRecital(ctx, req.Url, repository, pool, logger)
			if err != nil {
				respondWithOpaqueMessage(ctx, w, r, err, logger)
				return
			}

			res := response{
				Id: id,
			}
			w.Header().Add("Location", fmt.Sprintf("/api/v1/recitals/%d", id))
			if err := marshal.Encode(ctx, w, r, http.StatusCreated, res); err != nil {
				writeErrHeader(w, err, logger)
				return
			}
		},
	)
}

func listRecitalsHandler(logger logging.Logger) http.Handler {
	type response []queries.Recital
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		offset, err := readIntQueryParam("offset", 0, w, r, logger)
		if err != nil {
			return
		}

		limit, err := readIntQueryParam("limit", 30, w, r, logger)
		if err != nil {
			return
		}

		ctx := r.Context()
		repository, ok := ctx.Value(constants.RepositoryKey).(*queries.Queries)
		if !ok {
			respondWithOpaqueMessage(ctx, w, r, fmt.Errorf("failed loading repository"), logger)
			return
		}

		recitals, err := services.ListRecitals(ctx, int32(limit), int32(offset), repository, logger)
		if err != nil {
			respondWithOpaqueMessage(ctx, w, r, err, logger)
			return
		}

		if err := marshal.Encode(ctx, w, r, int(http.StatusOK), recitals); err != nil {
			writeErrHeader(w, err, logger)
			return
		}
	})
}

func getRecitalHandler(logger logging.Logger) http.Handler {
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
			respondWithOpaqueMessage(ctx, w, r, fmt.Errorf("failed loading repository"), logger)
		}

		recital, err := services.GetRecital(ctx, int32(id), repository, logger)
		if err != nil {
			res := messageResponse{Message: fmt.Sprintf("Could not find recital with id %d", id)}
			if err := marshal.Encode(ctx, w, r, http.StatusNotFound, res); err != nil {
				logFailedEncodingResponse(ctx, w, r, err, logger)
			}
			return
		}

		if err := marshal.Encode(ctx, w, r, http.StatusOK, recital); err != nil {
			logFailedEncodingResponse(ctx, w, r, err, logger)
		}
	})
}

func deleteRecitalHandler(logger logging.Logger) http.Handler {
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
			respondWithOpaqueMessage(ctx, w, r, fmt.Errorf("failed loading repository"), logger)
		}
		if err := services.DeleteRecital(ctx, int32(id), repository, logger); err != nil {
			res := messageResponse{Message: fmt.Sprintf("Could not find recital with id %d", id)}
			if err := marshal.Encode(ctx, w, r, http.StatusNotFound, res); err != nil {
				logFailedEncodingResponse(ctx, w, r, err, logger)
			}
			return
		}
		status := http.StatusNoContent
		w.WriteHeader(status)
		// A bit hacky way towrite override the existing context
		ctx = context.WithValue(ctx, constants.StatusCodeKey, status)
		*r = *(r.WithContext(ctx))
	})
}

func streamRecitalAudioHandler(logger logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			res := messageResponse{Message: "invalid id"}
			if err := marshal.Encode(r.Context(), w, r, http.StatusBadRequest, res); err != nil {
				writeErrHeader(w, err, logger)
			}
			return
		}

		filename := path.Join(services.BaseOutputPath, fmt.Sprintf("%d_out.wav", id))
		file, err := os.Open(filename)
		if err != nil {
			if os.IsNotExist(err) {
				http.NotFound(w, r)
			} else {
				http.Error(w, "open failed", http.StatusInternalServerError)
			}
			return
		}
		defer file.Close()

		stat, err := file.Stat()
		if err != nil {
			http.Error(w, "stat failed", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "audio/wav")
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
	})
}

func repondWithErrorMessage(w http.ResponseWriter, r *http.Request, message string, code int, logger logging.Logger) {
	res := messageResponse{
		Message: message,
	}
	if err := marshal.Encode(r.Context(), w, r, code, res); err != nil {
		writeErrHeader(w, err, logger)
	}
}

func respondWithOpaqueMessage(ctx context.Context, w http.ResponseWriter, r *http.Request, err error, logger logging.Logger) {
	logger.Err.Println(err)
	repondWithErrorMessage(w, r, "Something went wrong.", int(http.StatusInternalServerError), logger)
}

func writeErrHeader(w http.ResponseWriter, err error, logger logging.Logger) {
	w.WriteHeader(http.StatusInternalServerError)
	if _, err := w.Write([]byte(err.Error())); err != nil {
		logger.Err.Println("failed writing error")
	}
}

func logFailedEncodingResponse(ctx context.Context, w http.ResponseWriter, r *http.Request, err error, logger logging.Logger) {
	logger.Err.Println(err)
	res := messageResponse{"failed to encode response"}
	if err := marshal.Encode(ctx, w, r, http.StatusInternalServerError, res); err != nil {
		writeErrHeader(w, err, logger)
	}
}

func readIntQueryParam(name string, otherwise int, w http.ResponseWriter, r *http.Request, logger logging.Logger) (int, error) {
	valueString := r.URL.Query().Get(name)
	if valueString == "" {
		return otherwise, nil
	} else {
		value, err := strconv.Atoi(valueString)
		if err != nil {
			repondWithErrorMessage(w, r, fmt.Sprintf("Invalid %s: expected integer", name), http.StatusBadRequest, logger)
			return 0, err
		}
		return value, nil
	}
}

func swaggerHandler(logger logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		swaggerPath := "openapi.json"
		swaggerData, err := os.ReadFile(swaggerPath)
		if err != nil {
			logger.Err.Printf("Failed to read swagger file: %v", err)
			repondWithErrorMessage(w, r, "Failed to read OpenAPI specification", http.StatusInternalServerError, logger)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(swaggerData); err != nil {
			logger.Err.Printf("Failed to write swagger response: %v", err)
		}
	})
}
