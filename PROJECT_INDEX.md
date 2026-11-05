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
| PATCH | `/pacientes/{id}` | PatchPaciente | GetByID + Update | 200 OK `Paciente` / 404 / 400 |
| DELETE | `/pacientes/{id}` | DeletePaciente | Delete | 204 No Content / 404 |
| GET | `/pacientes/search` | SearchPacientes | GetAll + filtro en memoria | 200 OK `[Paciente]` paginado |
| GET | `/pacientes/stats` | StatsPacientes | GetAll + agregación en memoria | 200 OK `{total, edad_promedio, peso_promedio, porcentaje_grasa_promedio, edades[], pesos[]}` |
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

## 6. Estado actual (post C.4)

| Área | Estado |
|------|--------|
| Código fuente | Sin cambios pendientes en la rama. `main` en `c0fc50a` (sync docs ODD) sobre `9c6f605` feat: PATCH /pacientes/{id} actualización parcial. CRUD completo: GET/POST/PUT/PATCH/DELETE por ID, filtro nombre, search, stats. |
| Tests | `go test ./...` verde (`handlers`, `models`, `pkg/firestore`). Incluye 4 tests de PATCH parcial. |
| Build | `go build ./...` y `go vet ./...` OK. |
| Docs | `PROJECT_INDEX.md` consolidado. `odd/task/estado-actual-proyecto.md` vigente. |
| Seguimiento ODD | `odd/task/estado-actual-proyecto.md` vigente (único archivo de definición de estado actual). `odd/tasks/feature-c4-search.md` archivado en `odd/tasks/archive/`. |
| Ignorados | `*.exe`, `*.dll`, `.atl/` vía `.gitignore` (commit `3f78f12`). |
| Pendiente | Push de commits pendientes a origin/main (verificar estado remoto). |

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

### 6.2 Dificultades encontradas y cómo las resolví (movidas al final, sección 6.6)

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

### 6.3 Feature C.4: `GET /pacientes/search` (búsqueda con filtros y paginación)

- `SearchPacientes` filtra por `nombre` (parcial, case-insensitive), `edad_min`/`edad_max`, `peso_min`/`peso_max`, y pagina con `limit` (default 10) / `offset` (default 0). Usa `repo.GetAll` y filtra en memoria, igual que `GetPacientes`, sin tocar la interfaz `PacienteRepository`.
- Ruta `GET /pacientes/search` registrada en `SetupRoutes`.
- Tests: `TestSearchPacientesSinFiltros`, `TestSearchPacientesFiltrarPorNombre`, `TestSearchPacientesFiltrarPorEdadMinMax`, `TestSearchPacientesFiltrarPorPesoMinMax`, `TestSearchPacientesPaginacionLimitOffset`, `TestSearchPacientesValoresPorDefectoLimitOffset`, `TestSearchPacientesResultadoVacio`, `TestSearchPacientesCombinadoFiltros` (8 tests).
- Commit: `a4675bb feat: endpoint GET /pacientes/search con filtros y paginación (C.4)`. Pusheado a `origin/main`.

### 6.4 Feature C.5: `GET /pacientes/stats` (resumen de solo lectura)

- `StatsPacientes` devuelve `{total, edad_promedio, peso_promedio, porcentaje_grasa_promedio, edades[], pesos[]}` agregando en memoria sobre `repo.GetAll`. No modifica datos ni toca contratos existentes.
- Ruta `GET /pacientes/stats` registrada en `SetupRoutes`.
- Tests: `TestStatsPacientesOK` (2 pacientes del mock: total 2, edad promedio 27.5, peso 68.75, grasa 20.25) y `TestStatsPacientesListaVacia` (repositorio vacío → ceros y listas vacías).
- **Dificultad encontrada:** ninguna nueva. Se reutilizan los mismos patrones de test que C.4 (`httptest.NewRequest`, `c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes()`).
- Commit: `f06b5aa feat: endpoint GET /pacientes/stats con resumen de solo lectura (C.5)`. Pusheado a `origin/main`.

### 6.5 Feature C.6: `PUT /pacientes/{id}` (actualizar paciente con validación)

