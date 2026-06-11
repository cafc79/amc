// pkg/terraform/client.go
package terraform // Define el paquete 'terraform'.

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"pipeline-framework/pkg/config"
	"pipeline-framework/pkg/pipeline"
)

// TerraformStep implements pipeline.Step
// TerraformStep implementa la interfaz pipeline.Step para ejecutar pasos de Terraform.
type TerraformStep struct {
	cfg config.StepConfig // Configuración específica del paso.
}

// NewStep creates a new Terraform step
// NewStep crea una nueva instancia de un paso de Terraform.
func NewStep(cfg config.StepConfig) (pipeline.Step, error) {
	return &TerraformStep{cfg: cfg}, nil // Retorna un puntero a TerraformStep inicializado con la configuración.
}

func (s *TerraformStep) Name() string {
	return s.cfg.Name // Retorna el nombre del paso desde la configuración.
}

func (s *TerraformStep) Type() string {
	return s.cfg.Type // Retorna el tipo de paso desde la configuración.
}

func (s *TerraformStep) Execute(ctx context.Context) (*pipeline.StepResult, error) {
	start := time.Now() // Registra el tiempo de inicio de la ejecución.

	client, err := NewClient(s.cfg.WorkingDir) // Crea un nuevo cliente de Terraform con el directorio de trabajo especificado.
	if err != nil {                            // Verifica si hubo error al crear el cliente.
		return &pipeline.StepResult{ // Retorna un resultado de fallo si no se pudo crear el cliente.
			Name:      s.cfg.Name,            // Nombre del paso.
			Status:    pipeline.StatusFailed, // Estado fallido.
			Error:     err,                   // Error capturado.
			Duration:  time.Since(start),     // Duración hasta el error.
			Timestamp: start,                 // Marca de tiempo.
		}, err // Retorna el error.
	}

	var execErr error      // Variable para capturar errores de ejecución de comandos.
	action := s.cfg.Action // Obtiene la acción de Terraform a ejecutar (init, plan, apply, destroy).

	switch action { // Selecciona la acción a ejecutar.
	case "init":
		execErr = client.Init(ctx) // Ejecuta 'terraform init'.
	case "plan":
		outFile := "tfplan"                                 // Nombre por defecto para el archivo de salida del plan.
		if val, ok := s.cfg.Config["output"].(string); ok { // Verifica si se especificó un nombre de salida en la config.
			outFile = val // Usa el nombre especificado.
		}
		execErr = client.Plan(ctx, outFile) // Ejecuta 'terraform plan'.
	case "apply":
		planFile := ""                                         // Nombre del archivo de plan a aplicar (opcional).
		if val, ok := s.cfg.Config["plan_file"].(string); ok { // Verifica si se especificó un archivo de plan.
			planFile = val // Usa el archivo especificado.
		}
		execErr = client.Apply(ctx, planFile) // Ejecuta 'terraform apply'.
	case "destroy":
		execErr = client.Destroy(ctx) // Ejecuta 'terraform destroy'.
	default:
		execErr = fmt.Errorf("unknown terraform action: %s", action) // Retorna error si la acción no es reconocida.
	}

	status := pipeline.StatusSuccess // Estado por defecto es éxito.
	if execErr != nil {              // Si hubo un error en la ejecución.
		status = pipeline.StatusFailed // Cambia el estado a fallido.
	}

	return &pipeline.StepResult{ // Retorna el resultado final del paso.
		Name:      s.cfg.Name,        // Nombre del paso.
		Status:    status,            // Estado (éxito o fallo).
		Error:     execErr,           // Error si lo hubo.
		Duration:  time.Since(start), // Duración total.
		Timestamp: start,             // Marca de tiempo de inicio.
	}, execErr // Retorna el error si lo hubo.
}

