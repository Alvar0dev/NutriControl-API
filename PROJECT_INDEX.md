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
├── .github/
│   └── workflows/
│       └── ci.yml                 # CI en GitHub Actions (vet, test -race, build).
├── main.go                        # Entry point. Echo + middlewares + timeouts + graceful shutdown.
├── main_test.go                   # Tests de arranque, middlewares, timeouts y graceful shutdown.
├── Dockerfile                     # Build multi-stage (golang:1.26-alpine → distroless:nonroot).
├── .dockerignore                  # Excluye .git, docs, odd, binarios y archivos no-runtime del build.
├── go.mod / go.sum                # Módulo `nutricontrol`.
├── PROJECT_INDEX.md               # ESTE archivo: navegación, comandos, estado.
├── AGENTS.md                      # Contexto operativo breve (arranque rápido).
├── README.md                      # Bitácora histórica día a día (no editar por sesión).
├── plan.md                        # Hoja de ruta operativa para próximas sesiones.
├── todo                           # Checklist ligero de la sesión actual.
├── handlers/                      # Capa presentation (HTTP).
│   ├── paciente.go                # Handlers + SetupRoutes (inyecta repo).
│   └── paciente_test.go           # Tests (verdes).
├── models/                        # Capa domain (entidades puras, sin deps).
│   ├── paciente.go                # struct Paciente.
│   └── paciente_test.go           # Test de serialización JSON.
├── pkg/
│   └── firestore/                 # Adaptador Firestore (puerto: PacienteRepository).
│       ├── paciente_repo.go       # Interfaz PacienteRepository.
│       ├── repo.go                # Impl Firestore real.
│       ├── mock.go                # Impl mock en memoria.
│       ├── mock_test.go           # Tests unitarios del mock.
│       └── repo_test.go           # Tests Firestore sobre gRPC in-process / Emulator.
├── docs/
│   ├── constitution.md            # Principios innegociables (hexagonal, testing, minimalismo).
│   ├── CHANGELOG.md               # Historia de cambios contractuales.
│   ├── arquitectura-y-defensa.md  # Decisiones de arquitectura y guía Q&A de defensa técnica.
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

Lee env vars: `PORT` (default `:8080`), `BODY_LIMIT` (default `"1M"`), `CORS_ALLOWED_ORIGINS` (opcional, separado por comas; desactivado por defecto), `RATE_LIMIT_RPS` / `RATE_LIMIT_BURST` (opcionales; desactivado por defecto si `RATE_LIMIT_RPS` falta o es `<= 0`), `API_KEY` (opcional; protege todas las rutas clínicas salvo `OPTIONS`, `GET /health` y `GET /`), `NUTRICONTROL_USE_MOCK=true`, `GCP_PROJECT_ID=<id>`.

- Si mock → `firestore.NewMockPacienteRepository()`.
- Si real → `firestore.NewPacienteFirestoreRepo(ctx, projectID)` (falla abiertamente si no hay credenciales GCP).
- Construye Echo vía `loadServerConfig()` + `newEchoWithConfig(repo, cfg)` (`strictJSONSerializer{}`, middlewares `RequestID` → `Logger` → `Recover` → `CORSWithConfig` opcional → `RateLimiterWithConfig` opcional → `KeyAuthWithConfig` opcional → `BodyLimit`, `handlers.SetupRoutes(e, repo)` y `GET /`).
- Arranca `*http.Server` (`newServerWithConfig`) con timeouts explícitos (`ReadHeaderTimeout: 5s`, `ReadTimeout: 15s`, `WriteTimeout: 15s`, `IdleTimeout: 60s`) y apagado ordenado (`runServer` / `serve`) ante `SIGINT`/`SIGTERM` con timeout de `10s`.

### 3.2 Rutas (contrato API)

| Método | Ruta | Handler | Repo método | Éxito |
|--------|------|---------|-------------|-------|
| GET | `/health` | HealthCheck | — (inspección de adaptador) | 200 OK `{"status": "ok", "storage": "mock"\|"firestore"}` |
| GET | `/pacientes` | GetPacientes | GetAll(ctx, limit, offset) | 200 OK `[Paciente]` |
| POST | `/pacientes` | CreatePaciente | Create | 201 Created `Paciente` (el mock asigna `ID` si viene vacío) |
| POST | `/pacientes/batch` | CreatePacientesBatch | Create (múltiple) | 201 Created `[Paciente]` con IDs asignados / 400 Bad Request si vacío o algún elemento inválido |
| GET | `/pacientes/{id}` | GetPacienteByID | GetByID | 200 OK `Paciente` / 404 |
| PUT | `/pacientes/{id}` | UpdatePaciente | Update | 200 OK `Paciente` / 404 / 400 |
| PATCH | `/pacientes/{id}` | PatchPaciente | GetByID + Update | 200 OK `Paciente` / 404 / 400 |
| DELETE | `/pacientes/{id}` | DeletePaciente | Delete (soft delete) | 204 No Content / 404 |
| GET | `/pacientes/inactivos` | GetPacientesInactivos | GetInactivos | 200 OK `[Paciente]` |
| POST | `/pacientes/{id}/restore` | RestorePaciente | Restore | 200 OK `Paciente` / 404 |
| GET | `/pacientes/{id}/controles` | GetControlesPaciente | GetByID + GetControles + filtro `?desde=&hasta=` y paginación `?limit=&offset=` | 200 OK `[ControlClinico]` / 400 / 404 |
| GET | `/pacientes/{id}/controles/resumen` | GetResumenControlesPaciente | GetByID + GetControles + filtro `?desde=&hasta=` | 200 OK `ResumenControles` / 400 / 404 |
| GET | `/pacientes/{id}/controles/export` | ExportControlesPacienteCSV | GetByID + GetControles + filtro `?desde=&hasta=` | 200 OK `text/csv` (`attachment; filename="controles.csv"`) / 400 / 404 |
| GET | `/pacientes/{id}/controles/ultimo` | GetUltimoControlPaciente | GetByID + GetControles | 200 OK `ControlClinico` / 404 |
| GET | `/pacientes/{id}/controles/{controlId}` | GetControlPacienteByID | GetByID + GetControlByID | 200 OK `ControlClinico` / 404 |
| POST | `/pacientes/{id}/controles` | CreateControlPaciente | AddControl | 201 Created `ControlClinico` / 400 / 404 |
| PUT | `/pacientes/{id}/controles/{controlId}` | UpdateControlPaciente | UpdateControl | 200 OK `ControlClinico` / 400 / 404 |
| PATCH | `/pacientes/{id}/controles/{controlId}` | PatchControlPaciente | GetByID + GetControlByID + UpdateControl | 200 OK `ControlClinico` / 400 / 404 |
| DELETE | `/pacientes/{id}/controles/{controlId}` | DeleteControlPaciente | DeleteControl | 204 No Content / 404 |
| GET | `/pacientes/search` | SearchPacientes | GetAll(ctx, limit, offset) + filtro en memoria | 200 OK `[Paciente]` paginado |
| GET | `/pacientes/stats` | StatsPacientes | GetAll(ctx, limit, offset) + agregación en memoria | 200 OK `{total, edad_promedio, peso_promedio, porcentaje_grasa_promedio, edades[], pesos[]}` |
| GET | `/pacientes/reporte` | ReportePacientes | GetAll(ctx, 0, 0) + agregación demográfica y nutricional | 200 OK `RespuestaReporte` `{total, grupos_etarios, clasificacion_imc, promedios}` |
| GET | `/pacientes/count` | CountPacientes | GetAll(ctx, 0, 0) | 200 OK `{"total": <n>}` |
| GET | `/pacientes/export` | ExportPacientesCSV | GetAll(ctx, 0, 0) | 200 OK `text/csv` (`attachment; filename="pacientes.csv"`) |
| GET | `/pacientes/edad/{edad}` | GetPacientesPorEdad | GetAll(ctx, 0, 0) + filtro en memoria por edad exacta | 200 OK `[Paciente]` (vacío si no hay coincidencia) / 400 si edad no numérica |
| GET | `/pacientes/imc/{id}` | GetPacienteIMC | GetByID | 200 OK `RespuestaIMC` `{id, nombre, peso, altura, imc, clasificacion}` / 404 / 400 si sin altura o peso inválido |
| GET | `/pacientes/{id}/imc` | GetPacienteIMC | GetByID | 200 OK `RespuestaIMC` (alias REST sub-recurso) / 404 / 400 |
| GET | `/` | (inline echo) | — | 200 "API NutriControl Operativa" |

