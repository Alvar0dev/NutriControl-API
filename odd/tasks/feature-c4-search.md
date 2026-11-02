# Feature C.4 — Search múltiple con paginación (GET /pacientes/search)

## Header
- Objetivo: endpoint que permite buscar pacientes por múltiples criterios (nombre parcial, rango de edad, rango de peso) con paginación simple (limit/offset).
- Branch: `main` (work-unit inline, commit directo por autorización del usuario).
- Delivery strategy: `ask-on-risk` (default).
- Test runner: `go test ./...`

## Specs

S1 (usuario): "Implementar un endpoint GET /pacientes/search que permita buscar por nombre (coincidencia parcial, sin distinguir mayúsculas), rango de edad (edad_min, edad_max) y rango de peso (peso_min, peso_max), con paginación por limit y offset. Responder con la lista filtrada y paginada."

S2 (contrato de comportamiento): El endpoint debe ser independiente de `GetPacientes`; debe usar el repositorio `GetAll` (no modificar la interfaz `PacienteRepository`, que ya está establecida) y filtrar en memoria, igual que `GetPacientes`. No debe romper `GetPacientes` existente. Si no hay filtros, debe devolver todos los resultados paginados. Si `limit` o `offset` faltan, usar valores por defecto (`limit=10`, `offset=0`).

S3 (tests): Al menos un test RED que falle antes de implementar (busqueda sin filtros, búsqueda con nombre parcial, rango de edad, rango de peso, paginación, valores por defecto, y caso vacío). Todos los tests deben pasar (`go test ./...` verde) antes del commit.

## Tasks

| ID | S# | Route | Status | Evidence / Commit |
|----|----|-------|--------|-------------------|
| T1 | S1 | inline | done | Diseño del endpoint y filtros |
| T2 | S2 | inline | done | Implementación del handler `SearchPacientes` + registro en `SetupRoutes` |
| T3 | S3 | inline | done | Tests RED → GREEN para search (nombre, edad, peso, paginación, vacío, por defecto) |
| T4 | S3 | inline | done | Documentar dificultades encontradas (PROJECT_INDEX.md sección 6.2) |
| T5 | S1 | inline | done | Commit `a4675bb` + push a origin/main |

## Log

L1 (usuario, hoy): "Haz la A: elegir C.4, implementar, documentar dificultades encontradas, organizar mañana en Engram/todo, luego commit y pushealo."
L2 (asistente, inicio): Creado `odd/tasks/feature-c4-search.md`. Diseño: endpoint `/pacientes/search` con filtros por query params (`nombre`, `edad_min`, `edad_max`, `peso_min`, `peso_max`, `limit`, `offset`). Filtro en memoria sobre `repo.GetAll`, igual que `GetPacientes`. No se modifica interfaz `PacienteRepository`.

L3 (asistente, cierre): Implementado y commiteado `a4675bb feat: endpoint GET /pacientes/search con filtros y paginación (C.4)`.
- `SearchPacientes` en `handlers/paciente.go`: filtros por nombre (parcial, case-insensitive), edad_min/edad_max, peso_min/peso_max, paginación limit (default 10) / offset (default 0).
- 7 tests nuevos en `handlers/paciente_test.go`: sin filtros, nombre, edad min/max, peso min/max, paginación, valores por defecto, resultado vacío.
- `PROJECT_INDEX.md` actualizado: tabla de rutas con `/pacientes/search`, estado post C.4, sección 6.2 dificultades encontradas y resoluciones.
- `go test ./...`, `go vet ./...`, `go build ./...` todos verdes.
- Pusheado a `origin/main`. Rama limpia.

Dificultades encontradas y resoluciones (documentadas en PROJECT_INDEX.md §6.2):
1. Echo v4 elimino `c.Response().Body`; se usa `c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes()` en tests.
2. `c.Param("id")` vacío al llamar handlers directo en tests; helper `setParam` con `reflect` + `unsafe.Pointer`.
3. `MockPacienteRepository.Create` no asignaba ID al crear; contador atómico `nextID` asigna `paciente-<n>`.
4. `go.sum` crece al agregar `cloud.google.com/go/firestore`; `go mod tidy` antes de commitear.
