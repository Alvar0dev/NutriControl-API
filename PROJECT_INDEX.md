# PROJECT_INDEX — NutriControl API

> Mapa de navegación arquitectónica para agentes exploradores. No reemplaza a docs/constitution.md ni a specs/*.

## 1. Propósito y stack

**NutriControl API** es un backend Go minimalista (Echo v4) para la gestión de pacientes nutritionist-controlled. En fase inicial — CRUD de pacientes sobre Firestore, con alternancia mock/emulador.

- **Runtime:** Go 1.26.1
- **HTTP:** `github.com/labstack/echo/v4` (router + bind + JSON)
- **Persistencia:** `cloud.google.com/go/firestore` v1.28 (Firestore real o emulador)
- **Testing:** `github.com/stretchr/testify` (assert/require)
- **Inyección:** manual vía argumento a handlers (no DI container).

## 2. Estructura de carpetas (alto nivel)

```
nutricontrol/
├── main.go                        # Entry point. Setup de Echo + inyección repo.
├── go.mod / go.sum                # Módulo `nutricontrol`.
├── PROJECT_INDEX.md               # ESTE archivo.
├── DEV_RUNBOOK.md                 # Comandos build/test/env.
├── AGENTS.md                      # Contexto operativo por día.
├── handlers/                      # Capa presentation (HTTP).
│   ├── paciente.go                # Handlers + SetupRoutes (inyecta repo).
│   └── paciente_test.go           # Tests (rotos: Echo v4 Response.Body).
├── models/                        # Capa domain (entidades puras, sin deps).
│   └── paciente.go                # struct Paciente.
├── pkg/
│   └── firestore/                 # Adaptador Firestore (puerto: PacienteRepository).
│       ├── paciente_repo.go       # Interfaz PacienteRepository.
│       ├── repo.go                # Impl Firestore real: NewPacienteFirestoreRepo + CRUD.
│       └── mock.go                # Impl mock en memoria: NewMockPacienteRepository + CRUD.
└── docs/
    ├── constitution.md            # Principios innegociables (hexagonal, testing purista).
    └── CHANGELOG.md               # Historia de cambios contractuales.
```

- `models/` = dominio. Sin dependencias externas. `Paciente` es el único agregado.
- `pkg/firestore/` = adaptador de persistencia. Expone la interfaz + dos implementaciones (real + mock). Es el puente entre dominio y Firestore, no un "service".
- `handlers/` = orquestación HTTP. Recibe `firestore.PacienteRepository` por argumento; no construye el repo.

## 3. Entry points

### 3.1 Arranque — `main.go`
- Lee env vars: `NUTRICONTROL_USE_MOCK=true`, `GCP_PROJECT_ID=<id>`.
- Si mock → `firestore.NewMockPacienteRepository()`.
- Si real → `firestore.NewPacienteFirestoreRepo(ctx, projectID)` (falla abiertamente si no hay credenciales GCP).
- Registra rutas vía `handlers.SetupRoutes(e, repo)`.

### 3.2 Rutas (contrato API)
Definidas en `handlers/paciente.go :: SetupRoutes`. Formato `{método} {ruta} -> handler`.

| Método | Ruta | Handler | Repo método | Éxito |
|--------|------|---------|-------------|-------|
| GET | `/pacientes` | GetPacientes | GetAll | 200 OK `[Paciente]` |
| POST | `/pacientes` | CreatePaciente | Create | 201 Created `Paciente` |
| GET | `/pacientes/{id}` | GetPacienteByID | GetByID | 200 OK `Paciente` / 404 |
| PUT | `/pacientes/{id}` | UpdatePaciente | Update | 200 OK `Paciente` / 404 |
| DELETE | `/pacientes/{id}` | DeletePaciente | Delete | 204 No Content / 404 |
| GET | `/` | (inline echo) | — | 200 "API NutriControl Operativa" |

- `{id}` proviene de `c.Param("id")`.
- Validaciones actuales: `nombre` requerido en POST/PUT. 400 con `{"error": "..."}` si falta.
- Mock no persiste `ID` por omisión (struct mock no lo setea). Cuidado: busquedas por ID pueden no encontrar mock data.

## 4. Contratos clave (puertos/adaptadores)

### 4.1 Puerta `PacienteRepository` — `pkg/firestore/paciente_repo.go`
Interfaz de repositorio (el "puerto"). Los handlers dependen SOLO de ella, no de la impl.

```go
type PacienteRepository interface {
    Create(ctx, models.Paciente) (*models.Paciente, error)
    GetAll(ctx) ([]models.Paciente, error)
    GetByID(ctx, id) (*models.Paciente, error)
    Update(ctx, id, models.Paciente) (*models.Paciente, error)
    Delete(ctx, id) error
}
```

### 4.2 Adaptadores
- **Firestore real** (`pkg/firestore/repo.go`): `PacienteFirestoreRepo`. `NewPacienteFirestoreRepo(ctx, projectID)` falla si `option.WithEndpoint("firestore.googleapis.com:443")` no resuelve credenciales. Métodos: Create/Add doc, GetAll/itera Documents, GetByID/doc.Get, Update/Set, Delete/Doc(id).Delete.
- **Mock** (`pkg/firestore/mock.go`): `MockPacienteRepository`. Store en `[]models.Paciente` (slice, no persiste entre arranques). NOTA: `Update` corrige índice del slice (antes bug de variable local). `Delete` re-slicea preservando orden.

### 4.3 Modelo de dominio — `models/paciente.go`
```go
type Paciente struct {
    ID              string  `json:"id,omitempty"`
    Nombre          string  `json:"nombre"`
    Edad            int     `json:"edad"`
    Peso            float64 `json:"peso"`
    PorcentajeGrasa float64 `json:"porcentaje_grasa"`
    Notas           string  `json:"notas"`
}
```
- Es el DTO y la entidad. Una sola fuente de verdad para schema JSON y Firestore.
- Regla de negocio crítica (refs en handler + model): `nombre` obligatorio en POST/PUT. La entidad misma no valida (validación en handlers → ver constitución prin. 2).

## 5. Flujo de datos crítico (request → persistencia)

### Flujo típico `POST /pacientes`:
```
HTTP client
  -> Echo router (handlers.SetupRoutes)
  -> CreatePaciente(c, repo)
     -> c.Bind(&models.Paciente)            # Echo decodea JSON a entidad dominio
     -> VALIDACIÓN (nombre != "")           # 400 Bad Request si falla
     -> repo.Create(ctx, paciente)          # ← IYECCIÓN: solo conoce la interfaz
        -> [Firestore] col.Add(doc)         # docRef.ID asignado por Firestore
        -> paciente.ID = docRef.ID          # se setea id de vuelta en la entidad
     -> 201 Created JSON {id, ...}
```
- Punto de sustitución para testing: pasar `MockPacienteRepository` en lugar del Firestore real. Ver `handlers/paciente_test.go` (usa mock).

### Flujo típico `GET /pacientes/{id}`:
```
HTTP -> GetPacienteByID -> c.Param("id") -> repo.GetByID -> 200 | 404
```
- Firestore GetByID falla con error genérico si Firestore devuelve not-found; el handler mapea a 404 con mensaje. El mock falla con error de "no se encontró" → también 404. **Importante:** ambos adapters devuelven error en "no existe", el handler unifica a 404.

## 6. Puntos de sustitución / extensiones conocidos
- **Persistencia alternativa (SQL):** la constitución (#4) menciona sqlx. La interfaz `PacienteRepository` es el puente; un nuevo adaptador SQL implementaría la misma interfaz en `pkg/sql/` (no existe todavía).
- **Validación profunda:** validar edad/peso numéricos (RF-03 de spec antigua) no está implementado — solo `nombre`. Futuro candidate en `internal/validation/` o `models`.
- **Logging/trazas:** no implementado (out of MVP). Constitución #1 (minimalismo) implica no forzar loggers hasta que la complejidad lo exija.
- **Tests handlers (Echo v4):** `c.Response().Body` fue eliminado en Echo v4; se corrigió a `c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes()` en `b3754f1`. `go test ./...` pasa: `handlers` y `models` verdes.

## 7. Estado actual (post `366d4a8` en `main`)

| Área | Estado |
|------|--------|
| Código fuente | Intacto, sin cambios pendientes de commit salvo docs |
| Tests | `go test ./...` verde (`handlers`, `models`) |
| Docs | `docs/constitution.md` actualizado con contratos ODD; `specs/archive/` eliminado |
| Seguimiento ODD | `odd/tasks/000-reestructuracion-documental.md` vigente; archivo `todo` en raíz con pendiente para mañana |
| Pendiente | `B`) completar `PROJECT_INDEX.md` / `DEV_RUNBOOK.md` (en curso); `C`) nueva feature (repo o endpoint) |
