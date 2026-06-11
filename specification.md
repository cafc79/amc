# Pipeline Framework Specification

Este documento proporciona la especificación del Framework de Pipeline basado en Python. Cualquier modificación futura al framework debe reflejarse y diseñarse primero aquí.

## 1. Overview

El `pipeline-framework` es una aplicación en Python diseñada para ejecutar pipelines operacionales, de infraestructura y de despliegue basados en una configuración YAML. Funciona como un orquestador, despachando distintos tipos de tareas (steps) ya sea de forma secuencial o en paralelo.

## 2. Directory Structure

```text
C:\xtra\gcp\amc\pipeline-framework\
├── main.py                     (Entry point)
├── requirements.txt            (Project dependencies)
└── pipeline_framework/
    ├── config/
    │   └── config.py           (Configuration models and YAML parser)
    ├── core/
    │   ├── pipeline.py         (Pipeline engine logic)
    │   ├── step.py             (Abstract Base Classes and Data structures)
    │   └── factory.py          (Step creation registry)
    ├── steps/                  (Implementations of specific pipeline steps)
    │   ├── ansible.py          
    │   ├── gcp.py              
    │   ├── notification.py     
    │   ├── policy.py           
    │   ├── terraform.py        
    │   ├── test.py             
    │   └── generic.py          
    └── report/
        └── reporter.py         (JSON generation)
```

## 3. Core Components

### 3.1. configuration (`config.py`)

- Define las dataclasses PipelineConfig, StepConfig y RollbackConfig.
- Expone la función load_config(filename: str) utilizando yaml.safe_load.

### 3.2. Steps (`step.py` & `factory.py`)

- Cada step DEBE heredar de `pipeline_framework.core.step.Step`.
- Un Step debe implementar:
    - `name() -> str`: Retorna el nombre del step.
    - `type() -> str`: Retorna el identificador de tipo del step.
    - `execute() -> StepResult`: Ejecuta la lógica, captura errores arbitrarios y retorna un `StepResult` correctamente estructurado.
    - `rollback() -> None`: Ejecuta el rollback si falla la ejecución durante el pipeline.
- Los nuevos steps se registran mediante `pipeline_framework.core.factory.register_step_type(type_string, factory_callable)`.

### 3.3. Pipeline Execution (`pipeline.py`)

- Administrado por la clase Pipeline.
- Soporta ejecución secuencial y paralela (usando concurrent.futures.ThreadPoolExecutor).
- Soporta un número arbitrario de max_retries. Si un step falla, aplica un backoff lineal mediante time.sleep(retries) basado en el contador actual de reintentos.
- En caso de fallo de un step y sin reintentos disponibles, ejecuta rollback() en todos los steps previamente completados en orden inverso.

### 3.4 Plugins / Step Types

1. **`ansible`**: Ejecuta `ansible-playbook` mediante subprocess. Requiere la clave `playbook`. inventory es opcional.
2. **`gcp-verify`**: Utiliza `googleapiclient` para contar instancias de Compute Engine y validar contra un patrón `instance_name`.
3. **`terraform`**: Encapsula el binario de `terraform`.  Soporta  `init`, `plan`, `apply`, `destroy` dependiendo de la acción (`action`).
4. **`notification`**: Simula el envío de una notificación anteponiendo y mostrando un `message`.
5. **`policy-check`**: Simula una lógica de validación de políticas.
6. **`infrastructure-test`**: Simula un test de infraestructura como un ping a un endpoint o verificación TCP.
7. **`noop`** / **`wait`**: Steps genéricos utilitarios. wait puede interpretar duraciones legibles por humanos (por ejemplo, "10m").

## 4. Reporting (`reporter.py`)

Al finalizar cualquier ejecución (éxito o fallo), se genera un `PipelineReport` en un archivo JSON (por defecto: `pipeline-report.json`). Incluye duración, errores y estado de cada step.

## Modifying the Framework

Para agregar una nueva capacidad o step:

1. Documentar la adición en la sección 3.4 `Plugins / Tipos de Step` dentro de este archivo de especificación.
2. Implementar la lógica en Python dentro de `pipeline_framework/steps/`.
3. Crear la función factory `new_step`.
4. Registrar el nuevo step dentro de la función `main()` en `main.py`.