- `{id}` proviene de `c.Param("id")`.
- Validaciones: `nombre` requerido; `edad` entre 1 y 120; `peso` mayor que 0 y menor o igual a 500; `porcentaje_grasa` entre 0 y 100; `altura` (opcional) entre 0.30 y 2.50 metros si se envía. 400 con `{"error": "..."}` si falta o es inválido.
- El nombre se normaliza antes de guardar: se eliminan espacios sobrantes y se capitaliza la primera letra de cada palabra (`"  juan perez  "` → `"Juan Perez"`). El nombre compuesto solo por espacios se rechaza con 400 (tanto en `POST`/`PUT` como en `PATCH`).

## 4. Contratos clave

### 4.1 Puerto `PacienteRepository` — `pkg/firestore/paciente_repo.go`

```go
var ErrControlNoEncontrado = errors.New("control clínico no encontrado")

type PacienteRepository interface {
    Create(ctx context.Context, p models.Paciente) (*models.Paciente, error)
    GetAll(ctx context.Context, limit, offset int) ([]models.Paciente, error)
    GetByID(ctx context.Context, id string) (*models.Paciente, error)
    Update(ctx context.Context, id string, p models.Paciente) (*models.Paciente, error)
    Delete(ctx context.Context, id string) error
    Restore(ctx context.Context, id string) (*models.Paciente, error)
    GetInactivos(ctx context.Context) ([]models.Paciente, error)
    AddControl(ctx context.Context, pacienteID string, control models.ControlClinico) (*models.ControlClinico, error)
    GetControles(ctx context.Context, pacienteID string) ([]models.ControlClinico, error)
    GetControlByID(ctx context.Context, pacienteID, controlID string) (*models.ControlClinico, error)
    UpdateControl(ctx context.Context, pacienteID, controlID string, control models.ControlClinico) (*models.ControlClinico, error)
    DeleteControl(ctx context.Context, pacienteID, controlID string) error
}
```

### 4.2 Adaptadores

- **Firestore real** (`pkg/firestore/repo.go`): `PacienteFirestoreRepo` (filtra `DeletedAt == nil` antes de paginar, usa `Set` para `Delete`/`Restore`, subcolección `pacientes/{id}/controles`, soporta `FIRESTORE_EMULATOR_HOST` e inyección vía `NewPacienteFirestoreRepoWithClient`). Tests gRPC in-process en `pkg/firestore/repo_test.go`.
- **Mock** (`pkg/firestore/mock.go`): `MockPacienteRepository`. Store en `[]models.Paciente` y `controles map[string][]models.ControlClinico` (no persiste entre arranques).

### 4.3 Modelo de dominio — `models/paciente.go`

```go
type Paciente struct {
    ID              string     `json:"id,omitempty"`
    Nombre          string     `json:"nombre"`
    Edad            int        `json:"edad"`
    Peso            float64    `json:"peso"`
    Altura          float64    `json:"altura,omitempty"`
    PorcentajeGrasa float64    `json:"porcentaje_grasa"`
    Notas           string     `json:"notas"`
    DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}

type ControlClinico struct {
    ID              string    `json:"id,omitempty"`
    PacienteID      string    `json:"paciente_id"`
    Fecha           time.Time `json:"fecha"`
    Peso            float64   `json:"peso"`
    PorcentajeGrasa float64   `json:"porcentaje_grasa"`
    IMC             float64   `json:"imc,omitempty"`
    Notas           string    `json:"notas,omitempty"`
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
docker build -t nutricontrol-api .        # Imagen multi-stage distroless:nonroot (~22 MB)
docker run -p 8080:8080 -e NUTRICONTROL_USE_MOCK=true nutricontrol-api
```

### 5.3 Testing
```bash
go test ./...                    # Todos los paquetes
go test -v ./handlers/           # Solo handlers
go test -v ./models/             # Solo models
go test -v ./pkg/firestore/      # Solo adaptador/mock
go test -cover ./...             # Con cobertura
go vet ./...                     # Verificación estándar
```

**Estado actual:** `go test ./...` verde (`handlers`, `models`, `pkg/firestore`).

### 5.4 Variables de entorno

| Variable | Obligatorio | Valor | Propósito |
|----------|-------------|-------|-----------|
| `PORT` | No | `8080` o `:8080` | Puerto de escucha HTTP (default `:8080`). |
| `BODY_LIMIT` | No | `1M`, `512K`, `2M` | Límite máximo de payload JSON/HTTP (`413`, default `1M`). |
| `CORS_ALLOWED_ORIGINS` | No | `http://localhost:3000,...` | Orígenes CORS permitidos separados por coma (desactivado si vacío). |
| `RATE_LIMIT_RPS` | No | `15`, `2.5` | Peticiones/segundo por IP (`RateLimiterMemoryStore`, TTL `3m`; desactivado si falta o `<= 0`). |
| `RATE_LIMIT_BURST` | No | `30` | Ráfaga máxima por IP cuando `RATE_LIMIT_RPS > 0` (default `max(1, ceil(RATE_LIMIT_RPS))`). |
| `API_KEY` | No | `<secreto>` | Clave Bearer / `X-API-Key` que habilita autenticación HTTP en `/pacientes*` (desactivado si falta). |
| `NUTRICONTROL_USE_MOCK` | No | `true` | Forzar mock en memoria. |
| `GCP_PROJECT_ID` | No | `<id>` | ID proyecto GCP para Firestore real. |
| `GOOGLE_APPLICATION_CREDENTIALS` | No | `/path/cred.json` | Credenciales de servicio GCP. |

## 6. Estado actual (ODD)

| Área | Estado |
|------|--------|
| Código fuente | Servidor HTTP endurecido (`main.go`: `PORT`, `BODY_LIMIT`, `CORS_ALLOWED_ORIGINS`, `RATE_LIMIT_RPS`/`RATE_LIMIT_BURST`, `API_KEY`, `strictJSONSerializer`, `RequestID`/`Logger`/`Recover`/`CORS`/`RateLimiter`/`KeyAuth`/`BodyLimit`, timeouts, `graceful shutdown`), CRUD completo de pacientes (`GET/POST/PUT/PATCH/DELETE` soft delete, `restore`, `inactivos`, `PatchPaciente` con validación estricta de tipos) y controles clínicos (`GET/POST /pacientes/:id/controles`, `GET /pacientes/:id/controles/ultimo`, `GET/PUT/PATCH/DELETE /pacientes/:id/controles/:controlId`, `resumen` con `?desde=&hasta=`, `export`), `GET /health`, IMC, reporte analítico, alta masiva, búsqueda con filtros/orden/paginación, `stats`, `count`, exportación CSV con `altura`, filtro por edad y validaciones clínicas. |
| Tests | `go test -cover -count=1 ./...` verde (128 tests top-level: `nutricontrol` 70.2%, `handlers` 88.3%, `models` [no statements], `pkg/firestore` 85.0%). |
| Build / CI | `go build ./...`, `go vet ./...` y contenedor Docker multi-stage (`golang:1.26-alpine` → `gcr.io/distroless/static-debian12:nonroot`, ~22 MB) verificados. Workflow CI `.github/workflows/ci.yml` en `actions/checkout@v7` + `actions/setup-go@v7` (Node.js 24) con `go test -race -cover -count=1 ./...` en verde (runs `45597881616`, `45799288074` y `45982386942` en `1m17s` sin avisos Node 20). |
| Docs | `PROJECT_INDEX.md` consolidado. `odd/task/estado-actual-proyecto.md` vigente. |
| Seguimiento ODD | `odd/task/estado-actual-proyecto.md` vigente. `feature-c4-search.md` y `feature-b-paginacion-firestore.md` archivados en `odd/tasks/archive/`. |
| Ignorados | `*.exe`, `*.dll`, `.atl/` vía `.gitignore` (commit `3f78f12`). |

### 6.1 Features C.2 y C.3 (completadas)

**C.2 — Mock asigna ID autoincremental al crear**
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

### 6.2 Feature C.4: `GET /pacientes/search` (búsqueda con filtros y paginación)

