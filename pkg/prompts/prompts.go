package prompts

import (
	"errors"
	"fmt"
	"os"
	"path"

	constants "github.com/simondanielsson/recite/pkg"
	yaml "gopkg.in/yaml.v3"
)

// Prompts contains developer and user prompts
type Prompts struct {
	Developer string `yaml:"developer"`
	User      string `yaml:"user"`
}

// NewAugmentArticlePrompts returns developer and user messages for augmenting an article with code narration.
func NewAugmentArticlePrompts(article string) (Prompts, error) {
	file, err := os.Open(augmentArticlePromptPath())
	if err != nil {
		return Prompts{}, fmt.Errorf("could not find prompt file: %w", err)
	}
	defer file.Close()

	var prompts Prompts

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&prompts); err != nil {
		return Prompts{}, fmt.Errorf("failed decoding prompt file into Prompts struct: %w", err)
	}
	if prompts.Developer == "" || prompts.User == "" {
		return Prompts{}, errors.New("either developer or user message is empty")
	}

	prompts.User, err = hydratePrompt(prompts.User, map[string]string{"article": article})
	if err != nil {
		return Prompts{}, err
	}
	if prompts.User == "" {
		return Prompts{}, errors.New("user prompt empty after hydration")
	}
	return prompts, nil
}

// NewReciteArticlePrompt returns a prompt for reciting articles.
func NewReciteArticlePrompt() (Prompts, error) {
	file, err := os.Open(reciteArticlePromptPath())
	if err != nil {
		return Prompts{}, fmt.Errorf("could not find prompt file: %w", err)
	}
	defer file.Close()

	var prompts Prompts

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&prompts); err != nil {
		return Prompts{}, fmt.Errorf("failed decoding prompt file into Prompts struct: %w", err)
	}

	return prompts, nil
}

// getPromptPath returns the full path to a prompt file
func getPromptPath(filename string) string {
	return path.Join(constants.ConfigsDir, constants.PromptsSubDir, filename)
}

// augmentArticlePromptPath returns the full path to the prompt for augmenting articles
func augmentArticlePromptPath() string {
	return getPromptPath(constants.AugmentArticlePromptFileName)
}

// reicteArticlePromptPath returns the full path to the prompt for reciting articles
func reciteArticlePromptPath() string {
	return getPromptPath(constants.ReciteArticlePromptFileName)
}
