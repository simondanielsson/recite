package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/openai/openai-go"
	"github.com/simondanielsson/recite/pkg/article"
	"github.com/simondanielsson/recite/pkg/completions"
	"github.com/simondanielsson/recite/pkg/env"
	"github.com/simondanielsson/recite/pkg/player"
	"github.com/simondanielsson/recite/pkg/prompts"
)

func main() {
	// TODO: use a proper CLI library
	if len(os.Args) < 2 {
		log.Fatalf("A URL must be provided")
	}
	articleUrl := os.Args[1]
	if err := env.Load(); err != nil {
		log.Fatal("failed loading .env")
	}
	completion := completions.NewOpenAICompletion()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	articleReader := article.NewReader(3 * time.Second)
	articleContent, err := articleReader.Read(articleUrl)
	if err != nil {
		log.Fatalf("Failed reading article: %v", err)
	}
	articleContent = articleContent[:3500]

	augmentPrompts, err := prompts.NewAugmentArticlePrompts(articleContent)
	if err != nil {
		log.Fatal(err)
	}
	refinedContent, err := completion.NewText(ctx, augmentPrompts.Developer, augmentPrompts.User)
	if err != nil {
		log.Fatal(err)
	}
	recitePrompts, err := prompts.NewReciteArticlePrompt()
	if err != nil {
		log.Fatal(err)
	}
	stream, err := completion.NewSpeech(ctx, refinedContent, recitePrompts.Developer, openai.AudioSpeechNewParamsResponseFormatPCM)
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	if err := player.Play(stream); err != nil {
		log.Fatal(err)
	}
}
