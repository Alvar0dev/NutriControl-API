# CHANGELOG.md

## Versión 1.0.4 - 2026-11-02

### PATCH `PATCH /pacientes/{id}` (actualización parcial)

- `PatchPaciente` en `handlers/paciente.go`: aplica solo los campos enviados en el body (nombre, edad, peso, porcentaje_grasa, notas). Los campos no enviados se conservan.
- Requiere que el paciente exista (`GetByID`), si no existe devuelve 404.
- Valida `edad > 0` y `peso > 0` solo cuando esos campos se envían; 400 si inválidos.
- Ruta `PATCH /pacientes/{id}` registrada en `SetupRoutes`.
- Tests agregados: `TestPatchPacienteOK`, `TestPatchPacienteSoloPeso`, `TestPatchPacienteNoEncontrado`, `TestPatchPacienteEdadInvalida`.
- `go test ./...`, `go vet ./...`, `go build ./...` todos verdes.
- Commit: `9c6f605 feat: agregar PATCH /pacientes/{id} actualización parcial y archivar tracking C.4`.
- Tracking ODD archivado: `odd/tasks/feature-c4-search.md` movido a `odd/tasks/archive/`.
- Estado sincronizado en `odd/task/estado-actual-proyecto.md` (T11 + L10) y `PROJECT_INDEX.md`.

## Versión 1.0.3 - 2026-11-02

### [C.6] Completar `PUT /pacientes/{id}` con tests de éxito y no encontrado

**¿Qué se hizo y por qué:**
El handler `UpdatePaciente` y la ruta `PUT /pacientes/{id}` ya existían desde la versión 1.0.0, con validación de ID (`c.Param("id")`) y delegación al repositorio `Update`. Lo que faltaba para cerrar C.6 eran los **tests de éxito (200) y no encontrado (404)**, porque sin ellos el endpoint no tenía cobertura de su caso principal: actualizar un paciente existente y devolverlo con los datos nuevos, ni el caso de error cuando el ID no existe.

**Tests agregados en `handlers/paciente_test.go`:**
- `TestUpdatePacienteOK`: crea un paciente con `repo.Create` (el mock le asigna `paciente-<n>`), luego llama `UpdatePaciente` con payload nuevo y verifica `200 OK`, que `ID` se conserva y que `nombre`, `edad`, `peso`, `porcentaje_grasa` y `notas` se actualizan.
- `TestUpdatePacienteNoEncontrado`: llama `UpdatePaciente` con ID `inexistente` y verifica `404 Not Found` con mensaje `"No se encontró paciente"`.

**Tests preexistentes que cubren la validación de negocio:**
- `TestUpdatePacienteEdadInvalida`: `edad <= 0` → `400 Bad Request` con mensaje sobre `edad`.
- `TestUpdatePacientePesoInvalido`: `peso <= 0` → `400 Bad Request` con mensaje sobre `peso`.

**Diseño:**
- El mock `Update` modifica el paciente en `m.store` por índice y lo devuelve, así que el handler solo pasa el ID de `c.Param("id")` y el cuerpo parseado. Valida, delega al repo, devuelve el resultado.
- El helper `setParam` (reflect + unsafe.Pointer) inyecta `pnames`/`pvalues` en el contexto de Echo v4 para que `c.Param("id")` funcione cuando el test invoca el handler directo, sin pasar por el router.
- Commit: `85ed67f feat: completar C.6 PUT /pacientes/{id} con tests de éxito y no encontrado`.
- `go test ./...`, `go vet ./...`, `go build ./...` todos verdes.

## Versión 1.0.2 - 2026-11-02

### [C.5] Endpoint `GET /pacientes/stats` (resumen de solo lectura)
- `StatsPacientes` devuelve `{total, edad_promedio, peso_promedio, porcentaje_grasa_promedio, edades[], pesos[]}` agregando en memoria sobre `repo.GetAll`. Solo lectura, no modifica datos.
- Ruta `GET /pacientes/stats` registrada en `SetupRoutes`.
- Tests agregados: `TestStatsPacientesOK` y `TestStatsPacientesListaVacia`.
- `go test ./...`, `go vet ./...`, `go build ./...` todos verdes.

## Versión 1.0.1 - 2026-10-26

### [C.2] Mock asigne ID autoincremental al crear + GetByID/Delete
- `MockPacienteRepository.Create` ahora asigna `paciente-<n>` automáticamente cuando `p.ID` viene vacío, usando `atomic.AddInt64(&m.nextID, 1)`. Respeta el ID si ya viene seteado.
- Se implementaron `GetByID`, `Update` y `Delete` en el mock para permitir operaciones sobre datos creados en memoria.
- Handler `GetPacienteByID` (GET `/pacientes/{id}`): devuelve 200 con el paciente o 404 si no existe.
- Handler `DeletePaciente` (DELETE `/pacientes/{id}`): devuelve 204 o 404 si no existe.
- Tests agregados: `TestMockCreateAsignaIDCuandoEstaVacio`, `TestMockCreateRespetaIDExistente`, `TestGetPacienteByIDOK`, `TestGetPacienteByIDNoEncontrado`.
- Commit: `65b514d feat: tests de GetPacienteByID y mock asigna ID al crear`.

### [C.3] Endpoint `GET /pacientes?nombre=` (filtro por nombre)
- `GetPacientes` filtra por `c.QueryParam("nombre")` con coincidencia parcial y sin distinción de mayúsculas/minúsculas.
- Test agregado: `TestGetPacientesFiltrarPorNombre`.
- Commit: `9102ca9 feat: filtrar pacientes por nombre en GET /pacientes`.

## Versión 1.0.0 - 2026-10-20

### [A] Añadido endpoint `PUT /pacientes/{id}`
- Implementación del handler `UpdatePaciente` que permite actualizar un paciente por ID.
- Ruta `PUT /pacientes/{id}` ahora funcional con validación del ID.
- Se añadió el campo `ID string` al modelo `Paciente` para mantener la referencia única.
- Se implementó el método `Update` en `MockPacienteRepository` (mock) y se preparó la implementación real en `PacienteFirestoreRepository`.

### [B] Mejoras en el modelo `Paciente`
- Se agregó el campo `ID string` con tag JSON `"id,omitempty"` para asegurar que cada paciente tenga un identificador único.

### [C] Corrección de firmas de repositorio en handlers
- Los handlers y `SetupRoutes` ahora reciben `firestore.PacienteRepository` (interfaz por valor) en lugar de `*firestore.PacienteRepository` (puntero a interfaz).
- Motivo: un puntero a interfaz no expone los métodos de la interfaz, lo que rompía la compilación de `handlers` y `main`.
- `main.go` sigue igual: ya declaraba `var repo firestore.PacienteRepository`.

## Pendiente
- Implementar `Update` real en `pkg/firestore/repo.go` (integración con Firestore).
- Resolver dependencias faltantes (`cloud.google.com/go/firestore`) en `go.mod` (preexistente, no relacionado con Update).
- Agregar pruebas unitarias para `PUT /pacientes/{id}` y verificar con `curl` (servidor requiere módulo Firestore resuelto para arrancar).
- Verificar persistencia tras reinicio del servidor.