# Paginación nativa en Firestore real (GetAll con limit/offset)

**Feature**: B — paginación nativa en Firestore real  
**Rama**: `feat/paginacion-firestore-getall` (desde `main`, HEAD `2be01b6`)  
**Estado**: en planificación

## Contexto

El CRUD de NutriControl-API está completo y verde (GET/POST/PUT/PATCH/DELETE por ID, filtro nombre, search con paginación en memoria, stats). El endpoint `GET /pacientes/search` ya paginaría en memoria, pero el repositorio real de Firestore (`PacienteFirestoreRepo.GetAll`) trae **todos** los documentos sin aplicar `limit`/`offset` en la consulta de Firestore. La meta de esta feature es que `GetAll` acepte `limit` y `offset` y los aplique nativamente tanto en Firestore como en el mock, y que los handlers los pasen desde los query params.

## Specs

- **S1**: `PacienteRepository.GetAll(ctx, limit, offset)` firma incluye `limit int` y `offset int` (quote del usuario: "paginación nativa en Firestore real: limit/offset en GetAll del repo real").
- **S2**: `PacienteFirestoreRepo.GetAll` aplica `.Limit(limit)` y `.Offset(offset)` en la consulta de Firestore; cuando `limit <= 0` no aplica límite y cuando `offset <= 0` no aplica offset.
- **S3**: `MockPacienteRepository.GetAll(ctx, limit, offset)` aplica `limit`/`offset` en memoria sobre `m.store` (equivalente al comportamiento actual de `SearchPacientes`).
- **S4**: Handlers que llaman a `GetAll` (`GetPacientes`, `SearchPacientes`, `StatsPacientes`) leen `limit` (default 10) y `offset` (default 0) de los query params y se los pasan al repo; `SearchPacientes` deja de paginar en memoria (lo hace el repo).
- **S5**: Tests del mock y de handlers cubren paginación: `limit`, `offset`, valores por defecto, y resultado vacío cuando `offset` supera la cantidad de elementos.

## Tasks

| ID | Specs | Route | Commit |
|----|-------|-------|--------|
| T1 | S1, S2, S3 | RED: escribir tests que fallen (mock + repo real + handlers) | — |
| T2 | S1 | Cambiar interfaz `PacienteRepository.GetAll` para recibir `limit, offset` | `refactor: firma GetAll con limit/offset` |
| T3 | S2 | Implementar paginación nativa en `PacienteFirestoreRepo.GetAll` (Firestore `.Limit`/`.Offset`) | `feat: paginación nativa GetAll Firestore` |
| T4 | S3 | Implementar paginación en `MockPacienteRepository.GetAll` | `feat: paginación GetAll mock` |
| T5 | S4 | Actualizar handlers `GetPacientes`, `SearchPacientes`, `StatsPacientes` para pasar limit/offset al repo y quitar paginación en memoria de `SearchPacientes` | `feat: handlers pasan limit/offset a GetAll` |
| T6 | S5 | GREEN: correr tests, refactor, verificar diff | — |
| T7 | S1-S5 | `go mod tidy`, `go build ./...`, `go vet ./...`, `go test ./...` | — |
| T8 | S1-S5 | Commit convencional en feature branch, cierra la feature | `feat: paginación nativa GetAll en Firestore y mock` |

## Log

- **L1**: `haz el 2` (usuario eligió feature B: paginación nativa en Firestore real; ver TODO consolidado en Engmar `nutricontrol-api/todo-manana`)
- L2: rama `feat/paginacion-firestore-getall` creada desde `main` (HEAD `2be01b6`, working tree limpio)
- L3: todos los tests escritos y verdes (T1-T6), `go mod tidy`/`go build`/`go vet`/`go test` todos verdes (T7); commit pendiente (T8)
