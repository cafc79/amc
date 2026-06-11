// pkg/pipeline/pipeline.go
package pipeline // Define el paquete 'pipeline'.

import (
	"sync"
	"time"
)

// Pipeline represents a sequence of steps
// Pipeline representa una secuencia de pasos a ejecutar.
type Pipeline struct {
	name       string        // Nombre del pipeline.
	steps      []Step        // Lista de pasos (steps) que componen el pipeline.
	parallel   bool          // Indica si los pasos deben ejecutarse en paralelo.
	maxRetries int           // Número máximo de reintentos permitidos.
	timeout    time.Duration // Tiempo máximo de ejecución permitido para el pipeline.
	results    []StepResult  // Almacena los resultados de la ejecución de los pasos.
	mutex      sync.Mutex    // Mutex para asegurar acceso concurrente seguro a los campos.
}

// NewPipeline creates a new pipeline
// NewPipeline crea e inicializa una nueva instancia de Pipeline.
func NewPipeline(name string) *Pipeline {
	return &Pipeline{ // Retorna un puntero a un nuevo objeto Pipeline.
		name:       name,     // Asigna el nombre proporcionado.
		steps:      []Step{}, // Inicializa la lista de pasos vacía.
		parallel:   false,    // Por defecto, la ejecución no es paralela.
		maxRetries: 0,        // Por defecto, sin reintentos.
		timeout:    0,        // Por defecto, sin tiempo de espera (timeout).
	}
}

// AddStep adds a step to the pipeline
// AddStep agrega un nuevo paso a la lista de pasos del pipeline.
func (p *Pipeline) AddStep(step Step) *Pipeline {
	p.steps = append(p.steps, step) // Añade el paso al slice de pasos.
	return p                        // Retorna el pipeline para permitir encadenamiento de métodos.
}

// SetParallel sets whether steps should run in parallel
// SetParallel configura si los pasos deben ejecutarse en paralelo o secuencialmente.
func (p *Pipeline) SetParallel(parallel bool) *Pipeline {
	p.parallel = parallel // Establece el valor booleano para ejecución paralela.
	return p              // Retorna el pipeline.
}

// SetMaxRetries sets the maximum number of retries
// SetMaxRetries configura el número máximo de reintentos globales.
func (p *Pipeline) SetMaxRetries(retries int) *Pipeline {
	p.maxRetries = retries // Asigna el número de reintentos.
	return p               // Retorna el pipeline.
}

// SetTimeout sets the timeout for the pipeline
// SetTimeout define el tiempo límite para la ejecución del pipeline.
func (p *Pipeline) SetTimeout(timeout time.Duration) *Pipeline {
	p.timeout = timeout // Asigna la duración del timeout.
	return p            // Retorna el pipeline.
}
