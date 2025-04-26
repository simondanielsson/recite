package main

import (
	"context"
	"log"
	"time"

	"github.com/simondanielsson/recite/pkg/completions"
	"github.com/simondanielsson/recite/pkg/env"
	"github.com/simondanielsson/recite/pkg/player"
	"github.com/simondanielsson/recite/pkg/prompts"
)

func main() {
	if err := env.Load(); err != nil {
		log.Fatal("failed loading .env")
	}
	completion := completions.NewOpenAICompletion()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// TODO: load article from url and parse it and clean it.
	// TODO: check that article is not too long for input context limit for speech model

	article := `A New Bonsai API
Bonsai is the frontend web framework we use to build the vast majority of web apps at Jane Street. Here is some example code:

### 
type phase1_value
type phase1_witness
type phase2_value

val only_callable_in_phase1 : phase1_witness @ local -> phase1_value

val run_phase1 : (phase1_witness @ local -> 'a) -> 'a
val run_phase2 : phase1_value -> phase2_value
###

This leverages the compiler’s escape analysis for local values, with the phase1_witness serving as proof we are in the correct phase.
	`

	augmentPrompts, err := prompts.NewAugmentArticlePrompts(article)
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
	stream, err := completion.NewSpeech(ctx, refinedContent, recitePrompts.Developer)
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	if err := player.Play(stream); err != nil {
		log.Fatal(err)
	}
}
