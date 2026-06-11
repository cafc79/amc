// pkg/gcp/client.go
package gcp // Define el paquete 'gcp'.

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/compute/metadata"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"

	"pipeline-framework/pkg/config"
	"pipeline-framework/pkg/pipeline"
)

// GCPStep implements pipeline.Step
// GCPStep implementa la interfaz pipeline.Step para interacciones con Google Cloud.
type GCPStep struct {
	cfg config.StepConfig // Configuración específica para el paso.
}

// NewStep creates a new GCP step
// NewStep crea una nueva instancia de GCPStep.
func NewStep(cfg config.StepConfig) (pipeline.Step, error) {
	return &GCPStep{cfg: cfg}, nil // Retorna la instancia inicializada con la configuración.
}

func (s *GCPStep) Name() string {
	return s.cfg.Name // Retorna el nombre del paso desde la configuración.
}

func (s *GCPStep) Type() string {
	return s.cfg.Type // Retorna el tipo de paso configurado.
}

func (s *GCPStep) Execute(ctx context.Context) (*pipeline.StepResult, error) {
	start := time.Now() // Registra el tiempo de inicio.

	projectID := ""                                         // Inicializa variable para el ID del proyecto.
	if val, ok := s.cfg.Config["project_id"].(string); ok { // Intenta obtener el project_id de la configuración.
		projectID = val // Si existe, lo asigna.
	}

	client, err := NewClient(ctx, projectID) // Crea un nuevo cliente de GCP.
	if err != nil {                          // Verifica errores en la creación del cliente.
		return s.createResult(start, pipeline.StatusFailed, err), err // Retorna fallo si hay error.
	}

	var execErr error               // Variable para el error de ejecución.
	if s.cfg.Type == "gcp-verify" { // Verifica si el tipo de paso es 'gcp-verify'.
		instanceName := ""                                         // Variable para el nombre de la instancia.
		if val, ok := s.cfg.Config["instance_name"].(string); ok { // Obtiene el nombre de instancia de la config.
			instanceName = val // Asigna el valor.
		}
		execErr = client.VerifyComputeInstances(ctx, instanceName, 1) // Ejecuta verificación de instancias.
	} else if s.cfg.Type == "wait" { // Verifica si el tipo es 'wait' (lógica pendiente).
		// Wait logic should probably be in a utility or separate step,
		// but if it relies on GCP resource status, it fits here.
		// For generic wait, main.go defines "wait" type, we might handle it separately
		// or reuse this if mapped. But main.go maps "gcp-verification" to this.
		// We'll focus on gcp-verify.
	}

	status := pipeline.StatusSuccess // Estado inicial éxito.
	if execErr != nil {              // Si hubo error en ejecución.
		status = pipeline.StatusFailed // Cambia estado a fallido.
	}

	return s.createResult(start, status, execErr), execErr // Retorna resultado.
}

func (s *GCPStep) createResult(start time.Time, status pipeline.StepStatus, err error) *pipeline.StepResult {
	return &pipeline.StepResult{ // Crea objeto StepResult.
		Name:      s.cfg.Name,        // Nombre.
		Status:    status,            // Estado.
		Error:     err,               // Error.
		Duration:  time.Since(start), // Duración.
		Timestamp: start,             // Timestamp.
	}
}

func (s *GCPStep) Rollback(ctx context.Context) error {
	return nil // Rollback no implementado, retorna nil.
}

type Client struct {
	projectID string           // ID del proyecto GCP.
	compute   *compute.Service // Servicio de cliente de Google Compute.
}

func NewClient(ctx context.Context, projectID string) (*Client, error) {
	// Si no se proporciona projectID, intentar obtenerlo de metadata
	if projectID == "" { // Verifica si el projectID está vacío.
		var err error                         // Variable de error.
		projectID, err = metadata.ProjectID() // Intenta obtener del servicio de metadatos.
		if err != nil {                       // Si falla metadatos.
			return nil, fmt.Errorf("no se pudo obtener project_id de metadata: %w", err) // Retorna error.
		}
	}

	computeService, err := compute.NewService(ctx, option.WithScopes(compute.CloudPlatformScope)) // Crea servicio Compute.
	if err != nil {                                                                               // Si falla.
		return nil, fmt.Errorf("error creando cliente Compute: %w", err) // Retorna error.
	}

	return &Client{ // Retorna cliente.
		projectID: projectID,      // Asigna proyecto.
		compute:   computeService, // Asigna servicio compute.
	}, nil
}

// VerifyComputeInstances verifica que existan instancias con ciertos criterios
// VerifyComputeInstances comprueba la existencia de un número mínimo de instancias coincidiendo con un patrón.
func (c *Client) VerifyComputeInstances(ctx context.Context, namePattern string, minCount int) error {
	instances, err := c.compute.Instances.AggregatedList(c.projectID).Context(ctx).Do() // Lista instancias agregadas por proyecto.
	if err != nil {                                                                     // Si falla el listado.
		return fmt.Errorf("error listando instancias: %w", err) // Retorna error.
	}

	count := 0                                 // Contador de instancias coincidentes.
	for _, zoneList := range instances.Items { // Itera por zonas.
		for _, instance := range zoneList.Instances { // Itera por instancias en zona.
			if matchPattern(instance.Name, namePattern) { // Verifica si el nombre coincide.
				count++ // Incrementa contador.
			}
		}
	}

	if count < minCount { // Verifica si se alcanzó el recuento mínimo.
		return fmt.Errorf("solo se encontraron %d instancias (mínimo requerido: %d)", count, minCount) // Retorna error si falta.
	}

	return nil // Éxito.
}

// WaitForResourceReady espera hasta que un recurso esté en estado deseado
// WaitForResourceReady realiza espera activa (polling) hasta que un recurso alcance un estado o expire el tiempo.
func (c *Client) WaitForResourceReady(ctx context.Context, resourceType, resourceName, targetStatus string, timeout time.Duration) error {
	// Implementación simplificada - en producción usar polling con backoff
	ctx, cancel := context.WithTimeout(ctx, timeout) // Contexto con timeout.
	defer cancel()                                   // Cancelar al salir.

	ticker := time.NewTicker(5 * time.Second) // Ticker para polling cada 5s.
	defer ticker.Stop()                       // Parar ticker al salir.

	for { // Bucle infinito.
		select { // Selección de canales.
		case <-ctx.Done(): // Si el contexto termina (timeout o cancel).
			return fmt.Errorf("timeout esperando recurso %s/%s en estado %s", resourceType, resourceName, targetStatus) // Error de timeout.
		case <-ticker.C: // En cada tick.
			// Aquí iría la lógica específica para cada tipo de recurso
			// Por ejemplo, para Compute Engine:
			if resourceType == "compute-instance" { // Si es instancia compute.
				instance, err := c.compute.Instances.Get(c.projectID, "us-central1-a", resourceName).Context(ctx).Do() // Obtiene instancia (zona hardcodeada temporalmente).
				if err != nil {                                                                                        // Si falla obtención.
					continue // aún no existe, reintenta.
				}
				if instance.Status == targetStatus { // Si estado coincide.
					return nil // Éxito.
				}
			}
		}
	}
}

func matchPattern(name, pattern string) bool {
	// Implementación simple - en producción usar regexp
	if pattern == "*" { // Si patrón es wildcard total.
		return true // Coincide todo.
	}
	return name == pattern || (len(pattern) > 0 && pattern[len(pattern)-1] == '*' && name[:len(pattern)-1] == pattern[:len(pattern)-1]) // Coincidencia exacta o prefijo con *.
}
