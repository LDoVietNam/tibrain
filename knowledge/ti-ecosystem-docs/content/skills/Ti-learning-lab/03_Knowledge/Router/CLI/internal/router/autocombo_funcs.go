// Package router - convenience wrappers around autocombo selection.
package router

import "github.com/ti/cli/internal/autocombo"

// AutoComboSelect picks best provider+model for the given task type using
// the default pool and weights.
func AutoComboSelect(taskType string) autocombo.SelectionResult {
	return autocombo.Select(autocombo.AutoComboConfig{
		ID:              "ti-default",
		Name:            "Ti Default",
		Weights:         autocombo.DefaultWeights,
		ExplorationRate: 0.05,
	}, AutoComboPool(), taskType)
}

// SelectAutoModel returns provider and model strings selected for the task.
// The second parameter is reserved for future use (e.g., custom pool override).
func SelectAutoModel(taskType string, _ any) (provider, model string) {
	res := AutoComboSelect(taskType)
	return res.Provider, res.Model
}
