# PROJECT_INDEX — NutriControl API

Mapa de navegación, comandos de desarrollo y estado actual del proyecto. No reemplaza a `docs/constitution.md` (principios) ni a `odd/task/estado-actual-proyecto.md` (estado vivo ODD).

## 1. Propósito y stack

**NutriControl API** es un backend Go minimalista (Echo v4) para la gestión de pacientes. CRUD de pacientes sobre Firestore real o mock en memoria.

- **Runtime:** Go 1.26.1
- **HTTP:** `github.com/labstack/echo/v4` (router + bind + JSON)
- **Persistencia:** `cloud.google.com/go/firestore` v1.28 (Firestore real o emulador)
- **Testing:** `github.com/stretchr/testify` (assert/require)
- **Inyección:** manual vía argumento a handlers (no DI container).

## 2. Estructura de carpetas

```
nutricontrol/
├── main.go                        # Entry point. Setup de Echo + inyección repo.
├── go.mod / go.sum                # Módulo `nutricontrol`.
├── PROJECT_INDEX.md               # ESTE archivo: navegación, comandos, estado.
├── AGENTS.md                      # Contexto operativo breve (arranque rápido).
├── README.md                      # Bitácora histórica día a día (no editar por sesión).
├── todo                           # Checklist ligero de la sesión actual.
├── handlers/                      # Capa presentation (HTTP).
│   ├── paciente.go                # Handlers + SetupRoutes (inyecta repo).
│   └── paciente_test.go           # Tests (verdes).
├── models/                        # Capa domain (entidades puras, sin deps).
│   └── paciente.go                # struct Paciente.
│   └── paciente_test.go           # Test de serialización JSON.
├── pkg/
│   └── firestore/                 # Adaptador Firestore (puerto: PacienteRepository).
│       ├── paciente_repo.go       # Interfaz PacienteRepository.
│       ├── repo.go                # Impl Firestore real.
│       └── mock.go                # Impl mock en memoria.
├── docs/
│   ├── constitution.md            # Principios innegociables (hexagonal, testing, minimalismo).
│   ├── CHANGELOG.md               # Historia de cambios contractuales.
│   └── organizacion-documentacion.md  # Mapa de esta documentación y porqué.
└── odd/
    ├── task/
    │   └── estado-actual-proyecto.md  # Único archivo de definición de estado actual (ODD).
    └── tasks/
        └── archive/               # Feature trackings cerrados (no se editan).
```

- `models/` = dominio. Sin dependencias externas. `Paciente` es el único agregado.
- `pkg/firestore/` = adaptador de persistencia. Expone la interfaz + dos implementaciones (real + mock).
- `handlers/` = orquestación HTTP. Recibe `firestore.PacienteRepository` por argumento; no construye el repo.

## 3. Entry points

### 3.1 Arranque — `main.go`

Lee env vars: `NUTRICONTROL_USE_MOCK=true`, `GCP_PROJECT_ID=<id>`.

- Si mock → `firestore.NewMockPacienteRepository()`.
- Si real → `firestore.NewPacienteFirestoreRepo(ctx, projectID)` (falla abiertamente si no hay credenciales GCP).
- Registra rutas vía `handlers.SetupRoutes(e, repo)`.

### 3.2 Rutas (contrato API)

| Método | Ruta | Handler | Repo método | Éxito |
|--------|------|---------|-------------|-------|
| GET | `/pacientes` | GetPacientes | GetAll | 200 OK `[Paciente]` |
| POST | `/pacientes` | CreatePaciente | Create | 201 Created `Paciente` (el mock asigna `ID` si viene vacío) |
| GET | `/pacientes/{id}` | GetPacienteByID | GetByID | 200 OK `Paciente` / 404 |
| PUT | `/pacientes/{id}` | UpdatePaciente | Update | 200 OK `Paciente` / 404 |
| DELETE | `/pacientes/{id}` | DeletePaciente | Delete | 204 No Content / 404 |
| GET | `/` | (inline echo) | — | 200 "API NutriControl Operativa" |

- `{id}` proviene de `c.Param("id")`.
- Validaciones: `nombre` requerido; `edad` y `peso` deben ser > 0 (commit `8a1a47f`). 400 con `{"error": "..."}` si falta o es inválido.

## 4. Contratos clave

### 4.1 Puerta `PacienteRepository` — `pkg/firestore/paciente_repo.go`

```go
type PacienteRepository interface {
    Create(ctx context.Context, p models.Paciente) (*models.Paciente, error)
    GetAll(ctx context.Context) ([]models.Paciente, error)
    GetByID(ctx context.Context, id string) (*models.Paciente, error)
    Update(ctx context.Context, id string, p models.Paciente) (*models.Paciente, error)
    Delete(ctx context.Context, id string) error
}
```

### 4.2 Adaptadores

- **Firestore real** (`pkg/firestore/repo.go`): `PacienteFirestoreRepo`.
- **Mock** (`pkg/firestore/mock.go`): `MockPacienteRepository`. Store en `[]models.Paciente` (no persiste entre arranques).

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

## 5. Comandos de desarrollo

### 5.1 Dependencias
```bash
go mod tidy            # Limpia go.mod/go.sum
```

### 5.2 Build / Run
```bash
go build -o nutricontrol.exe ./main.go   # Windows
go run ./...                              # Ejecuta main.go (puerto :8080)
```

### 5.3 Testing
```bash
go test ./...                    # Todos los paquetes
go test -v ./handlers/           # Solo handlers
go test -v ./models/             # Solo models
go test -cover ./...             # Con cobertura
go vet ./...                     # Verificación estándar
```

**Estado actual:** `go test ./...` verde (`handlers`, `models`).

### 5.4 Variables de entorno