- `SearchPacientes` filtra por `nombre` (parcial, case-insensitive), `edad_min`/`edad_max`, `peso_min`/`peso_max`, y pagina con `limit` (default 10) / `offset` (default 0).
- Ruta `GET /pacientes/search` registrada en `SetupRoutes`.
- Tests: `TestSearchPacientesSinFiltros`, `TestSearchPacientesFiltrarPorNombre`, `TestSearchPacientesFiltrarPorEdadMinMax`, `TestSearchPacientesFiltrarPorPesoMinMax`, `TestSearchPacientesPaginacionLimitOffset`, `TestSearchPacientesValoresPorDefectoLimitOffset`, `TestSearchPacientesResultadoVacio`, `TestSearchPacientesCombinadoFiltros` (8 tests).
- Commit: `a4675bb feat: endpoint GET /pacientes/search con filtros y paginación (C.4)`.

### 6.3 Feature C.5: `GET /pacientes/stats` (resumen de solo lectura)

- `StatsPacientes` devuelve `{total, edad_promedio, peso_promedio, porcentaje_grasa_promedio, edades[], pesos[]}` agregando en memoria sobre `repo.GetAll`. No modifica datos ni toca contratos existentes.
- Ruta `GET /pacientes/stats` registrada en `SetupRoutes`.
- Tests: `TestStatsPacientesOK` y `TestStatsPacientesListaVacia`.
- Commit: `f06b5aa feat: endpoint GET /pacientes/stats con resumen de solo lectura (C.5)`.

### 6.4 Feature C.6: `PUT /pacientes/{id}` (actualizar paciente con validación)

- Se agregaron `TestUpdatePacienteOK` (200 con datos actualizados) y `TestUpdatePacienteNoEncontrado` (404 ID inexistente), completando la cobertura junto con `TestUpdatePacienteEdadInvalida` y `TestUpdatePacientePesoInvalido` (400).
- Commit: `85ed67f feat: completar C.6 PUT /pacientes/{id} con tests de éxito y no encontrado`.

### 6.5 Feature PATCH parcial: `PATCH /pacientes/{id}` (actualización parcial)

- `PatchPaciente` en `handlers/paciente.go`: aplica solo los campos enviados en el body (`nombre`, `edad`, `peso`, `porcentaje_grasa`, `notas`). Los campos no enviados se conservan.
- Requiere que el paciente exista (`GetByID`), si no existe devuelve 404.
- Tests: `TestPatchPacienteOK`, `TestPatchPacienteSoloPeso`, `TestPatchPacienteNoEncontrado`, `TestPatchPacienteEdadInvalida`.
- Commit: `9c6f605 feat: agregar PATCH /pacientes/{id} actualización parcial y archivar tracking C.4`.

### 6.6 Feature B: Paginación nativa `GetAll` con `limit`/`offset` en Firestore y mock

- `PacienteRepository.GetAll(ctx, limit, offset)` aplica `.Limit(limit)` y `.Offset(offset)` en Firestore real y recorta en memoria en `MockPacienteRepository`.
- Los handlers `GetPacientes`, `SearchPacientes` y `StatsPacientes` leen `limit` (default 10) y `offset` (default 0) mediante `paginacionDesdeQuery`.
- Tests: `TestGetPacientesPaginacionLimitOffset`, `TestGetPacientesOffsetSuperaTotal` y tests en `pkg/firestore/mock_test.go`.
- Commit: `d4f5c15 feat: paginación nativa GetAll con limit/offset en Firestore y mock`.

### 6.7 Validaciones de negocio extra y normalización de nombres (v1.0.5)

- `validarPaciente`: `edad` entre 1 y 120; `peso` mayor que 0 y menor o igual a 500.
- `normalizarNombre`: recorta espacios (`strings.TrimSpace`) y capitaliza cada palabra (`"  juan perez  "` → `"Juan Perez"`).
- Tests: `TestCreatePacienteEdadFueraDeRango`, `TestCreatePacientePesoFueraDeRango`, `TestCreatePacienteNombreConEspaciosSeNormaliza`, `TestCreatePacienteNombreSoloEspacios`.
- Commit: `f21a6a8 feat: validaciones de negocio extra y normalizar nombres en pacientes`.

### 6.8 Endpoint `GET /pacientes/edad/{edad}` (filtro por edad exacta, v1.0.6)

- `GetPacientesPorEdad` filtra sobre `repo.GetAll(ctx, 0, 0)` y devuelve los pacientes con `Edad == edad` (`200 OK` con lista vacía si no hay coincidencias; `400 Bad Request` si `edad` no es un entero válido).
- Tests: `TestGetPacientesPorEdadOK`, `TestGetPacientesPorEdadVacio`, `TestGetPacientesPorEdadInvalida`.
- Commit: `ae00eea feat: endpoint GET /pacientes/edad/{edad} filtro por edad exacta`.

### 6.9 Validación de `porcentaje_grasa` (0–100), `nombre` en PATCH y tests de `Delete`/`Edad` (v1.0.7)

- `validarPaciente` y `PatchPaciente` validan `porcentaje_grasa` en el rango `[0, 100]` (`porcentajeGrasaMaximo = 100`).
- `PatchPaciente` verifica que `nombre`, cuando se envía, no quede vacío tras `normalizarNombre` (rechaza `"   "` con `400 Bad Request`).
- Cobertura completada con `TestCreatePacientePorcentajeGrasaFueraDeRango`, `TestUpdatePacientePorcentajeGrasaFueraDeRango`, `TestPatchPacienteValidacionesExtra`, `TestDeletePacienteOK`, `TestDeletePacienteNoEncontrado` y `TestGetPacientesPorEdadInvalida`.
- Commits: `a0ad130` (saneamiento de `PROJECT_INDEX.md` y archivo de Feature B) y `0fd27a6` (validaciones y tests v1.0.7).

### 6.10 Endpoints `GET /health`, `GET /pacientes/count` y filtros `porcentaje_grasa_min`/`max` en `search` (v1.0.8)

- `HealthCheck` (`GET /health`): devuelve `200 OK` con `{"status": "ok", "storage": "mock"|"firestore"}` inspeccionando el tipo concreto de `PacienteRepository`.
- `CountPacientes` (`GET /pacientes/count`): devuelve `200 OK` con `{"total": <n>}` sobre `repo.GetAll(ctx, 0, 0)`.
- `SearchPacientes` (`GET /pacientes/search`): incorpora los filtros opcionales `porcentaje_grasa_min` y `porcentaje_grasa_max` (rango inclusivo en `float64`).
- En `SetupRoutes` se reordenaron todas las rutas estáticas (`/pacientes/search`, `/pacientes/stats`, `/pacientes/count`, `/pacientes/edad/{edad}`) antes de `/pacientes/{id}`.
- Tests: `TestHealthCheckOK`, `TestCountPacientesOK`, `TestCountPacientesVacio` y `TestSearchPacientesFiltrarPorPorcentajeGrasaMinMax`.
- Commit: `8d32dd6 feat: agregar GET /health, GET /pacientes/count y filtro porcentaje_grasa en search`.

### 6.11 Ordenamiento dinámico y exportación CSV (v1.0.9)

- `ordenarPacientes` en `handlers/paciente.go`: soporte de ordenamiento en memoria por `nombre`, `edad`, `peso` y `porcentaje_grasa` en sentidos `asc` y `desc`. Aplicado en `GetPacientes` (`GET /pacientes?sort=...&order=...`) y en `SearchPacientes` (`GET /pacientes/search?sort=...&order=...`).
- `ExportPacientesCSV` (`GET /pacientes/export`): exporta el padrón completo a CSV compatible con RFC 4180 (`id,nombre,edad,peso,porcentaje_grasa,notas`) forzando descarga con `Content-Disposition: attachment; filename="pacientes.csv"`. Registrado en `SetupRoutes` antes de `/pacientes/{id}`.
- Tests: `TestGetPacientesOrdenamiento`, `TestSearchPacientesOrdenamiento`, `TestExportPacientesCSVOK`, `TestExportPacientesCSVVacio` y `TestExportPacientesCSVRutaRegistrada`.

### 6.12 Feature: Alta masiva de pacientes `POST /pacientes/batch` (v1.0.10)