**¿Qué se hizo y por qué:**
El handler `UpdatePaciente` y la ruta `PUT /pacientes/{id}` ya existían desde la versión 1.0.0 (commit `8a1a47f`), con validación de ID y delegación al repositorio `Update`. Lo que faltaba para cerrar C.6 eran los **tests de éxito (200) y no encontrado (404)**. Se agregaron porque sin ellos el endpoint no tenía cobertura de su caso principal: actualizar un paciente existente y devolverlo con los datos nuevos, ni el caso de error cuando el ID no existe.

**Cambios realizados:**
- `handlers/paciente_test.go`: agregados `TestUpdatePacienteOK` y `TestUpdatePacienteNoEncontrado`.
  - `TestUpdatePacienteOK`: crea un paciente con `repo.Create` (el mock le asigna `paciente-101`), luego llama `UpdatePaciente` con payload nuevo y verifica `200 OK`, que `ID` se conserva y que `nombre`, `edad`, `peso`, `porcentaje_grasa` y `notas` se actualizan.
  - `TestUpdatePacienteNoEncontrado`: llama `UpdatePaciente` con ID `inexistente` y verifica `404 Not Found` con mensaje `"No se encontró paciente"`.
- Los tests preexistentes `TestUpdatePacienteEdadInvalida` y `TestUpdatePacientePesoInvalido` cubren la validación de negocio (`edad <= 0` y `peso <= 0` → `400 Bad Request`).

**¿Por qué se hace así (diseño):**
- El mock `Update` modifica el paciente en `m.store` por índice y lo devuelve, así que el handler solo necesita pasar el ID de `c.Param("id")` y el cuerpo parseado. No hay lógica extra en el handler: valida, delega al repo, devuelve el resultado.
- El helper `setParam` (reflect + unsafe.Pointer) es necesario porque los tests invocan el handler directamente, sin pasar por el router de Echo, por lo que `c.Param("id")` estaría vacío sin esta inyección.

**Tests de C.6:**
- `TestUpdatePacienteOK` (200, datos actualizados)
- `TestUpdatePacienteNoEncontrado` (404, ID inexistente)
- `TestUpdatePacienteEdadInvalida` (400, edad <= 0)
- `TestUpdatePacientePesoInvalido` (400, peso <= 0)

**Commit:** `85ed67f feat: completar C.6 PUT /pacientes/{id} con tests de éxito y no encontrado` (commit local, aún no pusheado a `origin/main`).

### 6.6 Dificultades encontradas y cómo las resolví

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
- **Resolución:** Agregué contador atómico `nextID` en `MockPacienteRepository` y lógica en `Create` para asignar `paciente-<n>` cuando `p.ID == ""`, respetando el ID si ya viene seteado.

**4. go.sum grande tras agregar `cloud.google.com/go/firestore`**
- **Síntoma:** `go.sum` crecía ~30 líneas la primera vez que se resolvía la dependencia de Firestore real.
- **Resolución:** Ejecutar `go mod tidy` antes de commitear para que `go.mod` y `go.sum` queden consistentes.

### 6.7 Feature PATCH parcial: `PATCH /pacientes/{id}` (actualización parcial) — completada

- `PatchPaciente` en `handlers/paciente.go`: aplica solo los campos enviados en el body (nombre, edad, peso, porcentaje_grasa, notas). Los campos no enviados se conservan.
- Requiere que el paciente exista (`GetByID`), si no existe devuelve 404.
- Valida `edad > 0` y `peso > 0` solo cuando esos campos se envían; 400 si inválidos.
- Ruta `PATCH /pacientes/{id}` registrada en `SetupRoutes`.
- Tests: `TestPatchPacienteOK` (200, solo nombre cambia), `TestPatchPacienteSoloPeso` (200, solo peso cambia), `TestPatchPacienteNoEncontrado` (404), `TestPatchPacienteEdadInvalida` (400).
- Commit: `9c6f605 feat: agregar PATCH /pacientes/{id} actualización parcial y archivar tracking C.4` + sincronización de estado en `c0fc50a`. Pusheado a `origin/main`.

- Estilo: **Conventional Commits** (`feat:`, `fix:`, `refactor:`, `docs:`, `build:`, `chore:`).
- Rama de trabajo: crear feature branch desde `main` antes del primer cambio de fuente.
- Antes de commitear: `go mod tidy`, `go build ./...`, `go test ./...`.
- Push y creación de PR quedan a criterio del usuario bajo política ordinaria del repositorio.