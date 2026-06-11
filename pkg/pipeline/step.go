package pipeline

import (
	"context"
	"time"

	"pipeline-framework/pkg/config"
)

// StepStatus represents the status of a step
type StepStatus string

const (
	StatusPending StepStatus = "pending"
	StatusRunning StepStatus = "running"
	StatusSuccess StepStatus = "success"
	StatusFailed  StepStatus = "failed"
	StatusSkipped StepStatus = "skipped"
)

// StepResult stores the result of a step
type StepResult struct {
	Name      string
	Status    StepStatus
	Error     error
	Duration  time.Duration
	Timestamp time.Time
	Output    map[string]interface{}
}

// Step represents a single step in the pipeline
type Step interface {
	// Name returns the name of the step
	Name() string
	// Type returns the type of the step
	Type() string
	// Execute runs the step
	Execute(ctx context.Context) (*StepResult, error)
	// Rollback performs rollback actions if necessary
	Rollback(ctx context.Context) error
}

// StepFactory is a function that creates a new Step
type StepFactory func(cfg config.StepConfig) (Step, error)