- Handler `CreatePacientesBatch` (`POST /pacientes/batch`): permite registrar un lote de pacientes en una única petición HTTP.
- **Validación atómica en dos fases**:
  1. Fase 1 (Validación y normalización): Si el lote está vacío (`[]`), se rechaza con `400 Bad Request` (`"El lote de pacientes no puede estar vacío"`). Cada paciente es normalizado (`normalizarNombre`) y validado con `validarPaciente`. Si alguno falla, la petición completa se rechaza con `400 Bad Request` indicando la posición del fallo (`"Paciente en posición N inválido: ..."`) sin guardar nada en base de datos.
  2. Fase 2 (Persistencia): Una vez asegurada la validez del 100% del lote, se itera sobre los registros invocando `repo.Create(ctx, p)` y se responde con `201 Created` conteniendo el slice de pacientes con sus IDs asignados automáticamente.
- Tests: `TestCreatePacientesBatchOK`, `TestCreatePacientesBatchVacio`, `TestCreatePacientesBatchAtomicidadInvalido`, `TestCreatePacientesBatchJSONInvalido` y `TestCreatePacientesBatchRutaRegistrada`.

### 6.13 Feature: Extensión de Dominio con Altura y cálculo de IMC (v1.0.11)

- Se incorporó el campo `Altura float64` (`json:"altura,omitempty"`) al modelo de dominio `models.Paciente`.
- Se implementaron validaciones de rango clínico `alturaMinima = 0.30` y `alturaMaxima = 2.50` metros en `validarPaciente` y `PatchPaciente`. Si `Altura != 0`, se exige que cumpla dicho rango, rechazando con 400 Bad Request si está fuera.
- Se amplió `ordenarPacientes` para permitir `sort=altura` ascendente y descendente.
- Handler `GetPacienteIMC`: calcula el IMC ($peso / altura^2$) redondeado a 2 decimales y devuelve la clasificación OMS (`Bajo peso`, `Normal`, `Sobrepeso`, `Obesidad`). Falla con 404 si el ID no existe o 400 si el paciente no tiene altura registrada (`<= 0`) o peso válido (`<= 0`).
- Registro dual en `SetupRoutes`: accesible vía `/pacientes/imc/{id}` y `/pacientes/{id}/imc`. Se corrigió además la sintaxis en el router registrando con `:id` y `:edad` para garantizar la resolución en llamadas HTTP reales.
- Tests agregados: `TestPacienteJSONSerializacionConAltura`, `TestGetPacienteIMCOK`, `TestGetPacienteIMCRutasRegistradas`, `TestGetPacienteIMCNoEncontrado`, `TestGetPacienteIMCSinAltura`, `TestGetPacienteIMCClasificaciones`, `TestCreatePacienteAlturaValida`, `TestCreatePacienteAlturaFueraDeRango`, `TestPatchPacienteAltura`, `TestPatchPacienteAlturaInvalida`.

### 6.14 Feature: Endpoint analítico `GET /pacientes/reporte` (v1.0.12)

- Handler `ReportePacientes` (`GET /pacientes/reporte`): consulta el padrón completo mediante `repo.GetAll(ctx, 0, 0)` y devuelve `RespuestaReporte` con:
  - `total`: cantidad total de pacientes registrados.
  - `grupos_etarios` (`GruposEtarios`): distribución demográfica en `menores` (`< 18`), `adultos` (`18–65`) y `mayores` (`> 65`).
  - `clasificacion_imc` (`ClasificacionIMCPoblacion`): distribución nutricional según la OMS reutilizando `clasificarIMC` (`bajo_peso`, `normal`, `sobrepeso`, `obesidad`) y `sin_datos` para pacientes sin altura o peso válido.
  - `promedios` (`PromediosReporte`): promedios redondeados a 2 decimales de `edad`, `peso`, `porcentaje_grasa` e `imc` (este último calculado solo sobre los pacientes con `Altura > 0` y `Peso > 0`).
- Ruta `GET /pacientes/reporte` registrada en `SetupRoutes` antes de `/pacientes/:id` y `/pacientes/{id}`.
- Tests: `TestReportePacientesOK`, `TestReportePacientesVacio` y `TestReportePacientesRutaRegistrada`.
- Commit: `79d9c53 feat: endpoint GET /pacientes/reporte para analítica poblacional`.

### 6.15 Feature: Soft delete en pacientes e historial de controles clínicos (v1.0.13)

- **Soft delete y restauración**:
  - `models.Paciente` suma `DeletedAt *time.Time` (`json:"deleted_at,omitempty"`).
  - `Delete` marca `DeletedAt` con `time.Now().UTC()`; `GetAll`, `GetByID` y `Update` excluyen automáticamente registros con `DeletedAt != nil`.
  - Nuevos métodos en `PacienteRepository` y rutas HTTP: `GET /pacientes/inactivos` (`GetPacientesInactivos`) y `POST /pacientes/:id/restore` (`RestorePaciente`).
- **Historial de controles clínicos**:
  - Nuevo modelo `models.ControlClinico` (`id`, `paciente_id`, `fecha`, `peso`, `porcentaje_grasa`, `imc`, `notas`).
  - Nuevos métodos `AddControl` y `GetControles` (subcolección `pacientes/{id}/controles` en Firestore, mapa `nil`-safe en mock), restringidos a pacientes activos y ordenados por `Fecha` ascendente.
  - Endpoints `POST /pacientes/:id/controles` (valida peso/grasa, asigna fecha UTC por defecto y calcula IMC automático si `Altura > 0`) y `GET /pacientes/:id/controles`.
- Tests en `models/paciente_test.go`, `pkg/firestore/mock_test.go` y `handlers/paciente_test.go` (85 ejecuciones `--- PASS:` incluyendo subtests).
- Commit: `be4cc9a feat: soft delete en pacientes e historial de controles clínicos`.

### 6.16 Feature: Cobertura gRPC de `PacienteFirestoreRepo` y resumen evolutivo clínico `GET /pacientes/:id/controles/resumen` (v1.0.14)

- **Cierre de cobertura en `pkg/firestore/` (`repo_test.go`, `repo.go`)**:
  - `NewPacienteFirestoreRepoWithClient` y detección de `FIRESTORE_EMULATOR_HOST` en `NewPacienteFirestoreRepo`.
  - `fakeFirestoreServer` in-process sobre `firestorepb` (`Commit`, `BatchGetDocuments`, `RunQuery`) para ejercitar `Create`, `GetAll`, `GetByID`, `Update`, `Delete`, `Restore`, `GetInactivos`, `AddControl` y `GetControles` a través del `*firestore.Client` real (81.9% de cobertura en `pkg/firestore`).
  - Commit: `0660bf5 test: cubrir PacienteFirestoreRepo sobre gRPC in-process y Emulator`.
- **Resumen evolutivo clínico (`GET /pacientes/:id/controles/resumen`)**:
  - Struct `ResumenControles`, helper `redondear2` y handler `GetResumenControlesPaciente` en `handlers/paciente.go`: devuelve `total_controles`, `fecha_inicio`/`fecha_fin`, valores iniciales y actuales más `delta_peso`, `delta_porcentaje_grasa` y `delta_imc` (calculado solo cuando primer y último control tienen `IMC > 0`).
  - Tests: `TestGetResumenControlesPacienteOK`, `TestGetResumenControlesPacienteSinControlesYUnControl` y `TestGetResumenControlesPacienteNoEncontradoYSoftDeleted` (91 ejecuciones `--- PASS:` totales).
  - Commit: `e220d1c feat: endpoint GET /pacientes/:id/controles/resumen evolutivo`.

### 6.17 Feature: Filtrado por rango de fechas y paginación en `GET /pacientes/:id/controles` (v1.0.15)

- Helper `parseFechaQuery(valor string, finDeDia bool)` en `handlers/paciente.go` con soporte dual `RFC3339` y `YYYY-MM-DD` (ajustando `hasta` al cierre del día UTC `23:59:59.999999999`).
- `GetControlesPaciente` valida `desde <= hasta` (`400 Bad Request` si falla formato o rango invertido), filtra cronológicamente de forma inclusiva y aplica `paginacionDesdeQuery` (`limit=10`, `offset=0` por defecto).
- Tests: `TestGetControlesPacienteFiltroFechas`, `TestGetControlesPacienteFiltroFechasInvalidas` y `TestGetControlesPacientePaginacion` (94 tests top-level en verde).
- Commit: `8f4cad4 feat: filtrado por fechas y paginación en GET /pacientes/:id/controles`.

