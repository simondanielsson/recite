package main

import (
	"context"
	"log"
	"time"

	"github.com/simondanielsson/recite/pkg/completions"
	"github.com/simondanielsson/recite/pkg/env"
	"github.com/simondanielsson/recite/pkg/player"
)

func main() {
	if err := env.Load(); err != nil {
		log.Fatal("failed loading .env")
	}
	completion := completions.NewOpenAICompletion()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	articleContent := `A New Bonsai API
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

	devMessage := `You are an programming article narrator. Your task is to substitute any code blocks in the original content of an article by inserting an intuitive narration of the code to make it understandable without reading it (i.e. if one would listen to your narration). 

	In your narration of the code, refer to specific variable names or functions if this helps with the understanding of the code.

	Respond with the entire augmented article containing narrated code rather than the literal code, and nothing else.
	`
	userMessage := `This is the original content of the article: ` + articleContent + `.\nThis is the augmented article:`
	refinedContent, err := completion.NewText(ctx, devMessage, userMessage)
	if err != nil {
		log.Fatal(err)
	}

	instructions := "Read the article as if you were an audio book narrator."
	stream, err := completion.NewSpeech(ctx, refinedContent, instructions)
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	if err := player.Play(stream); err != nil {
		log.Fatal(err)
	}
}
