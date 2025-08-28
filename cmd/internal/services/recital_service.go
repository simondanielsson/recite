package services

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/openai/openai-go"
	"github.com/simondanielsson/recite/cmd/internal/db"
	"github.com/simondanielsson/recite/cmd/internal/dto"
	"github.com/simondanielsson/recite/cmd/internal/logging"
	"github.com/simondanielsson/recite/cmd/internal/queries"
	"github.com/simondanielsson/recite/pkg/article"
	"github.com/simondanielsson/recite/pkg/audio"
	"github.com/simondanielsson/recite/pkg/completions"
	"github.com/simondanielsson/recite/pkg/prompts"
)

func CreateRecital(ctx context.Context, url string, repository *queries.Queries, pool *pgxpool.Pool, logger logging.Logger) (int, error) {
	recital, err := repository.CreateRecital(ctx, queries.CreateRecitalParams{
		Url:         url,
		Title:       "TODO",
		Description: "TODO",
		Status:      "generating",
		Path:        "",
		CreatedAt:   pgtype.Date{Time: time.Now(), InfinityModifier: pgtype.Finite, Valid: true},
	})
	if err != nil {
		return 0, fmt.Errorf("failed creating recital: %w", err)
	}
	go func() {
		if err := generateRecital(recital.ID, url, pool, logger); err != nil {
			logger.Err.Println(err)
			statusFailedArgs := queries.UpdateRecitalStatusParams{ID: recital.ID, Status: string(Failed)}
			if err := repository.UpdateRecitalStatus(ctx, statusFailedArgs); err != nil {
				logger.Err.Println(err)
			}
		}
	}()

	return int(recital.ID), nil
}

func generateRecital(id int32, url string, pool *pgxpool.Pool, logger logging.Logger) error {
	completionClient := completions.NewOpenAICompletion()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	repository, commit, rollback, err := db.NewRepositoryWithTx(ctx, pool)
	if err != nil {
		return err
	}
	defer func() {
		// Rollback if an error occurred
		if err != nil {
			if err := rollback(ctx); err != nil {
				logger.Err.Println(err)
			}
		}
	}()

	articleReader := article.NewReader(3 * time.Second)
	articleContent, err := articleReader.Read(url)
	if err != nil {
		return fmt.Errorf("failed reading article: %w", err)
	}
	if articleContent == "" {
		return fmt.Errorf("article empty")
	}
	articleContent = articleContent[:min(len(articleContent), maxArticleContentCharLength)]

	augmentPrompts, err := prompts.NewAugmentArticlePrompts(articleContent)
	if err != nil {
		return fmt.Errorf("failed loading augmentation prompts: %w", err)
	}
	logger.Out.Println("Refining content")
	refinedContent, err := completionClient.NewText(ctx, augmentPrompts.Developer, augmentPrompts.User)
	if err != nil {
		return fmt.Errorf("failed loading refinement prompts: %w", err)
	}
	recitePrompts, err := prompts.NewReciteArticlePrompt()
	if err != nil {
		return fmt.Errorf("failed loading recital prompts: %w", err)
	}
	// TODO: run in background and store in blob?
	logger.Out.Println("Starting streaming")
	stream, err := completionClient.NewSpeech(ctx, refinedContent, recitePrompts.Developer, openai.AudioSpeechNewParamsResponseFormatWAV)
	if err != nil {
		return fmt.Errorf("failed creating recital: %w", err)
	}
	defer stream.Close()

	logger.Out.Println("Persisting stream")
	outPath := path.Join(BaseOutputPath, fmt.Sprintf("%d_out.wav", id))
	if err := audio.Persist(stream, outPath); err != nil {
		return fmt.Errorf("failed persisting audio: %w", err)
	}
	logger.Out.Println("Finished persisting stream")

	if err := repository.UpdateRecitalPath(ctx, queries.UpdateRecitalPathParams{ID: id, Path: outPath}); err != nil {
		return err
	}
	if err := repository.UpdateRecitalStatus(ctx, queries.UpdateRecitalStatusParams{ID: id, Status: string(Completed)}); err != nil {
		return err
	}

	if err := commit(ctx); err != nil {
		return err
	}
	logger.Out.Println("Successfully generated and persisted recital.")

	return nil
}

func ListRecitals(ctx context.Context, limit int32, offset int32, repository *queries.Queries, logger logging.Logger) ([]dto.Recital, error) {
	dbRecitals, err := repository.ListRecitals(ctx, queries.ListRecitalsParams{Limit: limit, Offset: offset})
	if err != nil {
		return []dto.Recital{}, err
	}

	var recitals []dto.Recital
	for _, recital := range dbRecitals {
		recitals = append(recitals, dtoFromQuery(recital))
	}
	return recitals, nil
}

func GetRecital(ctx context.Context, id int32, repository *queries.Queries, logger logging.Logger) (dto.Recital, error) {
	recital, err := repository.GetRecital(ctx, id)
	if err != nil {
		return dto.Recital{}, err
	}

	return dtoFromQuery(recital), nil
}

func DeleteRecital(ctx context.Context, id int32, repository *queries.Queries, logger logging.Logger) error {
	return repository.DeleteRecital(ctx, id)
}

func dtoFromQuery(recital queries.Recital) dto.Recital {
	return dto.Recital{
		ID:          recital.ID,
		Title:       recital.Title,
		Description: recital.Description,
		Status:      recital.Status,
		CreatedAt:   recital.CreatedAt.Time,
	}
}
