// cmd/piper/main.go (con reportes y notificaciones finales)
package main // Define el paquete principal de la aplicación.

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"pipeline-framework/pkg/ansible"
	"pipeline-framework/pkg/config"
	"pipeline-framework/pkg/gcp"
	"pipeline-framework/pkg/notification"
	"pipeline-framework/pkg/pipeline"
	"pipeline-framework/pkg/policy"
	"pipeline-framework/pkg/report"
	"pipeline-framework/pkg/terraform"
	"pipeline-framework/pkg/test"
)

func main() {
	configFile := flag.String("config", "pipeline.yaml", "configuration file")          // Define el flag para el archivo de configuración.
	outputReport := flag.String("report", "pipeline-report.json", "output report file") // Define el flag para el archivo de reporte de salida.
	flag.Parse()                                                                        // Parsea los argumentos de la línea de comandos.

	// Load configuration
	cfg, err := config.Load(*configFile) // Carga la configuración desde el archivo especificado.
	if err != nil {                      // Verifica si hubo error al cargar la configuración.
		log.Fatalf("❌ Error loading config: %v", err) // Muestra el error y termina el programa si falla la carga.
	}

	// Create pipeline
	p := pipeline.NewPipeline(cfg.Name). // Crea una nueva instancia de pipeline con el nombre de la configuración.
						SetTimeout(cfg.Timeout).      // Establece el tiempo límite global del pipeline.
						SetMaxRetries(cfg.MaxRetries) // Establece el número máximo de reintentos globales.

	// Register step types
	pipeline.RegisterStepType("terraform", terraform.NewStep)       // Registra el tipo de paso 'terraform'.
	pipeline.RegisterStepType("gcp-verify", gcp.NewStep)            // Registra el tipo de paso 'gcp-verify'.
	pipeline.RegisterStepType("policy-check", policy.NewStep)       // Registra el tipo de paso 'policy-check'.
	pipeline.RegisterStepType("infrastructure-test", test.NewStep)  // Registra el tipo de paso 'infrastructure-test'.
	pipeline.RegisterStepType("notification", notification.NewStep) // Registra el tipo de paso 'notification'.
	pipeline.RegisterStepType("ansible", ansible.NewStep)           // Registra el tipo de paso 'ansible'.

	// Register generic steps
	pipeline.RegisterStepType("noop", newNoopStep) // Registra un paso genérico 'noop' (no operation).
	pipeline.RegisterStepType("wait", newWaitStep) // Registra un paso genérico 'wait' (espera).

	// Add steps from config
	for _, stepCfg := range cfg.Steps { // Itera sobre cada paso definido en la configuración.
		step, err := pipeline.CreateStep(stepCfg) // Crea la instancia del paso según su configuración.
		if err != nil {                           // Verifica si hubo error al crear el paso.
			log.Fatalf("❌ Error creating step '%s': %v", stepCfg.Name, err) // Termina el programa si no se puede crear un paso.
		}
		p.AddStep(step) // Agrega el paso creado al pipeline.
	}

	// Execute pipeline
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout) // Crea un contexto con el timeout configurado.
	defer cancel()                                                        // Asegura la cancelación del contexto al salir.

	fmt.Printf("🚀 Running pipeline: %s\n", cfg.Name)                             // Imprime mensaje de inicio del pipeline.
	fmt.Printf("⏱️  Timeout: %s | Retries: %d\n\n", cfg.Timeout, cfg.MaxRetries) // Imprime configuración de timeout y reintentos.

	startTime := time.Now()           // Registra el tiempo de inicio.
	results, err := p.Execute(ctx)    // Ejecuta el pipeline y obtiene resultados.
	duration := time.Since(startTime) // Calcula la duración total de la ejecución.

	// Generate report
	rpt, errReport := report.GenerateReport(cfg.Name, results) // Genera el reporte final con los resultados.
	if errReport != nil {                                      // Verifica si hubo error al generar el reporte.
		log.Printf("⚠️  Error generating report: %v", errReport) // Loguea advertencia si falla la generación.
	} else {
		if err := rpt.SaveToFile(*outputReport); err != nil { // Intenta guardar el reporte en archivo.
			log.Printf("⚠️  Error saving report: %v", err) // Loguea advertencia si falla el guardado.
		} else {
			fmt.Printf("\n📄 Report saved: %s\n", *outputReport) // Confirma que el reporte se guardó exitosamente.
		}
	}

	// Show summary
	fmt.Printf("\n=== PIPELINE SUMMARY ===\n")          // Imprime cabecera del resumen.
	fmt.Printf("Pipeline: %s\n", cfg.Name)              // Imprime nombre del pipeline.
	fmt.Printf("Duration: %.2fs\n", duration.Seconds()) // Imprime duración total en segundos.
	fmt.Printf("Status: ")                              // Etiqueta de estado.
	if err != nil {                                     // Verifica si hubo error global en la ejecución.
		fmt.Printf("❌ FAILED\n")       // Imprime estado FALLIDO.
		fmt.Printf("Error: %v\n", err) // Detalla el error.
		os.Exit(1)                     // Termina con código de error 1.
	} else {
		fmt.Printf("✅ SUCCESS\n") // Imprime estado EXITOSO.
	}
}

