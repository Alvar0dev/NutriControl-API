# AGENTS.md - NutriControl API Operational Context (ODD / Gentle Shell)

> Flujo ODD (Organic Driven Development): cambios emergen del código, no de specs rígidas previas. La especificación SDD antigua (`specs/001-persistence-layer/`) está archivada en `specs/archive/`.

## Project Summary
NutriControl API proporciona endpoints para gestión de pacientes (Go + Echo v4). CRUD completo con persistencia Firestore, repositorio abstracto (`PacienteRepository`) y mock para desarrollo/test.

## Current Architecture
- **Language**: Go 1.26.1
- **Web Framework**: Echo v4.16.0
- **Structure**: Domain (`models/`) / HTTP (`handlers/`) / Adaptador (`pkg/firestore/`).
- **Persistence**: Firestore (`pkg/firestore/repo.go`, `paciente_repo.go`) con mock (`mock.go`).
- **State**: Persistencia operativa, CRUD completo (`GET/POST/PUT/DELETE` por ID). Tests presentes (`handlers/`, `models/`).

## Key Commands
```bash
# Run the API server
go run ./...

# Project structure
./main.go           # Entry point and server setup
./handlers/         # HTTP handlers and routes
./models/           # Domain models
./go.mod           # Dependencies
```

## Established Conventions
- **Code Structure**: Domain (`models/`) / Presentation (`handlers/`) / Adapter (`pkg/firestore/`).
- **JSON Tags**: Struct fields include `json:"field_name"` annotations (`models/paciente.go`).
- **Error Handling**: Basic HTTP error responses with JSON bodies (`handlers/paciente.go`).
- **Mocking**: Mock repository (`pkg/firestore/mock.go`) + Firestore real (`pkg/firestore/repo.go`).
- **Inyección**: Repositorio por argumento a handlers (`SetupRoutes`).
- **Commit Style**: Conventional commits.
- **Documentación de arranque**: `PROJECT_INDEX.md`, `DEV_RUNBOOK.md`. Especificación SDD archivada (`specs/archive/001-persistence-layer/`).

## Boundaries (No-touch zones without explicit permission)
- **Data Persistence**: `pkg/firestore/` es la capa de persistencia activa; no crear `data/`, `db/`, `storage/` sin acuerdo.
- **Testing Strategy**: Tests existentes (`handlers/paciente_test.go`, `models/paciente_test.go`). Nuevos tests deben seguir Go (`testing` + `testify`).
- **API Documentation**: No OpenAPI/Swagger aún. Contactos por archivo (`handlers/paciente.go`).
- **Configuration**: No `.env` estructurado aún (`main.go` lee env vars directa).