func (s *TerraformStep) Rollback(ctx context.Context) error {
	if s.cfg.Rollback == nil { // Verifica si hay configuración de rollback.
		return nil // Si no hay, no hace nada y retorna éxito.
	}
	// Basic rollback logic implementation
	if s.cfg.Rollback.Action == "destroy" { // Verifica si la acción de rollback es "destroy".
		client, err := NewClient(s.cfg.WorkingDir) // Crea un nuevo cliente de Terraform.
		if err != nil {                            // Si falla la creación del cliente.
			return err // Retorna el error.
		}
		return client.Destroy(ctx) // Ejecuta 'terraform destroy' para revertir.
	}
	return nil // Retorna éxito si no se ejecutó ninguna acción conocida.
}

// Client handles Terraform operations
// Client maneja las operaciones de Terraform.
type Client struct {
	workingDir string   // Directorio de trabajo donde se ejecutarán los comandos.
	env        []string // Variables de entorno para los comandos.
}

// NewClient crea un nuevo cliente de Terraform
// NewClient inicializa y retorna un nuevo cliente de Terraform.
func NewClient(workingDir string) (*Client, error) {
	absPath, err := filepath.Abs(workingDir) // Obtiene la ruta absoluta del directorio de trabajo.
	if err != nil {                          // Si hay error al obtener la ruta.
		return nil, err // Retorna el error.
	}

	// Verificar que el directorio existe
	if _, err := os.Stat(absPath); os.IsNotExist(err) { // Comprueba si el directorio existe.
		return nil, fmt.Errorf("terraform directory does not exist: %s", absPath) // Retorna error si no existe.
	}

	return &Client{ // Retorna el puntero al nuevo cliente.
		workingDir: absPath,      // Asigna la ruta absoluta.
		env:        os.Environ(), // Hereda las variables de entorno del proceso actual.
	}, nil // Retorna nil como error.
}

// Init inicializa el directorio de Terraform
func (c *Client) Init(ctx context.Context) error {
	return c.runCommand(ctx, "init", "-input=false") // Ejecuta 'init' sin esperar input del usuario.
}

// Plan genera un plan de Terraform
func (c *Client) Plan(ctx context.Context, outFile string) error {
	args := []string{"plan", "-input=false", "-out=" + outFile} // Construye argumentos: plan sin input y guarda en archivo.
	return c.runCommand(ctx, args...)                           // Ejecuta el comando con los argumentos.
}

// Apply aplica el plan de Terraform
func (c *Client) Apply(ctx context.Context, planFile string) error {
	args := []string{"apply", "-input=false", "-auto-approve"} // Argumentos base: apply sin input y auto-aprobación.
	if planFile != "" {                                        // Si se proporciona un archivo de plan.
		args = append(args, planFile) // Lo añade a los argumentos.
	}
	return c.runCommand(ctx, args...) // Ejecuta el comando.
}

// Destroy destruye la infraestructura
func (c *Client) Destroy(ctx context.Context) error {
	return c.runCommand(ctx, "destroy", "-input=false", "-auto-approve") // Ejecuta destroy sin input y con auto-aprobación.
}

// runCommand ejecuta un comando de Terraform
func (c *Client) runCommand(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "terraform", args...) // Prepara el comando terraform con contexto y argumentos.
	cmd.Dir = c.workingDir                                // Establece el directorio de trabajo del comando.
	cmd.Env = c.env                                       // Establece las variables de entorno.

	var stdout, stderr bytes.Buffer // Buffers para capturar salida estándar y de error.
	cmd.Stdout = &stdout            // Asigna buffer de stdout.
	cmd.Stderr = &stderr            // Asigna buffer de stderr.

	if err := cmd.Run(); err != nil { // Ejecuta el comando y espera a que termine. Verifica error.
		return fmt.Errorf("terraform command failed: %w\nstdout: %s\nstderr: %s", // Formatea mensaje de error con detalles.
			err, stdout.String(), stderr.String())
	}

	return nil // Retorna éxito si el comando finalizó correctamente.
}
