package prompts

import (
	"fmt"

	"github.com/noirbizarre/gonja"
)

// hydratePrompt hydrates a prompt template with values
func hydratePrompt(prompt string, values map[string]string) (string, error) {
	template, err := gonja.FromString(prompt)
	if err != nil {
		return "", fmt.Errorf("failed loading prompt as jinja template: %w", err)
	}
	gonjaCtx := gonja.Context{}
	for key, val := range values {
		gonjaCtx[key] = val
	}
	hydratedPrompt, err := template.Execute(gonjaCtx)
	if err != nil {
		return "", fmt.Errorf("failed hydrating prompt template with values: %w", err)
	}
	return hydratedPrompt, nil
}
