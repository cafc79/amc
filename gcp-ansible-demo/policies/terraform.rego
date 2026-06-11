# examples/gcp-ansible-demo/policies/terraform.rego (OPA Policy)

package policy // Define el paquete 'policy' para las reglas de OPA.

# Requerir etiquetas obligatorias
deny[msg] { // Define una regla 'deny' que genera un mensaje si se cumple la condición (conjunto).
    resource := input.resource_changes[_] // Itera sobre cada recurso cambiado en la entrada (plan de Terraform).
    resource.type == "google_compute_instance" // Filtra solo los recursos de tipo instancia de Google Compute.
    resource.change.actions[_] == "create" // Verifica si la acción es 'create' (creación de recurso).
    not has_label(resource.change.after, "environment") // Llama al helper 'has_label' para negar si tiene la etiqueta "environment".
    msg := sprintf("❌ instancia '%s' requiere etiqueta 'environment'", [resource.name]) // Formatea el mensaje de error con el nombre del recurso.
}

deny[msg] { // Otra regla 'deny' para verificar otra etiqueta obligatoria.
    resource := input.resource_changes[_] // Itera sobre los cambios de recursos.
    resource.type == "google_compute_instance" // Filtra por instancias de Google Compute.
    resource.change.actions[_] == "create" // Verifica si es una creación.
    not has_label(resource.change.after, "cost-center") // Niega si el recurso tiene la etiqueta "cost-center".
    msg := sprintf("❌ instancia '%s' requiere etiqueta 'cost-center'", [resource.name]) // Genera mensaje de error.
}

# Restringir tipos de máquina en producción
deny[msg] { // Regla 'deny' para restringir tipos de máquinas en producción.
    resource := input.resource_changes[_] // Itera sobre los cambios.
    resource.type == "google_compute_instance" // Filtra instancias de computación.
    resource.change.actions[_] == "create" // Solo aplica a creaciones.
    has_label(resource.change.after, "environment") // Verifica que tenga la etiqueta environment.
    resource.change.after.labels.environment == "production" // Verifica que el entorno sea "production".
    not allowed_prod_machine(resource.change.after.machine_type) // Niega si el tipo de máquina no está permitido por el helper.
    msg := sprintf("❌ máquina '%s' no permitida en producción", [resource.change.after.machine_type]) // Genera mensaje de error con el tipo de máquina.
}

# Bloquear buckets públicos
deny[msg] { // Regla 'deny' para controles de seguridad en buckets.
    resource := input.resource_changes[_] // Itera sobre los cambios.
    resource.type == "google_storage_bucket" // Filtra buckets de Google Storage.
    resource.change.actions[_] == "create" // Solo aplica a creaciones.
    resource.change.after.force_destroy == true // Verifica si 'force_destroy' está habilitado (true).
    msg := sprintf("❌ bucket '%s' no debe tener force_destroy=true", [resource.name]) // Genera mensaje de prohibición.
}

# Helpers
has_label(obj, key) { // Helper para verificar existencia de una etiqueta.
    obj.labels[key] // Evalúa si la clave 'key' existe dentro del mapa 'labels' del objeto.
}

allowed_prod_machine(machine_type) { // Helper para definir máquinas permitidas. Opción 1.
    contains(machine_type, "e2-micro") // Permite si el tipo contiene "e2-micro".
}

allowed_prod_machine(machine_type) { // Helper para definir máquinas permitidas. Opción 2 (OR lógico en Rego).
    contains(machine_type, "e2-small") // Permite si el tipo contiene "e2-small".
}

# Implementación simple de contains (OPA tiene strings.contains en versiones recientes)
contains(str, substr) { // Implementación manual de función 'contains'.
    count := len(str) // Obtiene longitud de la cadena principal.
    subcount := len(substr) // Obtiene longitud de la subcadena.
    count >= subcount // Verifica que la cadena sea al menos tan larga como la subcadena.
    some i // Declara variable 'i' para iterar índices.
    i <= count - subcount // Restringe el rango de 'i'.
    substring(str, i, i + subcount) == substr // Verifica si la subsección desde 'i' coincide con 'substr'.
}

substring(str, start, end) = result { // Helper para extraer subcadenas.
    chars := split(str, "") // Divide el string en caracteres individuales.
    slice := slice_chars(chars, start, end) // Obtiene el slice de caracteres según índices.
    result := concat("", slice) // Une los caracteres nuevamente en un string.
}

slice_chars(chars, start, end) = slice { // Helper para hacer slicing de arrays.
    slice := [chars[i] | some i; i >= start; i < end] // Comprensión de lista para generar el slice desde 'start' hasta 'end'.
}