### 6.18 Feature: Exportación CSV de controles clínicos y columna `altura` en `ExportPacientesCSV` (v1.0.16)

- Refactorización con helpers compartidos `obtenerControlesFiltrados` (compartido por `GetControlesPaciente` y `ExportControlesPacienteCSV`) y `escribirCSV` (escritura RFC 4180 compartida entre exportadores).
- `ExportPacientesCSV` (`GET /pacientes/export`): incorpora `altura` como 7.ª columna (`""` si no está registrada).
- `ExportControlesPacienteCSV` (`GET /pacientes/:id/controles/export`): exporta controles cronológicos (`id,paciente_id,fecha,peso,porcentaje_grasa,imc,notas`) con soporte de filtro `?desde=&hasta=` y cabecera `attachment; filename="controles.csv"`.
- Tests: actualizados `TestExportPacientesCSVOK` y `TestExportPacientesCSVVacio`; agregados `TestExportControlesPacienteCSVOK`, `TestExportControlesPacienteCSVFiltroFechasYVacios` y `TestExportControlesPacienteCSVNoEncontradoYSoftDeleted` (97 tests top-level en verde).
- Commit: `986116a feat: exportación CSV de controles clínicos y altura en ExportPacientesCSV`.

### 6.19 Feature: Edición/eliminación puntual de controles clínicos y pipeline CI (v1.0.17)

- **Edición y eliminación puntual de controles (`PUT/DELETE /pacientes/:id/controles/:controlId`)**:
  - `pkg/firestore/paciente_repo.go`: nuevo error centinela `ErrControlNoEncontrado` y métodos `UpdateControl` y `DeleteControl` en `PacienteRepository`.
  - `pkg/firestore/mock.go` y `pkg/firestore/repo.go`: paridad completa verificando paciente activo (`DeletedAt == nil`), preservando `Fecha` original si llega en cero en `PUT`, aislando controles por paciente y traduciendo `codes.NotFound`/`!doc.Exists()` a `ErrControlNoEncontrado` mediante `obtenerControlExistente`.
  - `handlers/paciente.go`: helpers compartidos `validarControlClinico` y `calcularIMCControl` (reutilizados entre `POST` y `PUT`), handlers `UpdateControlPaciente` (`200 OK` con recálculo de IMC) y `DeleteControlPaciente` (`204 No Content`), y registro en `SetupRoutes`.
  - Tests: `TestUpdateControlPacienteOK`, `TestUpdateControlPacienteErrores`, `TestDeleteControlPacienteOK`, `TestDeleteControlPacienteErrores` (`handlers/paciente_test.go`), `TestMockUpdateYDeleteControl` (`pkg/firestore/mock_test.go`) y `TestFirestoreRepoUpdateYDeleteControl` (`pkg/firestore/repo_test.go`) — 103 tests top-level en verde.
  - Commit: `027f1af feat: edición y eliminación de controles clínicos puntuales`.
- **Pipeline CI en GitHub Actions (`.github/workflows/ci.yml`)**:
  - Ejecuta `go vet ./...`, `go test -race -count=1 ./...` y `go build ./...` en `ubuntu-latest` ante `push`/`pull_request` sobre `main` con permisos `contents: read`.
  - Commit: `b3df81f ci: agregar workflow de GitHub Actions con vet, test -race y build`.

### 6.20 Feature: Hardening HTTP en `main.go` y consulta/edición parcial de controles `GET/PATCH` (v1.0.18)

- **Hardening HTTP y ciclo de vida (`main.go`, `main_test.go`)**:
  - `resolveAddr` lee `PORT` (default `:8080`), `newEcho` registra `RequestID` → `Logger` → `Recover` (`HideBanner = true`), `newServer` configura timeouts explícitos (`5s`/`15s`/`15s`/`60s`) y `serve`/`runServer` gestionan apagado ordenado ante `SIGINT`/`SIGTERM` con timeout de `10s`.
  - Tests: `TestResolveAddr`, `TestNewEchoMiddlewaresYRoutes`, `TestNewServerTimeouts` y `TestServeGracefulShutdown`.
  - Commit: `161afbf feat: hardening HTTP, PORT por env var, middlewares y graceful shutdown en main.go`.
- **Consulta puntual y actualización parcial de controles (`GET/PATCH /pacientes/:id/controles/:controlId`)**:
  - `GetControlByID` en `PacienteRepository`, `MockPacienteRepository` (`nil`-safe, copia defensiva) y `PacienteFirestoreRepo` (reutilizando `obtenerControlExistente`).
  - `GetControlPacienteByID` (`200 OK` / `404`) y `PatchControlPaciente` (actualiza solo campos presentes: `fecha`, `peso`, `porcentaje_grasa`, `notas`, con validación estricta de tipos JSON y recálculo de `IMC`), registrados en `SetupRoutes` después de `/resumen` y `/export`.
  - Tests: `TestGetControlPacienteByIDOK`, `TestGetControlPacienteByIDErrores`, `TestPatchControlPacienteOK`, `TestPatchControlPacienteErrores`, `TestMockGetControlByID` y `TestFirestoreRepoGetControlByID` (113 tests top-level en verde).
  - Commit: `db52369 feat: consulta puntual y actualización parcial de controles clínicos (GET/PATCH)`.
- **Verificación de CI en GitHub Actions**: corrida `44898033376` sobre `main` (`46e9d59`) completada en verde (`1m26s`, incluyendo `-race`), con advertencia de deprecación de Node.js 20 en `actions/checkout@v4` y `actions/setup-go@v5`.

### 6.21 Feature: CI v7 con Node.js 24 + cobertura y endpoint `GET /pacientes/:id/controles/ultimo` (v1.0.19)

- **Actualización de CI (`.github/workflows/ci.yml`)**:
  - Migrado de `actions/checkout@v4` y `actions/setup-go@v5` a `actions/checkout@v7` (`v7.0.1`) y `actions/setup-go@v7` (`v7.1.0`, runtime `node24` verificado vía `gh api`), e incorporado `-cover` en `go test -race -cover -count=1 ./...`.
  - Commit: `894c816 ci: actualizar checkout y setup-go a v7 (Node 24) y sumar cobertura`.
- **Consulta rápida del último control (`GET /pacientes/:id/controles/ultimo`)**:
  - Handler `GetUltimoControlPaciente` en `handlers/paciente.go`: valida `id` (`400`), paciente activo (`404`), obtiene historial cronológico (`repo.GetControles`), responde `404` si aún no existen controles y `200 OK` con `controles[len(controles)-1]`. Registrado en `SetupRoutes` antes de `:controlId`.
  - Tests: `TestGetUltimoControlPacienteOK` y `TestGetUltimoControlPacienteErrores` (`115` tests top-level en verde).
  - Commit: `c89e0f5 feat: endpoint GET /pacientes/:id/controles/ultimo`.

### 6.22 Feature: Hardening HTTP (`BodyLimit`, `CORS`), `strictJSONSerializer` y tipos estrictos en `PatchPaciente` (v1.0.20)

- **Hardening de entrada HTTP (`main.go`, `main_test.go`, `go.mod`)**:
  - Configurables vía env vars (`BODY_LIMIT` default `"1M"` validado con `gommon/bytes.Parse`, `CORS_ALLOWED_ORIGINS` separado por coma y desactivado por defecto con `AllowCredentials: false`).
  - Orden en `newEchoWithConfig`: `RequestID` → `Logger` → `Recover` → `CORSWithConfig` (opcional) → `BodyLimit`, con `strictJSONSerializer` (`io.ReadAll` + `json.Unmarshal`).
  - Tests: `TestResolveBodyLimit`, `TestParseCORSOrigins`, `TestLoadServerConfig`, `TestNewEchoBodyLimit` (streaming `ContentLength = -1` en los 7 endpoints de escritura), `TestNewEchoCORS` y `TestStrictJSONSerializerDeserializePropaga413`.
