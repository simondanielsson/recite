package services

import (
	"context"
	"fmt"
	"time"

	"github.com/openai/openai-go"
	"github.com/simondanielsson/recite/pkg/article"
	"github.com/simondanielsson/recite/pkg/audio"
	"github.com/simondanielsson/recite/pkg/completions"
	"github.com/simondanielsson/recite/pkg/prompts"
)

func CreateRecital(ctx context.Context, url string) error {
	completion := completions.NewOpenAICompletion()
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
	articleContent = articleContent[:min(3500, len(articleContent))]

	augmentPrompts, err := prompts.NewAugmentArticlePrompts(articleContent)
	if err != nil {
		return fmt.Errorf("failed loading augmentation prompts: %w", err)
	}
	refinedContent, err := completion.NewText(ctx, augmentPrompts.Developer, augmentPrompts.User)
	if err != nil {
		return fmt.Errorf("failed loading refinement prompts: %w", err)
	}
	recitePrompts, err := prompts.NewReciteArticlePrompt()
	if err != nil {
		return fmt.Errorf("failed loading recital prompts: %w", err)
	}
	// TODO: run in background?
	stream, err := completion.NewSpeech(ctx, refinedContent, recitePrompts.Developer, openai.AudioSpeechNewParamsResponseFormatWAV)
	if err != nil {
		return fmt.Errorf("failed creating recital: %w", err)
	}
	defer stream.Close()

	outPath := "2out.wav"
	if err := audio.Persist(stream, outPath); err != nil {
		return fmt.Errorf("failed persisting audio: %w", err)
	}
	return nil
}
