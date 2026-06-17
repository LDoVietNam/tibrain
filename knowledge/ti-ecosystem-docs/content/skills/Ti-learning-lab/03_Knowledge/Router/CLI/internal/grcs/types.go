package grcs

import (
	"context"
)

// Builder interface for training GRCS models
type Builder interface {
	AddSample(ctx context.Context, content string, label bool) error
	Build(ctx context.Context, clusters int) (*GRCS, error)
}

// Generator interface for creating candidate outputs
type Generator interface {
	Generate(ctx context.Context, prompt string, k int) ([]*Candidate, error)
}

// JudgeUI interface for human labeling
type JudgeUI interface {
	StartServer(port int) error
	LabelSample(ctx context.Context, sampleID string) (bool, error)
}