- **Endurecimiento de `PatchPaciente` y lectura JSON (`handlers/paciente.go`, `handlers/paciente_test.go`)**:
  - Helpers compartidos `manejarBindError`, `bindJSONStrict`, `leerBodyPatchJSON`, `extraerStringJSON` y `extraerFloatJSON`.
  - `PatchPaciente` rechaza `null`, decimales en `edad`, `NaN`/`Inf` y cualquier tipo erróneo en `nombre`, `edad`, `peso`, `porcentaje_grasa`, `altura` y `notas` (`400`, sin panic en `notas`).
  - Test: `TestPatchPacienteTiposInvalidos` (`122` tests top-level en verde; cobertura `59.1%` / `88.1%` / `85.0%`).
  - Commits: `32e92d3 fix: endurecer tipos JSON en PatchPaciente y lectura estricta bindJSONStrict` y `86e94d9 feat: agregar BodyLimit, CORS configurable y strictJSONSerializer en main.go`.

### 6.23 Feature: Filtro temporal en resumen de controles, verificación CI v7 y `Dockerfile` multi-stage (v1.0.21)

- **Filtrado `?desde=&hasta=` en `GET /pacientes/:id/controles/resumen` (`handlers/paciente.go`, `handlers/paciente_test.go`)**:
  - `GetResumenControlesPaciente` reutiliza `obtenerControlesFiltrados(c, repo)` para filtrar por fecha (`YYYY-MM-DD` inclusivo hasta fin de día UTC o `RFC3339`) manteniendo compatibilidad sin parámetros y devolviendo `200 OK` con `total_controles: 0` cuando el rango no tiene coincidencias, `400` ante formato inválido/rango invertido y `404` si el paciente no existe o fue eliminado.
  - Tests: `TestGetResumenControlesPacienteFiltroFechas` (5 subcasos) y `TestGetResumenControlesPacienteFiltroFechasInvalidas` (3 subcasos).
  - Commit: `92f26e3 feat: soportar filtro temporal desde/hasta en resumen de controles`.
- **Verificación CI v7 y contenerización (`Dockerfile`, `.dockerignore`)**:
  - Confirmada vía `gh run view 45597881616` la corrida en verde de `.github/workflows/ci.yml` (`bd622f1`, `1m24s` en Go `1.26.1` con `actions/checkout@v7`, `actions/setup-go@v7`, `-race`, `-cover` y cero avisos Node.js 20).
  - Agregados `Dockerfile` multi-stage (`golang:1.26-alpine`, `CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w"` → `gcr.io/distroless/static-debian12:nonroot`, `USER nonroot:nonroot`, `PORT=8080`, `ENTRYPOINT ["/nutricontrol"]`) y `.dockerignore`.
  - Validados suite Go (`124` tests top-level: `nutricontrol` `59.1%`, `handlers` `88.3%`, `models` `[no statements]`, `pkg/firestore` `85.0%`), build/smoke test Docker real (`22.049.372 bytes`, `/health`, `/`, `POST /pacientes` y cierre por `SIGTERM`) y auditoría independiente read-only `PASS`.
  - Commit: `714f97a build: agregar Dockerfile multi-stage y .dockerignore`.

### 6.24 Feature: Rate limiting opcional por IP en `main.go` y `plan.md` (v1.0.22)

- **Rate limiting opcional por IP (`main.go`, `main_test.go`, `go.mod`)**:
  - `resolveRateLimit` valida `RATE_LIMIT_RPS` y `RATE_LIMIT_BURST` sin `panic` (desactivado por defecto si falta, es `<= 0`, texto inválido o `NaN`/`Inf`).
  - `newEchoWithConfig` registra `middleware.RateLimiterWithConfig` + `middleware.NewRateLimiterMemoryStoreWithConfig` (`defaultRateLimitTTL = 3m`, `golang.org/x/time/rate`) entre `CORSWithConfig` y `BodyLimit`, eximiendo `OPTIONS` y `GET /health` (`c.Path() == "/health"`) y devolviendo `429 Too Many Requests` JSON (`{"error": ...}`) mediante `echo.NewHTTPError`.
  - `go.mod`: `golang.org/x/time v0.16.0` promovido a `require` directo vía `go mod tidy`.
  - Tests: `TestLoadServerConfig`, `TestResolveRateLimit` (12 subcasos) y `TestNewEchoRateLimit` (`126` tests top-level en verde: `nutricontrol` `67.3%`, `handlers` `88.3%`, `models` `[no statements]`, `pkg/firestore` `85.0%`; auditoría read-only `PASS`; CI `v1.0.21` run `45799288074` PASS en `39s`).
  - Commit: `76fdb3b feat: agregar rate limiting opcional por IP en main.go`.
- **Hoja de ruta operativa (`plan.md`)**: creada en raíz con el plan técnico para autenticación HTTP (`API_KEY` vs `JWT`), paginación con cursores `StartAfter` en Firestore (`GetPage` + `fakeFirestoreServer.RunQuery`) y checklist ODD.

### 6.25 Feature: Autenticación HTTP opcional por `API_KEY` (`KeyAuthWithConfig`) y verificación CI `v1.0.22` (v1.0.23)

- **Autenticación opcional por `API_KEY` (`main.go`, `main_test.go`)**:
  - `resolveAPIKey` normaliza `API_KEY` con `strings.TrimSpace` (si falta o solo contiene espacios desactiva auth sin `panic`, registrando únicamente `auth_enabled=%t` al iniciar).
  - `newEchoWithConfig`: inserta `middleware.KeyAuthWithConfig` entre `RateLimiterWithConfig` y `BodyLimit`, amplía `CORSConfig.AllowHeaders` con `"X-API-Key"`, acepta `Authorization: Bearer <key>` y `X-API-Key: <key>` validados con `crypto/subtle.ConstantTimeCompare`, exime solo `OPTIONS`, `GET /health` y `GET /` (protegiendo lectura y escritura clínica `/pacientes*`) y responde `401` JSON uniforme (`{"error": "No autorizado"}`) vía `echo.NewHTTPError`.
  - Tests: `TestLoadServerConfig`, `TestResolveAPIKey` y `TestNewEchoAPIKeyAuth` (`128` tests top-level en verde: `nutricontrol` `70.2%`, `handlers` `88.3%`, `models` `[no statements]`, `pkg/firestore` `85.0%`; auditoría read-only `PASS`; CI `v1.0.22` run `45982386942` PASS en `1m17s`).
  - Commit: `de5c6e5 feat: agregar autenticación opcional por API key en main.go`.

### 6.26 Dificultades acumuladas y decisiones de diseño

