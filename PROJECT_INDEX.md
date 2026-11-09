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
│   ├── paciente.go                # struct Paciente.
│   └── paciente_test.go           # Test de serialización JSON.
├── pkg/
│   └── firestore/                 # Adaptador Firestore (puerto: PacienteRepository).
│       ├── paciente_repo.go       # Interfaz PacienteRepository.
│       ├── repo.go                # Impl Firestore real.
│       ├── mock.go                # Impl mock en memoria.
│       └── mock_test.go           # Tests unitarios del mock.
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
| GET | `/pacientes` | GetPacientes | GetAll(ctx, limit, offset) | 200 OK `[Paciente]` |
| POST | `/pacientes` | CreatePaciente | Create | 201 Created `Paciente` (el mock asigna `ID` si viene vacío) |
| GET | `/pacientes/{id}` | GetPacienteByID | GetByID | 200 OK `Paciente` / 404 |
| PUT | `/pacientes/{id}` | UpdatePaciente | Update | 200 OK `Paciente` / 404 / 400 |
| PATCH | `/pacientes/{id}` | PatchPaciente | GetByID + Update | 200 OK `Paciente` / 404 / 400 |
| DELETE | `/pacientes/{id}` | DeletePaciente | Delete | 204 No Content / 404 |
| GET | `/pacientes/search` | SearchPacientes | GetAll(ctx, limit, offset) + filtro en memoria | 200 OK `[Paciente]` paginado |
| GET | `/pacientes/stats` | StatsPacientes | GetAll(ctx, limit, offset) + agregación en memoria | 200 OK `{total, edad_promedio, peso_promedio, porcentaje_grasa_promedio, edades[], pesos[]}` |
| GET | `/pacientes/edad/{edad}` | GetPacientesPorEdad | GetAll(ctx, 0, 0) + filtro en memoria por edad exacta | 200 OK `[Paciente]` (vacío si no hay coincidencia) / 400 si edad no numérica |
| GET | `/` | (inline echo) | — | 200 "API NutriControl Operativa" |

- `{id}` proviene de `c.Param("id")`.
- Validaciones: `nombre` requerido; `edad` entre 1 y 120; `peso` mayor que 0 y menor o igual a 500; `porcentaje_grasa` entre 0 y 100. 400 con `{"error": "..."}` si falta o es inválido.
- El nombre se normaliza antes de guardar: se eliminan espacios sobrantes y se capitaliza la primera letra de cada palabra (`"  juan perez  "` → `"Juan Perez"`). El nombre compuesto solo por espacios se rechaza con 400 (tanto en `POST`/`PUT` como en `PATCH`).

## 4. Contratos clave

### 4.1 Puerto `PacienteRepository` — `pkg/firestore/paciente_repo.go`

```go
type PacienteRepository interface {
    Create(ctx context.Context, p models.Paciente) (*models.Paciente, error)
    GetAll(ctx context.Context, limit, offset int) ([]models.Paciente, error)
    GetByID(ctx context.Context, id string) (*models.Paciente, error)
    Update(ctx context.Context, id string, p models.Paciente) (*models.Paciente, error)
    Delete(ctx context.Context, id string) error
}
```

### 4.2 Adaptadores

- **Firestore real** (`pkg/firestore/repo.go`): `PacienteFirestoreRepo` (aplica `.Limit(limit)` y `.Offset(offset)` cuando son mayores que 0).
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
go test -v ./pkg/firestore/      # Solo adaptador/mock
go test -cover ./...             # Con cobertura
go vet ./...                     # Verificación estándar
```

**Estado actual:** `go test ./...` verde (`handlers`, `models`, `pkg/firestore`).

### 5.4 Variables de entorno

| Variable | Obligatorio | Valor | Propósito |
|----------|-------------|-------|-----------|
| `NUTRICONTROL_USE_MOCK` | No | `true` | Forzar mock en memoria. |
| `GCP_PROJECT_ID` | No | `<id>` | ID proyecto GCP para Firestore real. |
| `GOOGLE_APPLICATION_CREDENTIALS` | No | `/path/cred.json` | Credenciales de servicio GCP. |

## 6. Estado actual (ODD)

| Área | Estado |
|------|--------|
| Código fuente | CRUD completo (`GET/POST/PUT/PATCH/DELETE` por ID), filtro por nombre, `search` con paginación nativa `limit`/`offset`, `stats`, `GET /pacientes/edad/{edad}`, validaciones de negocio (edad 1-120, peso 0-500) y normalización de nombres. |
| Tests | `go test ./...` verde (`handlers`, `models`, `pkg/firestore`). |
| Build | `go build ./...` y `go vet ./...` OK. |
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

### 6.10 Dificultades acumuladas y decisiones de diseño

1. **`c.Response().Body.Bytes()` no compila en Echo v4**: Echo v4 eliminó `.Body` de `*echo.Response`. En los tests se lee el cuerpo mediante `c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes()` (commit `b3754f1`).
2. **Parámetros de ruta (`c.Param`) vacíos al invocar handlers directo en tests**: Echo v4 guarda `pnames` y `pvalues` en campos no exportados del contexto. Se creó el helper `setParam(c, key, value)` con `reflect` + `unsafe.Pointer` para inyectar parámetros en tests unitarios sin pasar por el router.
3. **`MockPacienteRepository.Create` sin ID**: el mock original no asignaba `ID` cuando venía vacío, impidiendo probar `GetByID`, `Update` o `Delete` sobre pacientes recién creados. Se incorporó el contador atómico `nextID` (`paciente-<n>`).
4. **Receptor de interfaz por valor vs puntero**: los handlers reciben `firestore.PacienteRepository` por valor (un puntero a interfaz `*firestore.PacienteRepository` no expone los métodos de la interfaz).
5. **Límites clínicos y formato de tipos**: se acotó `edad` a `1–120`, `peso` a `>0 y <=500` y `porcentaje_grasa` a `0–100`. Las constantes enteras usan `%d` en `fmt.Errorf`/`fmt.Sprintf` para pasar `go vet`.
6. **Prioridad de rutas en Echo v4**: `/pacientes/edad/{edad}` comparte prefijo con `/pacientes/{id}`; Echo prioriza el segmento estático `edad` sobre el parámetro `{id}`.
7. **Simetría de invariantes en `PATCH` parcial**: como `PATCH` deserializa en `map[string]any` para no pisar campos omitidos, no puede invocar `validarPaciente` antes de fusionar cambios. Por eso aplica `normalizarNombre` y las cotas de rango campo por campo sobre los valores presentes antes de llamar a `repo.Update`.
8. **Recuperación de `PROJECT_INDEX.md` y separación de commits por unidad de trabajo**: ante un reemplazo accidental que había borrado las secciones `1` a `6.7` del índice en el working tree, se restauró desde `HEAD` en un commit documental independiente (`a0ad130`) antes de aplicar los cambios de código y tests (`0fd27a6`).

## 7. Convenciones de entrega

- Estilo: **Conventional Commits** (`feat:`, `fix:`, `refactor:`, `docs:`, `build:`, `chore:`).
- Antes de commitear: `go mod tidy`, `go build ./...`, `go vet ./...`, `go test ./...`.
