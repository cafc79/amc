package test // Define el paquete 'test'.

import (
	"context"
	"fmt"
	"time"

	"pipeline-framework/pkg/config"
	"pipeline-framework/pkg/pipeline"
)

// TestStep implements pipeline.Step
// TestStep implementa la interfaz pipeline.Step para pruebas de infraestructura.
type TestStep struct {
	cfg config.StepConfig // Configuración específica para el paso.
}

// NewStep creates a new Test step
// NewStep crea una nueva instancia de TestStep.
func NewStep(cfg config.StepConfig) (pipeline.Step, error) {
	return &TestStep{cfg: cfg}, nil // Retorna un puntero a TestStep inicializado.
}

func (s *TestStep) Name() string {
	return s.cfg.Name // Retorna el nombre del paso desde la configuración.
}

func (s *TestStep) Type() string {
	return s.cfg.Type // Retorna el tipo de paso desde la configuración.
}

func (s *TestStep) Execute(ctx context.Context) (*pipeline.StepResult, error) {
	start := time.Now() // Registra el tiempo de inicio de la ejecución.

	testType, _ := s.cfg.Config["type"].(string)              // Obtiene el tipo de prueba desde la configuración.
	fmt.Printf("Running infrastructure test: %s\n", testType) // Imprime el tipo de prueba que se está ejecutando.

	// Mock implementation
	// Implementación simulada (Mock)
	// Real implementation would check TCP connection or HTTP status
	// Una implementación real verificaría conexiones TCP o estados HTTP.

	return &pipeline.StepResult{ // Retorna el resultado de la prueba.
		Name:      s.cfg.Name,             // Nombre del paso.
		Status:    pipeline.StatusSuccess, // Estado exitoso por defecto en esta simulación.
		Duration:  time.Since(start),      // Duración.
		Timestamp: start,                  // Timestamp.
	}, nil // Retorna nil como error.
}

func (s *TestStep) Rollback(ctx context.Context) error {
	return nil // Rollback no implementado para pruebas, retorna nil.
}