// Inline implementation for simple steps

type GenericStep struct {
	cfg  config.StepConfig                                      // Configuración del paso genérico.
	exec func(ctx context.Context, cfg config.StepConfig) error // Función que define la lógica de ejecución.
}

func (s *GenericStep) Name() string { return s.cfg.Name } // Retorna el nombre del paso.
func (s *GenericStep) Type() string { return s.cfg.Type } // Retorna el tipo de paso.
func (s *GenericStep) Execute(ctx context.Context) (*pipeline.StepResult, error) {
	start := time.Now()              // Registra tiempo de inicio.
	err := s.exec(ctx, s.cfg)        // Ejecuta la función lógica del paso.
	status := pipeline.StatusSuccess // Estado inicial éxito.
	if err != nil {                  // Si hay error.
		status = pipeline.StatusFailed // Cambia a fallido.
	}
	return &pipeline.StepResult{ // Retorna resultado encapsulado.
		Name:      s.cfg.Name,        // Nombre.
		Status:    status,            // Estado.
		Error:     err,               // Error.
		Duration:  time.Since(start), // Duración.
		Timestamp: start,             // Marca de tiempo.
	}, err // Error.
}
func (s *GenericStep) Rollback(ctx context.Context) error { return nil } // Rollback no implementado para genéricos (retorna nil).

func newNoopStep(cfg config.StepConfig) (pipeline.Step, error) {
	return &GenericStep{ // Retorna nueva instancia de GenericStep.
		cfg: cfg, // Asigna configuración.
		exec: func(ctx context.Context, cfg config.StepConfig) error { // Define lógica noop.
			if msg, ok := cfg.Config["message"].(string); ok { // Busca mensaje en config.
				fmt.Println(msg) // Imprime mensaje si existe.
			}
			return nil // Retorna éxito.
		},
	}, nil
}

func newWaitStep(cfg config.StepConfig) (pipeline.Step, error) {
	return &GenericStep{ // Retorna nueva instancia de GenericStep.
		cfg: cfg, // Asigna configuración.
		exec: func(ctx context.Context, cfg config.StepConfig) error { // Define lógica wait.
			durStr, _ := cfg.Config["duration"].(string) // Obtiene duración como string.
			dur, err := time.ParseDuration(durStr)       // Parsea la duración.
			if err != nil {                              // Si error en parseo.
				return err // Retorna error.
			}
			fmt.Printf("Waiting for %s...\n", dur) // Informa espera.
			select {
			case <-time.After(dur): // Espera el tiempo definido.
				return nil // Retorna éxito al completar tiempo.
			case <-ctx.Done(): // Si el contexto se cancela antes.
				return ctx.Err() // Retorna error del contexto.
			}
		},
	}, nil
}
