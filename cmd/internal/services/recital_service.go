package services

import (
	"context"
	"fmt"
	"log"
	"path"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/openai/openai-go"
	"github.com/simondanielsson/recite/cmd/internal/queries"
	"github.com/simondanielsson/recite/pkg/article"
	"github.com/simondanielsson/recite/pkg/audio"
	"github.com/simondanielsson/recite/pkg/completions"
	"github.com/simondanielsson/recite/pkg/prompts"
)

func CreateRecital(ctx context.Context, url string, repository *queries.Queries, logger *log.Logger) (int, error) {
	recital, err := repository.CreateRecital(ctx, queries.CreateRecitalParams{
		Url:         url,
		Title:       "TODO",
		Description: "TODO",
		Status:      "generating",
		Path:        "",
		CreatedAt:   pgtype.Date{Time: time.Now(), InfinityModifier: pgtype.Infinity, Valid: true},
	})
	if err != nil {
		return 0, fmt.Errorf("failed creating recital: %w", err)
	}
	go generateRecital(ctx, recital.ID, url, repository, logger)

	return int(recital.ID), nil
}

func generateRecital(ctx context.Context, id int32, url string, repository *queries.Queries, logger *log.Logger) error {
	completionClient := completions.NewOpenAICompletion()
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

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
	refinedContent, err := completionClient.NewText(ctx, augmentPrompts.Developer, augmentPrompts.User)
	if err != nil {
		return fmt.Errorf("failed loading refinement prompts: %w", err)
	}
	recitePrompts, err := prompts.NewReciteArticlePrompt()
	if err != nil {
		return fmt.Errorf("failed loading recital prompts: %w", err)
	}
	// TODO: run in background and store in blob?
	stream, err := completionClient.NewSpeech(ctx, refinedContent, recitePrompts.Developer, openai.AudioSpeechNewParamsResponseFormatWAV)
	if err != nil {
		return fmt.Errorf("failed creating recital: %w", err)
	}
	defer stream.Close()

	outPath := path.Join(baseOutputPath, "out.wav")
	if err := audio.Persist(stream, outPath); err != nil {
		return fmt.Errorf("failed persisting audio: %w", err)
	}

	if err := repository.UpdateRecitalPath(ctx, queries.UpdateRecitalPathParams{ID: id, Path: outPath}); err != nil {
		return err
	}
	if err := repository.UpdateRecitalStatus(ctx, queries.UpdateRecitalStatusParams{ID: id, Status: "completed"}); err != nil {
		return err
	}
	logger.Println("Successfully generated and persisted recital.")
	return nil
}
