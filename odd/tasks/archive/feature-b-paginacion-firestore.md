# Paginación nativa en Firestore real (GetAll con limit/offset)

**Feature**: B — paginación nativa en Firestore real  
**Rama**: `feat/paginacion-firestore-getall` (mergeada en `main`, commit `d4f5c15`)  
**Estado**: completada y archivada

## Contexto

El CRUD de NutriControl-API está completo y verde (GET/POST/PUT/PATCH/DELETE por ID, filtro nombre, search con paginación, stats). El repositorio real de Firestore (`PacienteFirestoreRepo.GetAll`) traía todos los documentos sin aplicar `limit`/`offset` en la consulta de Firestore. Esta feature actualizó `GetAll` para aceptar `limit` y `offset` y aplicarlos nativamente tanto en Firestore como en el mock.

## Specs

- **S1**: `PacienteRepository.GetAll(ctx, limit, offset)` firma incluye `limit int` y `offset int`.
- **S2**: `PacienteFirestoreRepo.GetAll` aplica `.Limit(limit)` y `.Offset(offset)` en la consulta de Firestore; cuando `limit <= 0` no aplica límite y cuando `offset <= 0` no aplica offset.
- **S3**: `MockPacienteRepository.GetAll(ctx, limit, offset)` aplica `limit`/`offset` en memoria sobre `m.store`.
- **S4**: Handlers que llaman a `GetAll` (`GetPacientes`, `SearchPacientes`, `StatsPacientes`) leen `limit` (default 10) y `offset` (default 0) de los query params y se los pasan al repo.
- **S5**: Tests del mock y de handlers cubren paginación: `limit`, `offset`, valores por defecto, y resultado vacío cuando `offset` supera la cantidad de elementos.

## Tasks

| ID | Specs | Route | Status | Commit |
|----|-------|-------|--------|--------|
| T1 | S1, S2, S3 | RED: escribir tests que fallen (mock + repo real + handlers) | done | `d4f5c15` |
| T2 | S1 | Cambiar interfaz `PacienteRepository.GetAll` para recibir `limit, offset` | done | `d4f5c15` |
| T3 | S2 | Implementar paginación nativa en `PacienteFirestoreRepo.GetAll` (Firestore `.Limit`/`.Offset`) | done | `d4f5c15` |
| T4 | S3 | Implementar paginación en `MockPacienteRepository.GetAll` | done | `d4f5c15` |
| T5 | S4 | Actualizar handlers `GetPacientes`, `SearchPacientes`, `StatsPacientes` para pasar limit/offset al repo | done | `d4f5c15` |
| T6 | S5 | GREEN: correr tests, refactor, verificar diff | done | `d4f5c15` |
| T7 | S1-S5 | `go mod tidy`, `go build ./...`, `go vet ./...`, `go test ./...` | done | `d4f5c15` |
| T8 | S1-S5 | Commit convencional y merge en `main` | done | `d4f5c15` |

## Log

- **L1**: `haz el 2` (usuario eligió feature B: paginación nativa en Firestore real)
- **L2**: rama `feat/paginacion-firestore-getall` creada desde `main` (HEAD `2be01b6`, working tree limpio)
- **L3**: todos los tests escritos y verdes (T1-T6), `go mod tidy`/`go build`/`go vet`/`go test` todos verdes (T7)
- **L4**: mergeado a `main` en commit `d4f5c15` (`feat: paginación nativa GetAll con limit/offset en Firestore y mock`) y archivado en `odd/tasks/archive/`
