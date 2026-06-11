package ansible // Define el paquete 'ansible'.

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"pipeline-framework/pkg/config"
	"pipeline-framework/pkg/pipeline"
)

// AnsibleStep implements pipeline.Step
// AnsibleStep implementa la interfaz pipeline.Step para ejecutar playbooks de Ansible.
type AnsibleStep struct {
	cfg config.StepConfig // Configuración específica del paso.
}

// NewStep creates a new Ansible step
// NewStep crea una nueva instancia de un paso de Ansible.
func NewStep(cfg config.StepConfig) (pipeline.Step, error) {
	return &AnsibleStep{cfg: cfg}, nil // Retorna un puntero a AnsibleStep inicializado.
}

func (s *AnsibleStep) Name() string {
	return s.cfg.Name // Retorna el nombre del paso desde la configuración.
}

func (s *AnsibleStep) Type() string {
	return s.cfg.Type // Retorna el tipo de paso desde la configuración.
}

func (s *AnsibleStep) Execute(ctx context.Context) (*pipeline.StepResult, error) {
	start := time.Now() // Registra el tiempo de inicio de la ejecución.

	// Parse config
	playbook, _ := s.cfg.Config["playbook"].(string)   // Obtiene la ruta del playbook desde el mapa de configuración.
	inventory, _ := s.cfg.Config["inventory"].(string) // Obtiene la ruta del inventario (opcional).

	if playbook == "" { // Verifica si el playbook fue especificado.
		err := fmt.Errorf("playbook is required")                     // Crea un error si falta el playbook.
		return s.createResult(start, pipeline.StatusFailed, err), err // Retorna resultado fallido.
	}

	cmdArgs := []string{"ansible-playbook", playbook} // Inicializa los argumentos del comando con el ejecutable y el playbook.
	if inventory != "" {                              // Si se especificó un inventario.
		cmdArgs = append(cmdArgs, "-i", inventory) // Agrega el flag -i y la ruta del inventario.
	}

	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...) // Crea el comando exec con contexto para permitir cancelación.
	cmd.Dir = s.cfg.WorkingDir                                  // Establece el directorio de trabajo del comando.
	cmd.Stdout = os.Stdout                                      // Redirige la salida estándar al stdout del proceso padre.
	cmd.Stderr = os.Stderr                                      // Redirige la salida de error al stderr del proceso padre.

	err := cmd.Run()                 // Ejecuta el comando y espera a que termine.
	status := pipeline.StatusSuccess // Estado inicial éxito.
	if err != nil {                  // Si hubo error en la ejecución.
		status = pipeline.StatusFailed // Cambia el estado a fallido.
	}

	return s.createResult(start, status, err), err // Retorna el resultado encapsulado y el error.
}

func (s *AnsibleStep) createResult(start time.Time, status pipeline.StepStatus, err error) *pipeline.StepResult {
	return &pipeline.StepResult{ // Crea y retorna un objeto StepResult.
		Name:      s.cfg.Name,        // Nombre del paso.
		Status:    status,            // Estado final.
		Error:     err,               // Error capturado.
		Duration:  time.Since(start), // Duración de la ejecución.
		Timestamp: start,             // Marca de tiempo de inicio.
	}
}

func (s *AnsibleStep) Rollback(ctx context.Context) error {
	return nil // Rollback no implementado para Ansible, retorna nil.
}
