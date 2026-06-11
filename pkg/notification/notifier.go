package notification // Define el paquete 'notification'.

import (
	"context"
	"fmt"
	"time"

	"pipeline-framework/pkg/config"
	"pipeline-framework/pkg/pipeline"
)

// NotificationStep implements pipeline.Step
// NotificationStep implementa la interfaz pipeline.Step para enviar notificaciones simples.
type NotificationStep struct {
	cfg config.StepConfig // Configuración específica para el paso.
}

// NewStep creates a new Notification step
// NewStep crea una nueva instancia de NotificationStep.
func NewStep(cfg config.StepConfig) (pipeline.Step, error) {
	return &NotificationStep{cfg: cfg}, nil // Retorna un puntero a NotificationStep inicializado.
}

func (s *NotificationStep) Name() string {
	return s.cfg.Name // Retorna el nombre del paso desde la configuración.
}

func (s *NotificationStep) Type() string {
	return s.cfg.Type // Retorna el tipo de paso desde la configuración.
}

func (s *NotificationStep) Execute(ctx context.Context) (*pipeline.StepResult, error) {
	start := time.Now() // Registra el tiempo de inicio de la ejecución.

	msg, _ := s.cfg.Config["message"].(string)   // Obtiene el mensaje a notificar desde la configuración.
	prefix, _ := s.cfg.Config["prefix"].(string) // Obtiene un prefijo opcional para el mensaje.

	fmt.Printf("[%s] %s\n", prefix, msg) // Imprime la notificación en la salida estándar (simbolizando el envío).

	return &pipeline.StepResult{ // Retorna el resultado exitoso.
		Name:      s.cfg.Name,             // Nombre del paso.
		Status:    pipeline.StatusSuccess, // Estado exitoso (siempre tiene éxito en esta implementación simple).
		Duration:  time.Since(start),      // Duración.
		Timestamp: start,                  // Timestamp.
	}, nil // Retorna nil como error.
}

func (s *NotificationStep) Rollback(ctx context.Context) error {
	return nil // Rollback no implementado para notificaciones (usualmente irreversible o innecesario), retorna nil.
}