| Variable | Obligatorio | Valor | Propósito |
|----------|-------------|-------|-----------|
| `NUTRICONTROL_USE_MOCK` | No | `true` | Forzar mock en memoria. |
| `GCP_PROJECT_ID` | No | `<id>` | ID proyecto GCP para Firestore real. |
| `GOOGLE_APPLICATION_CREDENTIALS` | No | `/path/cred.json` | Credenciales de servicio GCP. |

## 6. Estado actual (post C.2 y C.3)

| Área | Estado |
|------|--------|
| Código fuente | Sin cambios pendientes en la rama. `main` queda en `9102ca9`. |
| Tests | `go test ./...` verde (`handlers`, `models`, `pkg/firestore`). |
| Build | `go build ./...` y `go vet ./...` OK. |
| Docs | `PROJECT_INDEX.md` consolidado. `odd/task/estado-actual-proyecto.md` vigente. |
| Seguimiento ODD | `odd/task/estado-actual-proyecto.md` vigente (único archivo de definición de estado actual). |
| Ignorados | `*.exe`, `*.dll`, `.atl/` vía `.gitignore` (commit `3f78f12`). |
| Pendiente | Opción C: otra feature nueva (definir cuál). A y B cerradas. |

### 6.1 Features C.2 y C.3 (completadas)

**C.2 — Mock asigne ID autoincremental al crear**
- `MockPacienteRepository.Create` asigna `paciente-<n>` automáticamente cuando `p.ID` viene vacío, usando `atomic.AddInt64(&m.nextID, 1)`. Respeta el ID si ya viene seteado.
- Se agregaron `GetByID`, `Update` y `Delete` al mock, permitiendo que los handlers correspondientes funcionen con datos creados en memoria.
- Handler `GetPacienteByID` (GET `/pacientes/{id}`): devuelve 200 con el paciente o 404 si no existe.
- Handler `DeletePaciente` (DELETE `/pacientes/{id}`): devuelve 204 o 404 si no existe.
- Tests: `pkg/firestore/mock_test.go` (`TestMockCreateAsignaIDCuandoEstaVacio`, `TestMockCreateRespetaIDExistente`) y `handlers/paciente_test.go` (`TestGetPacienteByIDOK`, `TestGetPacienteByIDNoEncontrado`).
- Commit: `65b514d feat: tests de GetPacienteByID y mock asigna ID al crear`.

**C.3 — Endpoint `GET /pacientes?nombre=`**
- `GetPacientes` filtra por `c.QueryParam("nombre")` con coincidencia parcial y sin distinción de mayúsculas/minúsculas (`strings.Contains` + `strings.ToLower`).
- Test: `handlers/paciente_test.go` (`TestGetPacientesFiltrarPorNombre`).
- Commit: `9102ca9 feat: filtrar pacientes por nombre en GET /pacientes`.

### 6.2 Dificultades encontradas y cómo las resolví

**1. `c.Response().Body.Bytes()` no compila en Echo v4 (handlers de test)**
- **Síntoma:** `go test ./...` fallaba en `handlers/paciente_test.go` con error de compilación: `c.Response.Body undefined (type *echo.Response has no field or method Body)`.
- **Causa:** Echo v4 eliminó el campo `.Body` de `*echo.Response`. El cuerpo de respuesta en tests se lee ahora a través del `httptest.ResponseRecorder`.
- **Resolución:** Reemplacé `c.Response().Body.Bytes()` por `c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes()` en las 3 ocurrencias del test. Commit `b3754f1` (fix).

**2. Parámetros de ruta (`c.Param("id")`) no disponibles al llamar handlers directamente en tests**
- **Síntoma:** Al invocar `GetPacienteByID`, `UpdatePaciente` o `DeletePaciente` directamente desde el test (sin pasar por el router de Echo), `c.Param("id")` devolvía string vacío.
- **Causa:** Echo v4 almacena los parámetros de ruta en campos no exportados `pnames []string` y `pvalues []string` del contexto. Cuando el handler se llama directo, esos campos están vacíos.
- **Resolución:** Creé el helper `setParam(c echo.Context, key, value string)` que usa `reflect` + `unsafe.Pointer` para inyectar `pnames`/`pvalues` directamente en el contexto de Echo v4. Se usa en `TestUpdatePacienteEdadInvalida`, `TestUpdatePacientePesoInvalido`, `TestGetPacienteByIDOK` y `TestGetPacienteByIDNoEncontrado`.

**3. `MockPacienteRepository.Create` no asignaba ID al crear (bloqueaba GetByID/Delete sobre datos nuevos)**
- **Síntoma:** Al crear un paciente sin ID y luego intentar `GetByID` o `Delete`, el paciente no se encontraba porque `ID` quedó vacío.
- **Causa:** El mock devolvía el paciente tal cual sin asignarle un identificador automático.
- **Resolución:** Agregé contador atómico `nextID` en `MockPacienteRepository` y lógica en `Create` para asignar `paciente-<n>` cuando `p.ID == ""`, respetando el ID si ya viene seteado.

**4. go.sum grande tras agregar `cloud.google.com/go/firestore`**
- **Síntoma:** `go.sum` crecía ~30 líneas la primera vez que se resolvía la dependencia de Firestore real.
- **Resolución:** Ejecutar `go mod tidy` antes de commitear para que `go.mod` y `go.sum` queden consistentes.

## 7. Flujo de commits

- Estilo: **Conventional Commits** (`feat:`, `fix:`, `refactor:`, `docs:`, `build:`, `chore:`).
- Rama de trabajo: crear feature branch desde `main` antes del primer cambio de fuente.
- Antes de commitear: `go mod tidy`, `go build ./...`, `go test ./...`.
- Push y creación de PR quedan a criterio del usuario bajo política ordinaria del repositorio.