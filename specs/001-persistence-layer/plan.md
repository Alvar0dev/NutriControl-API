# Plan Técnico — Capa de Persistencia con Firestore

## Resumen Ejecutivo
Introducir una capa de repositorio que abstraiga Firestore, actualizando `handlers/paciente.go` para usarla y eliminando los datos mockeados. Se mantiene la arquitectura hexagonal actual (models y handlers separados).

---

## Archivos a Modificar/Crear

| Archivo | Acción | Propósito |
|---------|--------|-----------|
| `handlers/paciente.go` | MODIFICAR | Reemplazar datos mockeados por llamadas al repositorio |
| `go.mod` | MODIFICAR | Añadir dependencia `cloud.google.com/go/firestore` (con `go get`) |
| `pkg/firestore/client.go` | CREAR | Inicialización del cliente Firestore y acceso a base |
| `pkg/firestore/paciente_repo.go` | CREAR | `PacienteRepository` con métodos `Create`, `GetAll` |

---

## Modelos de Datos

### Existente — `models/paciente.go` (sin cambios)
```go
type Paciente struct {
    Nombre          string  `json:"nombre"`
    Edad            int     `json:"edad"`
    Peso            float64 `json:"peso"`
    PorcentajeGrasa float64 `json:"porcentaje_grasa"`
    Notas           string  `json:"notas"`
}
```
**Decisión:** Al no cambiar el modelo, el JSON externo sigue idéntico. El `id` generado por Firestore lo añadimos al devolver.

### Nuevo — `pkg/firestore/paciente_repo.go`
```go
type PacienteRepository interface {
    Create(ctx context.Context, p models.Paciente) (string, error)
    GetAll(ctx context.Context) ([]models.Paciente, error)
}
```

---

## Flujo Lógico (por endpoint)

### `POST /pacientes`
```
Cliente → handlers.CreatePaciente() 
         → Bind JSON a *models.Paciente
         → VALIDAR campos (RF-03)
         → repo.Create() [Firestore]
         → agrega ID generado
         → 201 Created con paciente
         ← Error de validación → 400 Bad Request
         ← Error Firestore → 500 Internal Server Error (RF-04)
```

### `GET /pacientes`
```
Cliente → handlers.GetPacientes()
         → repo.GetAll() [Firestore]
         → 200 OK con lista
         ← Error Firestore → 500 Internal Server Error (RF-04)
```

---

## Estrategia de Persistencia (Firestore Local)

- **Modo desarrollo:** Se usará el emulador de Firestore local o un archivo local (`firestore-kv.json`) para practicar sin necesidad de proyecto GCP.
- **Inyección de dependencias:** En `main.go` se inicializa el cliente y se pasa como dependencia a los handlers; evita código duplicado y facilita mocks en tests futuros.
- **Ruta de coexistencia:** Los handlers reciben un `PacienteRepository` como `func(c echo.Context, repo PacienteRepository)` o vía `echo.Context.Set()`/`Get()`. Se prefiere inyección directa para claridad.

---

## Estrategia de Pruebas (Definida pero implementada posterior)

Para esta fase inicial, se implementará una prueba manual/verificación estructural:
- Verificar con `curl` que `POST /pacientes` persiste y devuelve `id`.
- Verificar que `GET /pacientes` devuelve datos reales tras reinicio del servidor.
- Casos de error: probar datos inválidos (400) y error de Firestore (500).

---

## Dependencias Técnicas

```bash
# Firestore SDK (modo local o emulador)
go get cloud.google.com/go/firestore
```

**Alternativa menor:** Usar biblioteca `github.com/gorilla/sse` o archivo plano JSON si el enfoque Firestore local resulta complejo. Esto queda fuera del MVP.
</table>
</codeblock>
</contents>
</spec>

## Dependencies
- Go 1.26.1
- Echo v4.16.0
- cloud.google.com/go/firestore (nueva)

## Verification Steps
1. `go mod tidy`
2. Ejecutar tests de verificación
3. Reiniciar servidor y verificar persistencia</spec>