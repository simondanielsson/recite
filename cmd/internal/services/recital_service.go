package services

import (
	"context"
	"fmt"
	"log"
	"path"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/openai/openai-go"
	"github.com/simondanielsson/recite/cmd/internal/db"
	"github.com/simondanielsson/recite/cmd/internal/queries"
	"github.com/simondanielsson/recite/pkg/article"
	"github.com/simondanielsson/recite/pkg/audio"
	"github.com/simondanielsson/recite/pkg/completions"
	"github.com/simondanielsson/recite/pkg/prompts"
)

func CreateRecital(ctx context.Context, url string, repository *queries.Queries, pool *pgxpool.Pool, logger *log.Logger) (int, error) {
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
			logger.Println(err)
			repository.UpdateRecitalStatus(ctx, queries.UpdateRecitalStatusParams{ID: recital.ID, Status: "failed"})
		}
	}()

	return int(recital.ID), nil
}

func generateRecital(id int32, url string, pool *pgxpool.Pool, logger *log.Logger) error {
	completionClient := completions.NewOpenAICompletion()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	repository, _, err := db.NewRepositoryWithTx(ctx, pool)
	if err != nil {
		return err
	}

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
	logger.Println("Refining content")
	refinedContent, err := completionClient.NewText(ctx, augmentPrompts.Developer, augmentPrompts.User)
	if err != nil {
		return fmt.Errorf("failed loading refinement prompts: %w", err)
	}
	recitePrompts, err := prompts.NewReciteArticlePrompt()
	if err != nil {
		return fmt.Errorf("failed loading recital prompts: %w", err)
	}
	// TODO: run in background and store in blob?
	logger.Println("Starting streaming")
	stream, err := completionClient.NewSpeech(ctx, refinedContent, recitePrompts.Developer, openai.AudioSpeechNewParamsResponseFormatWAV)
	if err != nil {
		return fmt.Errorf("failed creating recital: %w", err)
	}
	defer stream.Close()

	logger.Println("Persisting stream")
	outPath := path.Join(baseOutputPath, fmt.Sprintf("%d_out.wav", id))
	if err := audio.Persist(stream, outPath); err != nil {
		return fmt.Errorf("failed persisting audio: %w", err)
	}
	logger.Println("Finished persisting stream")

	if err := repository.UpdateRecitalPath(ctx, queries.UpdateRecitalPathParams{ID: id, Path: outPath}); err != nil {
		return err
	}
	if err := repository.UpdateRecitalStatus(ctx, queries.UpdateRecitalStatusParams{ID: id, Status: string(Completed)}); err != nil {
		return err
	}
	logger.Println("Successfully generated and persisted recital.")
	return nil
}

func GetRecital(ctx context.Context, id int32, repository *queries.Queries, logger *log.Logger) (queries.Recital, error) {
	recital, err := repository.GetRecital(ctx, id)
	if err != nil {
		return queries.Recital{}, err
	}
	return recital, nil
}