1. **`c.Response().Body.Bytes()` no compila en Echo v4**: Echo v4 eliminó `.Body` de `*echo.Response`. En los tests se lee el cuerpo mediante `c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes()` (commit `b3754f1`).
2. **Parámetros de ruta (`c.Param`) vacíos al invocar handlers directo en tests**: Echo v4 guarda `pnames` y `pvalues` en campos no exportados del contexto. Se creó el helper `setParam(c, key, value)` con `reflect` + `unsafe.Pointer` para inyectar parámetros en tests unitarios sin pasar por el router.
3. **`MockPacienteRepository.Create` sin ID**: el mock original no asignaba `ID` cuando venía vacío, impidiendo probar `GetByID`, `Update` o `Delete` sobre pacientes recién creados. Se incorporó el contador atómico `nextID` (`paciente-<n>`).
4. **Receptor de interfaz por valor vs puntero**: los handlers reciben `firestore.PacienteRepository` por valor (un puntero a interfaz `*firestore.PacienteRepository` no expone los métodos de la interfaz).
5. **Límites clínicos y formato de tipos**: se acotó `edad` a `1–120`, `peso` a `>0 y <=500`, `porcentaje_grasa` a `0–100` y `altura` a `0.30–2.50`. Las constantes enteras usan `%d` y flotantes `%.2f` en `fmt.Errorf`/`fmt.Sprintf` para pasar `go vet`.
6. **Prioridad de rutas en Echo v4**: `/pacientes/search`, `/pacientes/stats`, `/pacientes/count`, `/pacientes/export`, `/pacientes/edad/{edad}` y `/pacientes/imc/{id}` se registran en `SetupRoutes` antes que `/pacientes/{id}` para priorizar los segmentos estáticos frente al parámetro `{id}`.
7. **Simetría de invariantes en `PATCH` parcial**: como `PATCH` deserializa en `map[string]any` para no pisar campos omitidos, no puede invocar `validarPaciente` antes de fusionar cambios. Por eso aplica `normalizarNombre` y las cotas de rango campo por campo sobre los valores presentes antes de llamar a `repo.Update`.
8. **Recuperación de `PROJECT_INDEX.md` y separación de commits por unidad de trabajo**: ante un reemplazo accidental que había borrado las secciones `1` a `6.7` del índice en el working tree, se restauró desde `HEAD` en un commit documental independiente (`a0ad130`) antes de aplicar los cambios de código y tests (`0fd27a6`).
9. **Inspección de adaptador en `HealthCheck` sin acoplar `PacienteRepository`**: `HealthCheck` usa `_, isMock := repo.(*firestore.MockPacienteRepository)` para informar `"storage": "mock"|"firestore"` sin añadir métodos de infraestructura a la interfaz de dominio/puerto.
10. **Elección de `GET /pacientes/count` acorde al modelo de dominio**: dado que inicialmente `models.Paciente` no incluía `Altura`, se postergó el cálculo de IMC hasta v1.0.11 donde se incorporó el campo de forma retrocompatible.
11. **Verificación contra el seed real del mock en tests table-driven**: en la primera corrida de `TestSearchPacientesFiltrarPorPorcentajeGrasaMinMax` se detectó que en `NewMockPacienteRepository()` `Juan Perez` tiene `18.5%` de grasa y `Maria Gomez` tiene `22.0%` (al revés de lo supuesto inicialmente), ajustándose las aserciones al estado real del seed.
12. **Colisión estática vs paramétrica en `/pacientes/export` y test de integración router-level**: para asegurar que `GET /pacientes/export` no fuera capturado por `/pacientes/{id}` con `{id} = "export"`, se registró antes y se cubrió con `TestExportPacientesCSVRutaRegistrada` usando `e.ServeHTTP`. Además, se utilizó `encoding/csv.NewWriter` para respetar RFC 4180 y formateo decimal con `strconv.FormatFloat(val, 'f', -1, 64)`.
13. **Diagnóstico de configuración en relay de revisión RDD**: el proceso de revisión nativa identificó falta de modelo para la clave `review-readability` en la configuración global de Gentle AI (`models.json`). Bajo el contrato RDD, el fallo en transporte de lentes externas queda documentado como blocker de configuración y no afecta los contratos de compilación, tests locales ni entrega ordinaria de Git.
14. **Atomicidad de validación en operaciones batch y firma de `Create`**: para prevenir estados parciales o registros huérfanos sin necesidad de transacciones complejas distribuidas a nivel de base de datos, se aplicó un ciclo de validación estricto en dos fases (validar todo antes de persistir). Además, dado que `repo.Create` devuelve `*models.Paciente`, el acumulador de respuesta tipa con `[]*models.Paciente` para devolver limpiamente el slice serializado en JSON.
15. **Sintaxis de parámetros en el router de Echo v4 (`:id` vs `{id}`) y compatibilidad hacia atrás en `models.Paciente`**: Echo v4 espera parámetros de ruta con la sintaxis de dos puntos (`:id`, `:edad`), mientras que `{id}` era interpretado como literal estático por el motor radix cuando se pasaban peticiones HTTP completas a `e.ServeHTTP`. Se normalizó `SetupRoutes` registrando ambas sintaxis para máxima compatibilidad con tests unitarios y clientes HTTP reales. Por otro lado, la adición de `Altura` con `json:"altura,omitempty"` garantizó que toda la suite previa y registros sin altura sigan funcionando sin requerir migraciones de datos ni romper clientes existentes.
16. **Invariante contable con `sin_datos` y promedio condicional de IMC en `GET /pacientes/reporte`**: dado que `Altura` es opcional (`omitempty`), los registros previos sin altura (como los 2 pacientes del seed del mock) no permiten calcular el IMC. En `ReportePacientes` se agregó el contador explícito `sin_datos` dentro de `clasificacion_imc` para que la suma de categorías coincida siempre con `total`, y `promedios.imc` se divide exclusivamente entre los pacientes con IMC válido (`conIMC > 0`), reutilizando `clasificarIMC` tras redondear a 2 decimales para mantener consistencia exacta con `GET /pacientes/imc/:id`.
17. **Puntero `DeletedAt *time.Time` y paginación sobre activos tras soft delete**: usar `*time.Time` con `omitempty` en vez de `Activo bool` evitó que el valor cero (`false`) invalidara todos los literales de `models.Paciente` en tests existentes; además, `GetAll` filtra `DeletedAt == nil` antes de aplicar `limit`/`offset` tanto en el mock como en `PacienteFirestoreRepo` para no devolver páginas recortadas ni perder documentos históricos sin campo `deleted_at`.
18. **`nil`-safety en el mock y simetría de nombres en Firestore**: `MockPacienteRepository` inicializa `controles` de forma perezosa para soportar instanciaciones directas `&MockPacienteRepository{}` en tests existentes; en `PacienteFirestoreRepo`, `Delete` y `Restore` usan `docRef.Set` sobre el struct decodificado en lugar de `Update("deleted_at")` porque `models.Paciente` carece de tags `firestore:"..."`.
19. **Cierre de cobertura de `PacienteFirestoreRepo` mediante gRPC in-process y `FIRESTORE_EMULATOR_HOST`**: la brecha detectada en v1.0.13 (`Risk: item 6`) se resolvió con `fakeFirestoreServer` sobre `firestorepb` (`Commit`, `BatchGetDocuments`, `RunQuery`) conectado a `*firestore.Client` por loopback (`81.9%` de cobertura en `pkg/firestore`), y condicionando `option.WithEndpoint` en `NewPacienteFirestoreRepo` para respetar `FIRESTORE_EMULATOR_HOST`.
20. **Deltas evolutivos seguros y fechas opcionales en `GET /pacientes/:id/controles/resumen`**: `GetResumenControlesPaciente` usa `redondear2` para evitar deriva IEEE 754 en las restas, calcula `delta_imc` solo cuando primer y último control poseen `IMC > 0`, y emplea `*time.Time` con `omitempty` en `fecha_inicio`/`fecha_fin` para no serializar fechas cero cuando el paciente aún no tiene controles.
21. **Rango de fechas inclusivo `YYYY-MM-DD` y paginación posterior al filtrado en `GET /pacientes/:id/controles`**: `parseFechaQuery` extiende `hasta` en formato `YYYY-MM-DD` a `23:59:59.999999999 UTC` para no omitir controles del último día, rechaza explícitamente rangos invertidos con `400 Bad Request` y aplica `paginacionDesdeQuery` tras filtrar por fechas sin alterar el acceso directo al historial completo en `GetResumenControlesPaciente`.
22. **Columna `altura` al final en CSV y seguridad de cabeceras en `ExportControlesPacienteCSV`**: agregar `altura` como 7.ª columna en `ExportPacientesCSV` (emitiendo `""` cuando `Altura <= 0`) evita romper índices `0..5` existentes; extraer `obtenerControlesFiltrados` y `escribirCSV` elimina código duplicado, y usar el nombre estático `controles.csv` en `Content-Disposition` previene inyección de cabeceras vía parámetro `id`.
23. **Separación entre `404` de dominio y `500` de infraestructura en controles puntuales**: definir `ErrControlNoEncontrado` en `pkg/firestore` y traducir únicamente `codes.NotFound` o `!doc.Exists()` en `obtenerControlExistente` permite a `UpdateControlPaciente` y `DeleteControlPaciente` usar `errors.Is` para responder `404` sin ocultar fallos reales de red/Firestore tras falsos `404`.
24. **Preservación de fecha histórica en `PUT /pacientes/:id/controles/:controlId`**: cuando el cliente edita peso, grasa o notas omitiendo `fecha` (`control.Fecha.IsZero()`), ambos adaptadores conservan `existing.Fecha` en vez de pisarla con `0001-01-01` o `time.Now()`, evitando corromper el orden cronológico y `GetResumenControlesPaciente`.
25. **Limitación local de `-race` en Windows y validación en CI Linux**: `go test -race` requiere CGO/`gcc`, ausente en este entorno Windows; por ello `.github/workflows/ci.yml` corre en `ubuntu-latest` y `go.mod` fue actualizado con `go mod tidy` para declarar `google.golang.org/grpc` y `google.golang.org/protobuf` como dependencias directas.
26. **Bloqueo externo de Opción 3 (`review-readability` en `models.json`)**: se constató que `C:\Users\alvar\.pi\agent\models.json` solo configura `providers` (`strata`/`ollama`) y carece de claves de lentes RDD; al pertenecer a configuración externa del host fuera del repositorio, queda documentado como ajuste manual del entorno RDD.
27. **Testabilidad determinista del `graceful shutdown` en `main.go`**: separar `newEcho`, `newServer`, `serve(ctx, srv, ln, timeout)` y `runServer` permitió probar peticiones reales y apagado ordenado sobre `127.0.0.1:0` drenando `errCh` (buffer `1`) tras `srv.Shutdown`, sin puertos fijos ni fugas de goroutines (`50/50` iteraciones en verde).
28. **Precedencia estática de `/resumen` y `/export` frente a `GET /pacientes/:id/controles/:controlId`**: al compartir método `GET` y profundidad de segmentos, la ruta paramétrica `:controlId` se registró explícitamente después de `/resumen` y `/export` y quedó cubierta en `TestGetControlPacienteByIDOK`.
29. **Validación estricta de tipos JSON en `PatchControlPaciente`**: al deserializar en `map[string]any`, validar cada campo con type assertion seguro (`ok`) evita panics ante tipos inesperados y devuelve `400 Bad Request` explícito; también dejó en evidencia el type assertion directo `notas.(string)` pendiente de endurecer en `PatchPaciente`.
30. **Advertencia de Node.js 20 en GitHub Actions**: la primera corrida verde de CI (`44898033376`, `1m26s`) confirmó `go test -race` en Linux y mostró un aviso de deprecación de Node.js 20 en `actions/checkout@v4` y `actions/setup-go@v5`.
31. **Colisión de Echo v4 entre `/controles/ultimo` y `/controles/:controlId` en RED**: antes de registrar la ruta estática `/ultimo` antes de `/:controlId`, Echo enrutaba `"ultimo"` como `:controlId` devolviendo `404` de `GetControlPacienteByID`; registrar `/ultimo` justo tras `/export` elimina la colisión.
32. **`404 Not Found` ante paciente activo sin controles en `/controles/ultimo`**: devolver un struct cero en `200 OK` crearía registros sintéticos inválidos; retornar `404` explícito mantiene coherencia REST con `GET /controles/:controlId`.
33. **Verificación de `node24` en `checkout@v7` y `setup-go@v7` vía `gh api`**: se comprobaron las etiquetas reales (`v7.0.1`, `v7.1.0`) y sus `action.yml` antes de actualizar `.github/workflows/ci.yml`, preservando `go-version-file: go.mod` y añadiendo `-cover`.
34. **Propagación de `413` en streaming (`strictJSONSerializer` y `bindJSONStrict`) entre Go `1.26` y `1.27`**: en `encoding/json` v1 (Go `<= 1.26`, versión en `go.mod` y CI), `json.Decoder` escanea el buffer antes de revisar errores de lectura, pudiendo descartar `echo.ErrStatusRequestEntityTooLarge` si el JSON cierra en el mismo `Read` (enmascarado localmente bajo Go `1.27.2` con `json/v2`). Reemplazar la decodificación streaming por `io.ReadAll` + `json.Unmarshal` y probar `partialLimitReader` garantiza `413` en ambas versiones.
35. **Validación preventiva de `BODY_LIMIT` y CORS seguro por defecto**: validar `BODY_LIMIT` con `bytes.Parse` evita el `panic` de `middleware.BodyLimit` ante valores inválidos o `<= 0`, mientras que omitir `CORSWithConfig` cuando `CORS_ALLOWED_ORIGINS` está vacío evita el wildcard `*` predeterminado de Echo.
36. **Endurecimiento uniforme de `PatchPaciente` con `extraerStringJSON` y `extraerFloatJSON`**: reemplaza el type assertion inseguro `notas.(string)`, rechaza `null`, decimales en `edad` y tipos erróneos que antes devolvían `200 OK` sin aplicar cambios.
37. **Reutilización directa de `obtenerControlesFiltrados` en `GetResumenControlesPaciente`**: elimina la duplicación de validaciones (`id`, `repo.GetByID`, `repo.GetControles`) unificando el soporte de `?desde=&hasta=`, la precedencia `404` antes que `400` y el ajuste inclusivo hasta fin de día UTC sin introducir paginación en el resumen.
38. **Semántica `200 OK` con `total_controles: 0` ante rangos vacíos en `/controles/resumen`**: a diferencia de `/controles/ultimo` (recurso puntual que responde `404`), `/resumen` mantiene `200 OK` omitiendo `fecha_inicio`/`fecha_fin` (`omitempty`), coherente con el estado inicial de un paciente sin controles.
39. **`gcr.io/distroless/static-debian12:nonroot`, `ENTRYPOINT` exec y `.dockerignore`**: frente a `scratch`, `distroless:nonroot` incluye los certificados CA requeridos por Cloud Firestore gRPC/TLS y ejecución sin privilegios `root` (~22 MB); el formato exec entrega `SIGTERM` directo a `signal.NotifyContext` en `main.go`, y `.dockerignore` evita enviar `.git/`, docs o binarios invalidando la caché de `go mod download`.
40. **Orden `CORS` → `RateLimiter` → `BodyLimit` y `c.Path()` en `Skipper`**: registrar `CORSWithConfig` antes de `RateLimiterWithConfig` conserva `Access-Control-Allow-Origin` en respuestas `429`, poner `BodyLimit` detrás evita leer el body en peticiones rechazadas, y como `e.Use` corre tras `Router.Find`, `c.Path() == "/health"` exime el healthcheck sin abrir rutas inexistentes.
41. **`echo.NewHTTPError(429, map[string]string{...})` frente a `c.JSON` directo en `DenyHandler`**: `RateLimiterWithConfig` invoca `c.Error(config.DenyHandler(...))`; retornar un `*echo.HTTPError` con mapa evita el envoltorio `{"message": ...}` de `DefaultHTTPErrorHandler`, mantiene el formato `{"error": ...}`, registra el error en `Logger` y emite `NoContent(429)` ante `HEAD`.
42. **Supuestos operacionales de `RateLimiterMemoryStore` y `c.RealIP()`**: mantener `RATE_LIMIT_RPS` opt-in reutiliza `golang.org/x/time/rate` sin nuevas dependencias, aunque `c.RealIP()` asume proxy confiable para `X-Forwarded-For` si no se fija `e.IPExtractor`, las cuotas en memoria (`TTL 3m`) son por réplica, no se emite `Retry-After` y valores astronómicos (> `math.MaxInt`) desbordarían `int(math.Ceil(rps))`.
43. **`ErrorHandler` `401` frente al `400 Bad Request` por defecto de Echo `KeyAuthWithConfig`**: sin `ErrorHandler`, la falta de cabecera (`ErrKeyAuthMissing`) devuelve `400` y una clave errónea `401`; retornar `echo.NewHTTPError(http.StatusUnauthorized, map[string]string{"error": "No autorizado"})` unifica todo fallo en `401` JSON uniforme, conserva cabeceras CORS y responde `NoContent(401)` en `HEAD`.
44. **`createExtractors` en `KeyLookup` y `"X-API-Key"` en `CORSConfig.AllowHeaders`**: `"header:Authorization,header:X-API-Key"` con `AuthScheme: "Bearer"` recorta `"Bearer "` de forma case-insensitive únicamente en `Authorization` (rechazando `"Bearer "` vacío) y preserva íntegro `X-API-Key`; incluir `"X-API-Key"` en `CORSConfig.AllowHeaders` evita bloqueos en preflight `OPTIONS`.
45. **Protección total de `/pacientes*`, orden `RateLimit` → `KeyAuth` → `BodyLimit` y supuestos operacionales**: eximir en `Skipper` solo `OPTIONS`, `GET /health` y `GET /` resguarda datos clínicos sensibles (protegiendo también `/health/`, mayúsculas y `POST /health` tras `Router.Find`), frena fuerza bruta con `429` antes de validar claves y rechaza `401` antes de leer el cuerpo en `BodyLimit`; requiere TLS externo (Cloud Run), clave de alta entropía (>= 32 bytes) y tiene en cuenta que `ConstantTimeCompare` corta antes si difiere la longitud.

## 7. Convenciones de entrega

- Estilo: **Conventional Commits** (`feat:`, `fix:`, `refactor:`, `docs:`, `build:`, `chore:`).
- Antes de commitear: `go mod tidy`, `go build ./...`, `go vet ./...`, `go test ./...`.
