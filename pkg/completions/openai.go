package completions

import (
	"context"
	"io"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/param"
)

// OpenAICompletion is a Completion using OpenAI
type OpenAICompletion struct {
	client openai.Client
}

// NewOpenAICompletion creates a new OpenAI Completion
func NewOpenAICompletion() OpenAICompletion {
	return OpenAICompletion{
		client: openai.NewClient(),
	}
}

// NewText generates text based on instructions
func (c OpenAICompletion) NewText(ctx context.Context, devMessage, userMessage string) (string, error) {
	params := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.DeveloperMessage(devMessage),
			openai.UserMessage(userMessage),
		},
		Seed:  openai.Int(0),
		Model: openai.ChatModelGPT4o,
	}
	completion, err := c.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", err
	}
	return completion.Choices[0].Message.Content, nil
}

// NewSpeech generates speech from content based on instructions
func (c OpenAICompletion) NewSpeech(ctx context.Context, content, instructions string) (io.ReadCloser, error) {
	res, err := c.client.Audio.Speech.New(ctx, openai.AudioSpeechNewParams{
		Input: content,
		Instructions: param.Opt[string]{
			Value: instructions,
		},
		Model:          openai.SpeechModelGPT4oMiniTTS,
		Voice:          openai.AudioSpeechNewParamsVoiceAlloy,
		ResponseFormat: openai.AudioSpeechNewParamsResponseFormatPCM,
	})
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}
