package pipeline // Define el paquete 'pipeline'.

import (
	"fmt"
	"sync"

	"pipeline-framework/pkg/config"
)

var (
	stepRegistry = make(map[string]StepFactory) // Mapa global para registrar las fábricas de pasos por tipo.
	registryMu   sync.RWMutex                   // Mutex de lectura/escritura para proteger el acceso concurrente al registro.
)

// RegisterStepType registers a new step type with its factory function
// RegisterStepType registra un nuevo tipo de paso asociado a su función de fábrica.
func RegisterStepType(stepType string, factory StepFactory) {
	registryMu.Lock()                // Bloquea el mutex para escritura exclusiva.
	defer registryMu.Unlock()        // Asegura que el mutex se desbloquee al salir de la función.
	stepRegistry[stepType] = factory // Asigna la función de fábrica al tipo de paso especificado en el mapa.
}

// CreateStep creates a new step based on the provided configuration
// CreateStep crea una nueva instancia de un paso basándose en la configuración proporcionada.
func CreateStep(cfg config.StepConfig) (Step, error) {
	registryMu.RLock()                        // Bloquea el mutex para lectura compartida.
	factory, exists := stepRegistry[cfg.Type] // Busca la fábrica correspondiente al tipo de paso en la configuración.
	registryMu.RUnlock()                      // Desbloquea el mutex de lectura.

	if !exists { // Si no existe una fábrica registrada para ese tipo.
		return nil, fmt.Errorf("unknown step type: %s", cfg.Type) // Retorna un error indicando que el tipo de paso es desconocido.
	}

	return factory(cfg) // Llama a la función de fábrica con la configuración y retorna el paso creado (o error).
}
