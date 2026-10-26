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
| POST | `/pacientes` | CreatePaciente | Create | 201 Created `Paciente` |
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

## 6. Estado actual (post `272b7cb` en `main`)

| Área | Estado |
|------|--------|
| Código fuente | Intacto, sin cambios pendientes de commit. |
| Tests | `go test ./...` verde (`handlers`, `models`). |
| Build | `go build ./...` OK. |
| Docs | `PROJECT_INDEX.md` consolidado (navegación + comandos + estado). `DEV_RUNBOOK.md` archivado. |
| Seguimiento ODD | `odd/task/estado-actual-proyecto.md` vigente (único archivo de definición de estado actual). |
| Ignorados | `*.exe`, `*.dll`, `.atl/` vía `.gitignore` (commit `3f78f12`). |
| Pendiente | Nueva feature: extender repo Firestore/mock o agregar endpoint (opción C). |

## 7. Flujo de commits

- Estilo: **Conventional Commits** (`feat:`, `fix:`, `refactor:`, `docs:`, `build:`, `chore:`).
- Rama de trabajo: crear feature branch desde `main` antes del primer cambio de fuente.
- Antes de commitear: `go mod tidy`, `go build ./...`, `go test ./...`.
- Push y creación de PR quedan a criterio del usuario bajo política ordinaria del repositorio.