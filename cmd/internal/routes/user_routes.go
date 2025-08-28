package routes

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	constants "github.com/simondanielsson/recite/cmd/internal"
	"github.com/simondanielsson/recite/cmd/internal/logging"
	"github.com/simondanielsson/recite/cmd/internal/queries"
	"github.com/simondanielsson/recite/cmd/internal/utils"
)

func createUserHandler(logger logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		email := r.FormValue("email")
		password := r.FormValue("password")

		ctx := r.Context()

		passwordHash, err := utils.HashPassword(password)
		if err != nil {
			respondWithOpaqueMessage(ctx, w, r, err, logger)
			return
		}

		repository, ok := ctx.Value(constants.RepositoryKey).(*queries.Queries)
		if !ok {
			respondWithOpaqueMessage(ctx, w, r, fmt.Errorf("failed loading repository"), logger)
			return
		}

		userArgs := queries.CreateUserParams{
			Email:        email,
			PasswordHash: passwordHash,
			CreatedAt:    pgtype.Date{Time: time.Now(), InfinityModifier: pgtype.Finite, Valid: true},
		}
		_, err = repository.CreateUser(ctx, userArgs)
		if err != nil {
			repondWithErrorMessage(w, r, "Failed creating user", http.StatusInternalServerError, logger)
			return
		}

		status := http.StatusOK
		w.WriteHeader(status)
		// A bit hacky way towrite override the existing context
		ctx = context.WithValue(ctx, constants.StatusCodeKey, status)
		*r = *(r.WithContext(ctx))
	})
}

func loginUserHandler(logger logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		repository, ok := ctx.Value(constants.RepositoryKey).(*queries.Queries)
		if !ok {
			respondWithOpaqueMessage(ctx, w, r, fmt.Errorf("failed loading repository"), logger)
			return
		}

		email, password, ok := r.BasicAuth()
		if !ok {
			repondWithErrorMessage(w, r, "could not get credentials", http.StatusBadRequest, logger)
			return
		}

		user, err := repository.GetUser(ctx, email)
		if err != nil {
			message := fmt.Sprintf("User with email %s not found", email)
			repondWithErrorMessage(w, r, message, http.StatusNotFound, logger)
			return
		}
		if err := utils.ValidatePassword(password, user.PasswordHash); err != nil {
			repondWithErrorMessage(w, r, "Incorrect password", http.StatusBadRequest, logger)
			return
		}

		status := http.StatusOK
		w.WriteHeader(status)
		// A bit hacky way towrite override the existing context
		ctx = context.WithValue(ctx, constants.StatusCodeKey, status)
		*r = *(r.WithContext(ctx))
	})
}
