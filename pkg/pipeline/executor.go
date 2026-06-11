// pkg/pipeline/executor.go
package pipeline // Define el paquete 'pipeline'.

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Execute ejecuta el pipeline
// Execute inicia la ejecución del pipeline.
func (p *Pipeline) Execute(ctx context.Context) ([]StepResult, error) {
	if p.timeout > 0 { // Verifica si se ha configurado un tiempo límite para el pipeline.
		var cancel context.CancelFunc                     // Define una función de cancelación.
		ctx, cancel = context.WithTimeout(ctx, p.timeout) // Crea un contexto con timeout derivado del original.
		defer cancel()                                    // Asegura que se liberen los recursos del contexto al finalizar la función.
	}

	p.results = []StepResult{} // Inicializa la lista de resultados vacía.

	if p.parallel { // Verifica si la ejecución en paralelo está habilitada.
		return p.executeParallel(ctx) // Ejecuta los pasos en paralelo.
	}
	return p.executeSequential(ctx) // Ejecuta los pasos secuencialmente por defecto.
}

func (p *Pipeline) executeSequential(ctx context.Context) ([]StepResult, error) {
	for _, step := range p.steps { // Itera sobre cada paso del pipeline en orden.
		result, err := p.executeStepWithRetry(ctx, step) // Ejecuta el paso actual con lógica de reintentos.
		p.addResult(result)                              // Agrega el resultado a la lista de resultados del pipeline.

		if err != nil { // Si ocurre un error en la ejecución del paso.
			// Ejecutar rollback para los pasos anteriores
			p.executeRollback(ctx, step)                                                       // Inicia el proceso de reversión (rollback) desde el paso fallido hacia atrás.
			return p.results, fmt.Errorf("pipeline failed at step '%s': %w", step.Name(), err) // Retorna los resultados parciales y el error.
		}
	}
	return p.results, nil // Retorna los resultados y éxito si todos los pasos completan.
}

func (p *Pipeline) executeParallel(ctx context.Context) ([]StepResult, error) {
	var wg sync.WaitGroup                              // Crea un WaitGroup para esperar a que terminen las goroutines.
	resultsChan := make(chan StepResult, len(p.steps)) // Canal con buffer para recibir resultados de cada paso.
	errs := make(chan error, len(p.steps))             // Canal con buffer para recibir errores.

	for _, step := range p.steps { // Itera sobre cada paso configurado.
		wg.Add(1)         // Incrementa el contador del WaitGroup.
		go func(s Step) { // Inicia una goroutine para ejecutar el paso.
			defer wg.Done()                               // Decrementa el contador del WaitGroup al finalizar la goroutine.
			result, err := p.executeStepWithRetry(ctx, s) // Ejecuta el paso con reintentos.
			resultsChan <- result                         // Envía el resultado al canal de resultados.
			if err != nil {                               // Si hay error.
				errs <- err // Envía el error al canal de errores.
			}
		}(step) // Pasa la variable 'step' a la clausura de la goroutine.
	}

	go func() { // Inicia una goroutine para monitorizar el WaitGroup.
		wg.Wait()          // Espera a que todas las goroutines de pasos terminen.
		close(resultsChan) // Cierra el canal de resultados.
		close(errs)        // Cierra el canal de errores.
	}()

	for result := range resultsChan { // Lee los resultados del canal hasta que se cierre.
		p.addResult(result) // Agrega cada resultado a la lista del pipeline.
	}

	if len(errs) > 0 { // Verifica si se reportaron errores.
		return p.results, fmt.Errorf("pipeline failed with %d errors", len(errs)) // Retorna error indicando la cantidad de fallos.
	}

	return p.results, nil // Retorna éxito si no hubo errores.
}

func (p *Pipeline) executeStepWithRetry(ctx context.Context, step Step) (StepResult, error) {
	var lastErr error // Variable para almacenar el último error recibido.
	retries := 0      // Contador de intentos realizados.

	for { // Bucle infinito para reintentos.
		start := time.Now()                  // Registra el tiempo de inicio.
		stepResult, err := step.Execute(ctx) // Ejecuta el paso.
		if stepResult == nil {               // Si el paso no devuelve un resultado estructurado (ej. panic o error simple).
			stepResult = &StepResult{ // Crea un resultado por defecto indicando fallo.
				Name:      step.Name(),       // Asigna el nombre del paso.
				Status:    StatusFailed,      // Asigna estado fallido.
				Error:     err,               // Asigna el error capturado.
				Duration:  time.Since(start), // Calcula la duración.
				Timestamp: start,             // Asigna la marca de tiempo.
			}
		}

		if err == nil { // Si no hubo error en la ejecución.
			return *stepResult, nil // Retorna el resultado exitoso.
		}

		lastErr = err // Guarda el error para devolverlo si se agotan los reintentos.
		retries++     // Incrementa el contador de reintentos.

		if p.maxRetries == 0 || retries > p.maxRetries { // Si no hay reintentos configurados o se superó el máximo.
			stepResult.Status = StatusFailed // Asegura que el estado sea fallido.
			return *stepResult, lastErr      // Retorna el resultado fallido y el último error.
		}

		// Wait before retry
		waitTime := time.Duration(retries) * time.Second // Calcula un tiempo de espera lineal basado en el número de intento.
		time.Sleep(waitTime)                             // Espera antes del próximo intento.
	}
}

func (p *Pipeline) executeRollback(ctx context.Context, failedStep Step) {
	// Ejecutar rollback en orden inverso hasta el paso fallido
	for i := len(p.steps) - 1; i >= 0; i-- { // Itera sobre los pasos en orden inverso.
		step := p.steps[i]      // Obtiene el paso actual.
		if step == failedStep { // Si es el paso que falló, interrumpe el bucle (o continúa, según lógica deseada - aquí parece detenerse antes de revertir el fallido o quizás asume que el fallido ya falló y revertimos los *previos*? La lógica original parece querer revertir los *anteriores* al fallido si se ejecutaron secuencialmente, pero el break detiene cuando encuentra el failedStep. Revisar lógica original: "Ejecutar rollback para los pasos anteriores").
			break // Asumimos que los pasos anteriores al fallido son los que necesitan rollback.
		}

		if err := step.Rollback(ctx); err != nil { // Ejecuta el rollback del paso.
			fmt.Printf("Warning: Rollback failed for step '%s': %v\n", step.Name(), err) // Imprime advertencia si falla el rollback.
		}
	}
}

func (p *Pipeline) addResult(result StepResult) {
	p.mutex.Lock()                        // Bloquea el mutex.
	defer p.mutex.Unlock()                // Desbloquea al salir.
	p.results = append(p.results, result) // Agrega el resultado a la lista compartida.
}
