package config // Define el paquete 'config'.

import (
	"fmt"
	"os"
	"time"

	//"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// PipelineConfig represents the top-level configuration
// PipelineConfig representa la configuración de nivel superior del pipeline.
type PipelineConfig struct {
	Name        string            `yaml:"name"`        // Nombre del pipeline.
	Timeout     time.Duration     `yaml:"timeout"`     // Tiempo máximo de ejecución global.
	MaxRetries  int               `yaml:"max_retries"` // Número máximo de reintentos globales.
	Environment map[string]string `yaml:"environment"` // Variables de entorno globales.
	Steps       []StepConfig      `yaml:"steps"`       // Lista de configuraciones de pasos.
}

// StepConfig contains the specific configuration for a step
// StepConfig contiene la configuración específica para un paso individual.
type StepConfig struct {
	Name       string                 `yaml:"name"`                  // Nombre del paso.
	Type       string                 `yaml:"type"`                  // Tipo de paso (ej. "exec", "http").
	When       string                 `yaml:"when,omitempty"`        // Condición de ejecución: "always", "success", "failure".
	Config     map[string]interface{} `yaml:"config,omitempty"`      // Configuración específica del tipo de paso (mapa genérico).
	Action     string                 `yaml:"action,omitempty"`      // Acción específica (si aplica).
	WorkingDir string                 `yaml:"working_dir,omitempty"` // Directorio de trabajo para comandos (ej. terraform/ansible).
	Rollback   *RollbackConfig        `yaml:"rollback,omitempty"`    // Configuración de reversión (rollback) en caso de fallo.
}

// RollbackConfig defines the rollback configuration for a step
// RollbackConfig define la configuración para revertir un paso.
type RollbackConfig struct {
	Action string                 `yaml:"action"`           // Acción de rollback: "destroy", "revert", etc.
	Config map[string]interface{} `yaml:"config,omitempty"` // Configuración adicional para el rollback.
}

// Load reads and parses the YAML configuration file
// Load lee y analiza un archivo de configuración YAML.
func Load(filename string) (*PipelineConfig, error) {
	data, err := os.ReadFile(filename) // Lee el contenido completo del archivo especificado.
	if err != nil {                    // Verifica si hubo un error al leer el archivo.
		return nil, fmt.Errorf("failed to read config file: %w", err) // Retorna un error envuelto con contexto.
	}

	var config PipelineConfig                             // Declara una variable para almacenar la configuración decodificada.
	if err := yaml.Unmarshal(data, &config); err != nil { // Decodifica el YAML en la estructura config.
		return nil, fmt.Errorf("failed to parse config file: %w", err) // Retorna un error si el parsing falla.
	}

	return &config, nil // Retorna un puntero a la configuración cargada y nil como error.
}
