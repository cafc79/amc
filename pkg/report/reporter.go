package report // Define el paquete 'report' que agrupa la funcionalidad de generación de reportes.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"pipeline-framework/pkg/pipeline"
)

// PipelineReport represents the final report of a pipeline execution
// PipelineReport define la estructura del reporte final de la ejecución de un pipeline.
type PipelineReport struct {
	PipelineName string                `json:"pipeline_name"` // Nombre del pipeline ejecutado. Se mapea a "pipeline_name" en JSON.
	Timestamp    time.Time             `json:"timestamp"`     // Marca de tiempo de cuándo se generó el reporte. Mapeado a "timestamp".
	Status       string                `json:"status"`        // Estado final del pipeline (ej. "success", "failed"). Mapeado a "status".
	Results      []pipeline.StepResult `json:"results"`       // Lista de resultados de cada paso del pipeline. Mapeado a "results".
}

// GenerateReport creates a new pipeline report
// GenerateReport es una función que crea y devuelve un nuevo reporte de pipeline.
func GenerateReport(pipelineName string, results []pipeline.StepResult) (*PipelineReport, error) {
	status := "success"           // Inicializa la variable 'status' con "success" por defecto.
	for _, res := range results { // Itera sobre cada resultado en la lista de resultados 'results'.
		if res.Status == pipeline.StatusFailed { // Verifica si el estado del paso actual es 'StatusFailed'.
			status = "failed" // Si hay un fallo, cambia el estado general a "failed".
			break             // Sale del bucle ya que un fallo es suficiente para marcar el pipeline como fallido.
		}
	}

	return &PipelineReport{ // Retorna una referencia a una nueva instancia de PipelineReport.
		PipelineName: pipelineName, // Asigna el nombre del pipeline recibido como argumento.
		Timestamp:    time.Now(),   // Asigna la hora actual como marca de tiempo.
		Status:       status,       // Asigna el estado calculado ("success" o "failed").
		Results:      results,      // Asigna la lista de resultados recibida.
	}, nil // Retorna nil como error, indicando que no hubo problemas.
}

// SaveToFile saves the report to a JSON file
// SaveToFile es un método de PipelineReport que guarda el reporte actual en un archivo JSON.
func (r *PipelineReport) SaveToFile(filename string) error {
	data, err := json.MarshalIndent(r, "", "  ") // Convierte la estructura (r) a formato JSON con indentación para legibilidad.
	if err != nil {                              // Verifica si ocurrió un error durante la conversión a JSON.
		return err // Retorna el error si la conversión falló.
	}

	dir := filepath.Dir(filename)                  // Obtiene el directorio padre de la ruta del archivo especificada.
	if err := os.MkdirAll(dir, 0755); err != nil { // Crea el directorio (y padres) si no existen, con permisos 0755. Comprueba errores.
		return err // Retorna el error si no se pudo crear el directorio.
	}

	return os.WriteFile(filename, data, 0644) // Escribe los datos JSON en el archivo con permisos 0644. Retorna cualquier error de escritura.
}
