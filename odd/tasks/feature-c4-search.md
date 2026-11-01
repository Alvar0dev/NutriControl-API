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
| T1 | S1 | inline | in_progress | Diseño del endpoint y filtros (en curso) |
| T2 | S2 | inline | pending | Implementación del handler `SearchPacientes` + registro en `SetupRoutes` |
| T3 | S3 | inline | pending | Tests RED → GREEN para search (nombre, edad, peso, paginación, vacío, por defecto) |
| T4 | S3 | inline | pending | Documentar dificultades encontradas (ésta es la parte clave que falta) |
| T5 | S1 | inline | pending | Commit + push (autorizado por usuario) |

## Log

L1 (usuario, hoy): "Haz la A: elegir C.4, implementar, documentar dificultades encontradas, organizar mañana en Engram/todo, luego commit y pushealo."
L2 (asistente, inicio): Creado `odd/tasks/feature-c4-search.md`. Diseño: endpoint `/pacientes/search` con filtros por query params (`nombre`, `edad_min`, `edad_max`, `peso_min`, `peso_max`, `limit`, `offset`). Filtro en memoria sobre `repo.GetAll`, igual que `GetPacientes`. No se modifica interfaz `PacienteRepository`.
