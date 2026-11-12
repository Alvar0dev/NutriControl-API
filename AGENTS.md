# AGENTS.md - NutriControl API Operational Context (ODD / Gentle Shell)

> Flujo ODD (Organic Driven Development): cambios emergen del código, no de specs rígidas previas. La especificación SDD antigua está archivada.

## Project Summary
NutriControl API proporciona endpoints para gestión de pacientes (Go + Echo v4). CRUD completo con persistencia Firestore, repositorio abstracto (`PacienteRepository`) y mock para desarrollo/test.

## Current Architecture
- **Language**: Go 1.26.1
- **Web Framework**: Echo v4.16.0
- **Structure**: Domain (`models/`) / HTTP (`handlers/`) / Adaptador (`pkg/firestore/`).
- **Persistence**: Firestore (`pkg/firestore/repo.go`, `paciente_repo.go`) con mock (`mock.go`).
- **State**: Persistencia operativa, CRUD completo (`GET/POST/PUT/PATCH/DELETE` por ID). Tests presentes (`handlers/`, `models/`).
- **Mock**: `Create` asigna `paciente-<n>` automáticamente cuando `p.ID` está vacío (contador atómico `nextID`). Respeta el ID si ya viene. `GetByID`, `Update` y `Delete` implementados sobre el store en memoria.
- **Filtro nombre y ordenamiento**: `GET /pacientes?nombre=&sort=&order=` filtra por coincidencia parcial y ordena por `nombre`, `edad`, `peso`, `porcentaje_grasa` en `asc`/`desc`.
- **Search**: `GET /pacientes/search` con filtros `nombre` (parcial, case-insensitive), `edad_min`/`edad_max`, `peso_min`/`peso_max`, `porcentaje_grasa_min`/`porcentaje_grasa_max`, ordenamiento `sort`/`order` y paginación `limit` (default 10) / `offset` (default 0). Filtro y orden en memoria sobre `repo.GetAll`.
- **Export CSV**: `GET /pacientes/export` descarga el listado completo en formato CSV compatible con RFC 4180 (`attachment; filename="pacientes.csv"`). Tests: `TestExportPacientesCSVOK`, `TestExportPacientesCSVVacio`, `TestExportPacientesCSVRutaRegistrada`.
- **Stats**: `GET /pacientes/stats` devuelve resumen de solo lectura: `total`, `edad_promedio`, `peso_promedio`, `porcentaje_grasa_promedio`, `edades[]`, `pesos[]`.
- **Count**: `GET /pacientes/count` devuelve `{"total": <n>}` consultando `repo.GetAll(ctx, 0, 0)`. Tests: `TestCountPacientesOK`, `TestCountPacientesVacio`.
- **Health**: `GET /health` devuelve `{"status": "ok", "storage": "mock"|"firestore"}` según el adaptador inyectado. Test: `TestHealthCheckOK`.
- **Filtro por edad exacta**: `GET /pacientes/edad/{edad}` devuelve lista de pacientes cuya edad coincida exactamente (200 con lista vacía si no hay coincidencia; 400 si `edad` no es numérico). Tests: `TestGetPacientesPorEdadOK`, `TestGetPacientesPorEdadVacio`, `TestGetPacientesPorEdadInvalida`.
- **PUT validado**: `PUT /pacientes/{id}` actualiza paciente existente con validación (`nombre` requerido, `edad` entre 1 y 120, `peso` > 0 y <= 500, `porcentaje_grasa` entre 0 y 100 → 400; ID inexistente → 404). El nombre se normaliza (trim + capitalización) antes de guardar. Tests: `TestUpdatePacienteOK`, `TestUpdatePacienteNoEncontrado`, `TestUpdatePacienteEdadInvalida`, `TestUpdatePacientePesoInvalido`, `TestUpdatePacientePorcentajeGrasaFueraDeRango`, `TestCreatePacienteEdadFueraDeRango`, `TestCreatePacientePesoFueraDeRango`, `TestCreatePacientePorcentajeGrasaFueraDeRango`, `TestCreatePacienteNombreConEspaciosSeNormaliza`, `TestCreatePacienteNombreSoloEspacios`.
- **PATCH parcial**: `PATCH /pacientes/{id}` actualiza solo los campos enviados (nombre, edad, peso, porcentaje_grasa, notas). Valida `nombre` no vacío, `edad` 1–120, `peso` > 0 y <= 500, y `porcentaje_grasa` 0–100 cuando se envían (400); ID inexistente → 404. El nombre se normaliza cuando se envía. Tests: `TestPatchPacienteOK`, `TestPatchPacienteSoloPeso`, `TestPatchPacienteNoEncontrado`, `TestPatchPacienteEdadInvalida`, `TestPatchPacienteValidacionesExtra`.
- **DELETE por ID**: `DELETE /pacientes/{id}` elimina el paciente por ID (204 No Content si existe; 404 si no existe). Tests: `TestDeletePacienteOK`, `TestDeletePacienteNoEncontrado`.

## Key Commands
```bash
# Run the API server
go run ./...

# Test
go test ./...

# Build
go build -o nutricontrol.exe ./main.go
```

## Project structure
```
./main.go           # Entry point and server setup
./handlers/         # HTTP handlers and routes
./models/           # Domain models
./pkg/firestore/    # Firestore adapter (repo + mock)
./go.mod           # Dependencies
./PROJECT_INDEX.md  # Navigation, commands, and current state
./AGENTS.md         # THIS file
./docs/constitution.md   # Non-negotiable principles
./odd/task/estado-actual-proyecto.md  # Single ODD state definition
```

## Established Conventions
- **Code Structure**: Domain (`models/`) / Presentation (`handlers/`) / Adapter (`pkg/firestore/`).
- **JSON Tags**: Struct fields include `json:"field_name"` annotations (`models/paciente.go`).
- **Error Handling**: Basic HTTP error responses with JSON bodies (`handlers/paciente.go`).
- **Mocking**: Mock repository (`pkg/firestore/mock.go`) + Firestore real (`pkg/firestore/repo.go`).
- **Inyección**: Repositorio por argumento a handlers (`SetupRoutes`).
- **Commit Style**: Conventional commits.
- **ODD**: Trabajo small inline sin documentos de feature; trabajo large con `odd/tasks/<feature>.md` + espejo en Engram.

## Boundaries (No-touch zones without explicit permission)
- **Data Persistence**: `pkg/firestore/` es la capa de persistencia activa; no crear `data/`, `db/`, `storage/` sin acuerdo.
- **Testing Strategy**: Tests existentes (`handlers/paciente_test.go`, `models/paciente_test.go`). Nuevos tests deben seguir Go (`testing` + `testify`).
- **API Documentation**: No OpenAPI/Swagger aún. Contactos por archivo (`handlers/paciente.go`).
- **Configuration**: No `.env` estructurado aún (`main.go` lee env vars directa).