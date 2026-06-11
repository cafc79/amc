package policy

import (
	"context"
	"fmt"
	"time"

	"pipeline-framework/pkg/config"
	"pipeline-framework/pkg/pipeline"
)

// PolicyStep implements pipeline.Step
type PolicyStep struct {
	cfg config.StepConfig
}

// NewStep creates a new Policy step
func NewStep(cfg config.StepConfig) (pipeline.Step, error) {
	return &PolicyStep{cfg: cfg}, nil
}

func (s *PolicyStep) Name() string {
	return s.cfg.Name
}

func (s *PolicyStep) Type() string {
	return s.cfg.Type
}

func (s *PolicyStep) Execute(ctx context.Context) (*pipeline.StepResult, error) {
	start := time.Now()

	// Mock implementation
	fmt.Printf("Validating policy for step %s\n", s.cfg.Name)

	// In a real implementation, we would run 'opa eval' or similar

	return &pipeline.StepResult{
			Name:      s.cfg.Name,
			Status:    pipeline.StatusSuccess,
			Duration:  time.Since(start),
			Timestamp: start,
		},
		nil
}

func (s *PolicyStep) Rollback(ctx context.Context) error {
	return nil
}
