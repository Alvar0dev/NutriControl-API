# Estado actual del proyecto — NutriControl-API (ODD)

## Objetivo
Documentar el estado actual del proyecto tras la migración a Organic Driven Development (ODD), sin artefactos SDD remanentes, y establecer el punto de partida para la siguiente feature.

Branch: `main` (origin: `https://github.com/Alvar0dev/NutriControl-API.git`)
Delivery strategy: `ask-on-risk` (default)
Test runner: `go test ./...`

## Specs

S1: "abandonamos la metodología Spec Driven Development (SDD) y operaremos estrictamente bajo el contrato de comportamiento Organic Driven Development (ODD)" — el proyecto ya no usa SDD; los artefactos SDD (`specs/archive/001-persistence-layer/`, 156 líneas) fueron eliminados en commit `0fb5efa`; el directorio `openspec/` nunca existió en este repositorio.

S2: "Centraliza el contexto utilizando exclusivamente Engram y un único archivo de definición" — el contexto vivo se mantiene en Engram (proyecto `nutricontrol-api`) y en este archivo `odd/task/estado-actual-proyecto.md` como único archivo de definición del estado actual.

S3: "Crea el archivo `odd/task/<feature>.md` para documentar la funcionalidad o estado actual del proyecto" — satisfecho por este mismo archivo.

## Tasks

| ID | Linked S# | Route | Status | Commit |
|----|-----------|-------|--------|--------|
| T1 | S1 | inline | done | `8a1a47f` feat: validar edad y peso > 0 en crear y actualizar paciente |
| T2 | S1 | inline | done | `95f019e` docs: completar PROJECT_INDEX.md y DEV_RUNBOOK.md (opción B) |
| T3 | S2 | inline | done | Engram registro de trabajo del día (mem_save id 11 y 12) |
| T4 | S3 | inline | done | `odd/task/estado-actual-proyecto.md` (este documento) |
| T5 | S1 | inline | done | `65b514d` feat: tests de GetPacienteByID y mock asigna ID al crear |
| T6 | S1 | inline | done | `9102ca9` feat: filtrar pacientes por nombre en GET /pacientes |
| T7 | S1 | inline | done | `b98f2a5` docs: simplificar organización de documentación según ODD |
| T8 | S1 | inline | done | `a4675bb` feat: endpoint GET /pacientes/search con filtros y paginación (C.4) |
| T9 | S1 | inline | done | `f06b5aa` feat: endpoint GET /pacientes/stats con resumen de solo lectura (C.5) |
| T10 | S1 | inline | done | `85ed67f` feat: completar C.6 PUT /pacientes/{id} con tests de éxito y no encontrado |
| T11 | S1 | inline | done | `9c6f605` feat: agregar PATCH /pacientes/{id} actualización parcial + archivar tracking C.4 |
| T12 | S1 | inline | done | `d4f5c15` feat: paginación nativa GetAll con limit/offset en Firestore y mock |
| T13 | S1 | inline | done | `f21a6a8` feat: validaciones de negocio extra (edad 1-120, peso 0-500) y normalizar nombres en crear/actualizar |
| T14 | S1 | inline | done | `ae00eea` feat: endpoint GET /pacientes/edad/{edad} filtro por edad exacta (handler + ruta + 2 tests) |
| T15 | S1 | inline | done | `a0ad130` docs: restaurar PROJECT_INDEX, archivar feature B y sincronizar estado ODD |
| T16 | S1 | inline | done | `0fd27a6` feat: validar porcentaje_grasa y nombre en PATCH y cubrir Delete y Edad en tests |
| T17 | S1 | inline | done | `8d32dd6` feat: endpoint `GET /health` (`HealthCheck` + `TestHealthCheckOK`) |
| T18 | S1 | inline | done | `8d32dd6` feat: endpoint `GET /pacientes/count` (`CountPacientes` + `TestCountPacientesOK`, `TestCountPacientesVacio`) |
| T19 | S1 | inline | done | `8d32dd6` feat: filtros `porcentaje_grasa_min`/`max` en `GET /pacientes/search` (`TestSearchPacientesFiltrarPorPorcentajeGrasaMinMax`) |
| T20 | S1 | inline | done | `05a97f8` feat: ordenamiento dinámico en pacientes y endpoint GET /pacientes/export (CSV) |
| T21 | S1 | inline | done | `05a97f8` feat: ordenamiento dinámico en pacientes y endpoint GET /pacientes/export (CSV) |
| T22 | S1 | inline | done | `00e49ce` feat: endpoint POST /pacientes/batch para alta masiva validada |
| T23 | S1 | inline | done | `d413d02` feat: extensión de dominio con Altura y endpoints de cálculo de IMC |
| T24 | S1 | inline | done | `79d9c53` feat: endpoint GET /pacientes/reporte para analítica poblacional |
| T25 | S1 | inline | done | `be4cc9a` feat: soft delete en pacientes e historial de controles clínicos |
| T26 | S1 | inline | done | `0660bf5` test: cubrir PacienteFirestoreRepo sobre gRPC in-process y Emulator |
| T27 | S1 | inline | done | `e220d1c` feat: endpoint GET /pacientes/:id/controles/resumen evolutivo |
| T28 | S1 | inline | done | `8f4cad4` feat: filtrado por fechas y paginación en GET /pacientes/:id/controles |
| T29 | S1 | inline | done | `986116a` feat: exportación CSV de controles clínicos y altura en ExportPacientesCSV |
| T30 | S1 | inline | done | `027f1af` feat: edición y eliminación de controles clínicos puntuales |
| T31 | S1 | inline | done | `b3df81f` ci: agregar workflow de GitHub Actions con vet, test -race y build |
| T32 | S1 | inline | done | `161afbf` feat: hardening HTTP, PORT por env var, middlewares y graceful shutdown en main.go |
| T33 | S1 | inline | done | `db52369` feat: consulta puntual y actualización parcial de controles clínicos (GET/PATCH) |

## Log

L1: [usuario] "Actúa como Agente Orquestador. A partir de este momento, abandonamos la metodología Spec Driven Development (SDD) y operaremos estrictamente bajo el contrato de comportamiento Organic Driven Development (ODD). ... Fase 1: Limpieza y Deprecación (SDD)... Fase 2: Nueva Estructura Documental (ODD)... Fase 3: Contrato de Ejecución ODD..."

L2: [asistente] Verificado: `openspec/` no existe; `specs/archive/` está vacío (contenido eliminado en `0fb5efa`); no hay plantillas de especificación sueltas. Creado `odd/task/estado-actual-proyecto.md` como único archivo de definición del estado actual.

L3: [usuario] Hoy: hacer A (GetByID con tests) y B (mock asigne ID al crear), actualizar documentación con lo hecho y dificultades encontradas, simplificar docs y que no parezcan de IA.

L4: [asistente] Implementado en branch `feat/getbyid-tests-y-mock-id`, commit `65b514d`.
- Tests de handler `GetPacienteByID`: 200 OK con paciente existente y 404 cuando no existe.
- `MockPacienteRepository.Create` asigna `paciente-<n>` automáticamente cuando `ID` está vacío; respeta el ID si ya viene.
- Dificultad: ninguna relevante. Los tests de handler necesitan `context.Background()` y `setParam` para inyectar el path param `{id}` directamente en el contexto de Echo v4.
- `go test ./...`, `go vet ./...` y `go build ./...` verdes.

L5: [asistente] Implementado C.3: `GET /pacientes?nombre=` con filtro por nombre (coincidencia parcial, sin distinción de mayúsculas/minúsculas). Commit `9102ca9 feat: filtrar pacientes por nombre en GET /pacientes`. Test: `TestGetPacientesFiltrarPorNombre`.

L6: [usuario] Hoy: documentar todo el proceso (C.2 y C.3) en la documentación del proceso con los errores encontrados como si fuera yo quien lo hubiera hecho, commitear y pushear.
L7: [asistente] Actualizado `PROJECT_INDEX.md` (sección 6 con estado C.2/C.3 + nueva sección 6.2 dificultades encontradas y resoluciones), `docs/CHANGELOG.md` (versión 1.0.1 con entradas C.2 y C.3), `docs/archive/DEV_RUNBOOK.md` (gotcha 4.2 marcado como resuelto), `odd/task/estado-actual-proyecto.md` (tasks T6/T7 y log L5-L7), `AGENTS.md` (Current Architecture). Todos los cambios commiteados y pusheados a `main`.

L8: [asistente] Implementado C.5: `GET /pacientes/stats` resumen de solo lectura (total, edad_promedio, peso_promedio, porcentaje_grasa_promedio, edades[], pesos[]). Handler en `handlers/paciente.go`, ruta registrada en `SetupRoutes`. Tests `TestStatsPacientesOK` y `TestStatsPacientesListaVacia` en `handlers/paciente_test.go`. Documentación: `PROJECT_INDEX.md` (tabla de rutas + sección 6.3), `docs/CHANGELOG.md` (v1.0.2), `AGENTS.md` (Current Architecture), `odd/task/estado-actual-proyecto.md` (task T9). `go test ./...`, `go vet ./...`, `go build ./...` verdes. Commit `f06b5aa` + pusheado a `origin/main`.

L9: [asistente] Completado C.6: `PUT /pacientes/{id}` (actualizar paciente con validación). Handler `UpdatePaciente` y ruta ya existían (v1.0.0). Se agregaron tests faltantes `TestUpdatePacienteOK` (200 con datos actualizados) y `TestUpdatePacienteNoEncontrado` (404 ID inexistente). Tests de validación preexistentes `TestUpdatePacienteEdadInvalida` y `TestUpdatePacientePesoInvalido` (400). `go test ./...`, `go vet ./...`, `go build ./...` verdes. Commit `85ed67f` en local (pendiente de push). Documentación pendiente de commitear (PROJECT_INDEX.md, CHANGELOG.md, AGENTS.md, odd/task/estado-actual-proyecto.md, odd/tasks/feature-c4-search.md).

L10: [usuario] "haz A y B Luego actualiza la documentacion, comitéalo y pusalo."
L10: [asistente] A: movido `odd/tasks/feature-c4-search.md` a `odd/tasks/archive/`. B: implementado `PATCH /pacientes/{id}` (actualización parcial, solo campos enviados) test-first con 4 tests (`TestPatchPacienteOK`, `TestPatchPacienteSoloPeso`, `TestPatchPacienteNoEncontrado`, `TestPatchPacienteEdadInvalida`). `go test ./...`, `go build ./...`, `go vet ./...` verdes. Documentación actualizada: `PROJECT_INDEX.md` (tabla de rutas + sección 6.7), `AGENTS.md` (Current Architecture), `odd/task/estado-actual-proyecto.md` (T11 + L10). Commit `9c6f605 feat: agregar PATCH /pacientes/{id} actualización parcial y archivar tracking C.4` + pusheado a `origin/main`. Rama limpia, working tree limpio.

L11: [usuario] "haz 1" (opción C: validaciones de negocio extra: rangos edad/peso, normalizar nombres)
L11: [asistente] Implementado en `handlers/paciente.go` + tests en `handlers/paciente_test.go`. Commit `f21a6a8 feat: validaciones de negocio extra y normalizar nombres en pacientes`.

L12: [usuario] Implementar Opción C: nuevo endpoint `GET /pacientes/edad/{edad}` (filtro por edad exacta), documentar proceso, commitear y pushear.
L12: [asistente] Implementado `GET /pacientes/edad/{edad}` en `handlers/paciente.go` + 2 tests en `handlers/paciente_test.go` (OK y vacío). Documentación actualizada: `docs/CHANGELOG.md` (v1.0.6), `AGENTS.md` (Current Architecture), `PROJECT_INDEX.md` (tabla de rutas), `odd/task/estado-actual-proyecto.md` (T13 + L12). `go test ./...`, `go vet ./...`, `go build ./...` verdes.
- `validarPaciente`: edad 1-120, peso >0 y <=500; mensajes incluyen el rango.
- `normalizarNombre`: trim + title case, usado en CreatePaciente, UpdatePaciente y PatchPaciente (cuando se envía).
- Nombre solo espacios → 400.
- Tests agregados: `TestCreatePacienteEdadFueraDeRango`, `TestCreatePacientePesoFueraDeRango`, `TestCreatePacienteNombreConEspaciosSeNormaliza`, `TestCreatePacienteNombreSoloEspacios`.
- `go test ./...`, `go vet ./...`, `go build ./...` todos verdes. Commit en `main`, pusheado a `origin/main`.

L13: [usuario] "Perfecto adelante con el 1 y el 2" / "Guarda todo el proceso en la documentacion con todas las dificultades encontradas y el porque se ha construido asi, commitealo y pushealo. y plantea el estado y los todo para mañana continuar con el proyecto."
L13: [asistente] Ejecutados los puntos 1 y 2 en dos unidades de trabajo y documentado el proceso completo:
- Unidad 1 (`a0ad130`): restaurado `PROJECT_INDEX.md` completo desde `HEAD` (se habían perdido las secciones 1–6.7 por un reemplazo accidental previo con texto corrupto), cerrado y archivado `odd/tasks/archive/feature-b-paginacion-firestore.md` (`d4f5c15`) y sincronizados `odd/task/estado-actual-proyecto.md` y `todo`.
- Unidad 2 (`0fd27a6`, v1.0.7): agregada validación de `porcentaje_grasa` en `[0, 100]` (`porcentajeGrasaMaximo = 100`) tanto en `validarPaciente` (`POST`/`PUT`) como en `PatchPaciente`, más rechazo de `nombre` vacío/solo espacios en `PATCH` post-normalización. Agregados tests table-driven (`TestCreatePacientePorcentajeGrasaFueraDeRango`, `TestUpdatePacientePorcentajeGrasaFueraDeRango`, `TestPatchPacienteValidacionesExtra`) y tests faltantes de `DeletePaciente` (`TestDeletePacienteOK`, `TestDeletePacienteNoEncontrado`) y `GetPacientesPorEdad` (`TestGetPacientesPorEdadInvalida`).
- Documentadas todas las dificultades encontradas y decisiones de diseño en `docs/CHANGELOG.md` (v1.0.7) y `PROJECT_INDEX.md` (secciones 6.9 y 6.10). `go test ./...`, `go vet ./...`, `go build ./...` todos verdes y pusheados a `origin/main`.

L14: [usuario] "continua con los 3" / "Guarda todo el proceso en la documentacion con todas las dificultades encontradas y el porque se ha construido asi, commitealo y pushealo. y plantea el estado y los todo para mañana continuar con el proyecto."
L14: [asistente] Implementadas test-first las 3 opciones planteadas (`v1.0.8`, commit `8d32dd6`):
- `GET /health` (`HealthCheck`): devuelve `200 OK` con `{"status": "ok", "storage": "mock"|"firestore"}` mediante aserción de tipo sobre `repo` sin tocar la interfaz `PacienteRepository`. Test: `TestHealthCheckOK`.
- `GET /pacientes/count` (`CountPacientes`): devuelve `200 OK` con `{"total": <n>}` usando `repo.GetAll(ctx, 0, 0)` (elegido frente a IMC porque `models.Paciente` no posee campo `Altura`). Tests: `TestCountPacientesOK` y `TestCountPacientesVacio`.
- `GET /pacientes/search` (`SearchPacientes`): agregados filtros `porcentaje_grasa_min` y `porcentaje_grasa_max`. Test table-driven: `TestSearchPacientesFiltrarPorPorcentajeGrasaMinMax` (ajustado al seed real de `mock.go`: Juan 18.5%, Maria 22.0%).
- Rutas estáticas agrupadas antes de `/pacientes/{id}` en `SetupRoutes`.
- Documentación completa con dificultades y decisiones de diseño actualizada en `AGENTS.md`, `PROJECT_INDEX.md` (secciones 6.10 y 6.11), `docs/CHANGELOG.md` (v1.0.8), `odd/task/estado-actual-proyecto.md` y `todo`. `go test ./...`, `go vet ./...`, `go build ./...` todos verdes y pusheados a `origin/main`.

L15: [usuario] "haz 1 y 2"
L15: [asistente] Implementadas test-first las opciones 1 y 2 (v1.0.9):
- Opción 1: Ordenamiento dinámico en memoria (`sort` por `nombre`, `edad`, `peso`, `porcentaje_grasa`; `order` por `asc` o `desc`) para `GET /pacientes` y `GET /pacientes/search`. Tests: `TestGetPacientesOrdenamiento` y `TestSearchPacientesOrdenamiento`.
- Opción 2: Endpoint `GET /pacientes/export` para volcado completo en formato CSV compatible con RFC 4180 mediante `encoding/csv` y cabeceras de descarga (`attachment; filename="pacientes.csv"`). Tests: `TestExportPacientesCSVOK`, `TestExportPacientesCSVVacio` y `TestExportPacientesCSVRutaRegistrada`.
- Rutas estáticas protegidas registrando `/pacientes/export` antes de `/pacientes/{id}`. `go test ./...`, `go vet ./...`, `go build ./...` verdes.

L16: [usuario] "Guarda todo el proceso en la documentacion con todas las dificultades encontradas y el porque se ha construido asi, commitealo y pushealo. y plantea el estado y los todo para mañana continuar con el proyecto."
L16: [asistente] Documentación consolidada con todas las dificultades y decisiones de diseño v1.0.9 (precedencia de rutas en router Echo v4, escape RFC 4180 con encoding/csv, formateo numérico con strconv.FormatFloat, ordenamiento estable con sort.SliceStable y diagnóstico de configuración RDD en host) en docs/CHANGELOG.md, PROJECT_INDEX.md, odd/task/estado-actual-proyecto.md y todo. Commit documental y push a origin/main completados.

L17: [usuario] "haz 1" (Endpoint POST /pacientes/batch alta masiva validada)
L17: [asistente] Implementado test-first el endpoint `POST /pacientes/batch` (`CreatePacientesBatch`):
- Validación atómica en 2 fases: rechazo de lote vacío `[]` con `400 Bad Request`, normalización de nombres y validación clínica (`validarPaciente`). Si algún elemento es inválido, rechaza inmediatamente con `400 Bad Request` indicando la posición del paciente y no persiste ningún registro en el repositorio (cero escrituras parciales).
- Persistencia secuencial de registros válidos y respuesta `201 Created` con el slice de pacientes y sus IDs generados.
- Ruta registrada en `SetupRoutes`.
- Tests agregados en `handlers/paciente_test.go`: `TestCreatePacientesBatchOK`, `TestCreatePacientesBatchVacio`, `TestCreatePacientesBatchAtomicidadInvalido`, `TestCreatePacientesBatchJSONInvalido` y `TestCreatePacientesBatchRutaRegistrada`.
- `go test ./...`, `go vet ./...` y `go build ./...` todos verdes (49 tests).

L18: [usuario] "Guarda todo el proceso en la documentacion con todas las dificultades encontradas y el porque se ha construido asi, commitealo y pushealo. y plantea el estado y los todo para mañana continuar con el proyecto."
L18: [asistente] Documentación consolidada con todas las dificultades y decisiones de diseño v1.0.10 (garantía de atomicidad sin transacciones distribuidas en storage, compatibilidad con la firma de repo.Create retornando *models.Paciente, rechazo estricto de array vacío) en docs/CHANGELOG.md, PROJECT_INDEX.md (3.2, 6.12 y dificultad 14), AGENTS.md, odd/task/estado-actual-proyecto.md y todo.
- Commiteado y pusheado a origin/main.
- Planteado el estado actual y los TODOs para la próxima sesión (Opción 1: Extensión de Dominio con Altura e IMC; Opción 2: Agrupación analítica por rangos etarios/nutricionales; Opción 3: Configuración de host RDD).

L19: [usuario] "haz el 1" (Opción 1: Extensión de Dominio con Altura y cálculo de IMC)
L19: [asistente] Implementada test-first la extensión de dominio con `Altura` y cálculo clínico de IMC (v1.0.11):
- `models.Paciente` ampliado con `Altura float64` (`json:"altura,omitempty"`).
- Reglas clínicas en `validarPaciente` y `PatchPaciente`: rango de `alturaMinima = 0.30` a `alturaMaxima = 2.50` metros si se envía altura (`p.Altura != 0`).
- Endpoint dual `GET /pacientes/imc/{id}` y `GET /pacientes/{id}/imc` (`GetPacienteIMC`): calcula $IMC = peso / altura^2$ (redondeado a 2 decimales) y clasifica según OMS (`Bajo peso`, `Normal`, `Sobrepeso`, `Obesidad`). Responde 200 con `RespuestaIMC`, 404 si el ID no existe, o 400 si el paciente no tiene altura registrada (`altura <= 0`) o peso válido (`peso <= 0`).
- Soporte para ordenamiento por altura en `ordenarPacientes` (`?sort=altura&order=asc|desc`) y actualización en `MockPacienteRepository.Update`.
- Normalización en `SetupRoutes` registrando rutas parametrizadas con `:id` y `:edad` para compatibilidad con el router nativo de Echo v4.
- 10 tests nuevos agregados: 59 tests totales en verde (`go test ./...`, `go vet ./...` y `go build ./...`). Commit `d413d02`.

L20: [usuario] "Guarda todo el proceso en la documentacion con todas las dificultades encontradas y el porque se ha construido asi, commitealo y pushealo. y plantea el estado y los todo para mañana continuar con el proyecto."
L20: [asistente] Documentación consolidada con todas las dificultades y decisiones de diseño v1.0.11 (compatibilidad hacia atrás con omitempty en Altura, sintaxis nativa de router Echo v4 :id vs {id}, ruta dual semántica /pacientes/imc/:id y sub-recurso /pacientes/:id/imc, y validación condicional de altura) en docs/CHANGELOG.md, PROJECT_INDEX.md (6.13 y dificultad 15), AGENTS.md, odd/task/estado-actual-proyecto.md y todo.
- Commiteado y pusheado a origin/main.
- Planteado el estado actual y los TODOs para mañana (Opción 1: Endpoint GET /pacientes/reporte; Opción 2: Configuración host RDD).

L21: [usuario] "haz el todo 1" (Endpoint GET /pacientes/reporte analítica poblacional)
L21: [asistente] Implementado test-first el endpoint `GET /pacientes/reporte` (`ReportePacientes`, v1.0.12):
- Estructuras `RespuestaReporte`, `GruposEtarios` (`menores` <18, `adultos` 18–65, `mayores` >65), `ClasificacionIMCPoblacion` (`bajo_peso`, `normal`, `sobrepeso`, `obesidad`, `sin_datos`) y `PromediosReporte` (`edad`, `peso`, `porcentaje_grasa`, `imc` redondeados a 2 decimales).
- Consulta del padrón completo con `repo.GetAll(ctx, 0, 0)`, reutilización de `clasificarIMC` y cálculo de promedio de IMC únicamente sobre pacientes con altura y peso válidos (`conIMC > 0`), contabilizando en `sin_datos` los registros sin altura.
- Ruta estática `GET /pacientes/reporte` registrada en `SetupRoutes` antes de `/pacientes/:id` y `/pacientes/{id}`.
- 3 tests nuevos agregados en `handlers/paciente_test.go`: `TestReportePacientesOK`, `TestReportePacientesVacio` y `TestReportePacientesRutaRegistrada`.
- `go test -count=1 ./...`, `go vet ./...` y `go build ./...` todos verdes. Commit `79d9c53`.

L22: [usuario] "Guarda todo el proceso en la documentacion con todas las dificultades encontradas y el porque se ha construido asi, commitealo y pushealo. y plantea el estado y los todo para mañana continuar con el proyecto."
L22: [asistente] Documentación consolidada con todas las dificultades y decisiones de diseño v1.0.12 (invariante contable con `sin_datos` para pacientes sin altura, promedio condicional de IMC sobre `conIMC > 0`, consulta del padrón completo con `GetAll(ctx, 0, 0)`, reutilización de `clasificarIMC` con redondeo previo a 2 decimales y precedencia de ruta estática `/pacientes/reporte` en Echo v4) en `docs/CHANGELOG.md`, `PROJECT_INDEX.md` (3.2, 6.14 y dificultad 16), `AGENTS.md`, `odd/task/estado-actual-proyecto.md` y `todo`.
- Commiteado y pusheado a `origin/main`.
- Planteado el estado actual y los TODOs para mañana (Opción 1: Soft delete en pacientes; Opción 2: Historial de seguimiento clínico por paciente; Opción 3: Configuración host RDD).

L23: [usuario] "haz 1 y 2" (Soft delete en pacientes + Historial de controles clínicos por paciente)
L23: [asistente] Implementadas test-first las opciones 1 y 2 (v1.0.13):
- Soft delete: agregado `DeletedAt *time.Time` (`json:"deleted_at,omitempty"`) en `models.Paciente`; `Delete` marca `DeletedAt` UTC; `GetAll`, `GetByID` y `Update` excluyen registros con `DeletedAt != nil` (filtrando activos antes de `limit`/`offset`); agregados `Restore` y `GetInactivos` junto con `GET /pacientes/inactivos` y `POST /pacientes/:id/restore`.
- Historial de controles clínicos: nuevo struct `models.ControlClinico`, métodos `AddControl` y `GetControles` (subcolección `pacientes/{id}/controles` en Firestore, mapa `nil`-safe en mock) y endpoints `POST /pacientes/:id/controles` (valida peso/grasa, fecha UTC por defecto, cálculo automático de IMC si `Altura > 0`) y `GET /pacientes/:id/controles` (orden cronológico ascendente).
- `go test -count=1 ./...` (85 ejecuciones `--- PASS:` incluyendo subtests), `go vet ./...` y `go build ./...` verdes; verificación independiente read-only `PASS` sobre paridad `repo.go`/`mock.go`. Commit `be4cc9a`.

L24: [usuario] "Guarda todo el proceso en la documentacion con todas las dificultades encontradas y el porque se ha construido asi, commitealo y pushealo. y plantea el estado y los todo para mañana continuar con el proyecto."
L24: [asistente] Documentación consolidada con todas las dificultades y decisiones de diseño v1.0.13 (`DeletedAt *time.Time` con `omitempty` frente a `Activo bool`, filtrado de activos previo a `limit`/`offset` en `GetAll`, inicialización perezosa `nil`-safe de `controles` en `MockPacienteRepository`, simetría `Set` vs `Update` sin tags `firestore:` en `PacienteFirestoreRepo`, precedencia de `/pacientes/inactivos`, cálculo condicional de IMC y brecha de tests en vivo para `repo.go`) en `docs/CHANGELOG.md`, `PROJECT_INDEX.md`, `AGENTS.md`, `odd/task/estado-actual-proyecto.md` y `todo`.
- Commiteado y pusheado a `origin/main`.
- Planteado el estado actual y los TODOs para mañana (Opción 1: Tests de integración Firestore Emulator; Opción 2: Resumen evolutivo de controles clínicos; Opción 3: Configuración host RDD).

L25: [usuario] "pues haz los risk" (cierre `Risk: item 6` sobre `pkg/firestore/repo.go`)
L25: [asistente] Agregado `NewPacienteFirestoreRepoWithClient` y soporte condicional para `FIRESTORE_EMULATOR_HOST` en `pkg/firestore/repo.go`, más `pkg/firestore/repo_test.go` con `fakeFirestoreServer` gRPC in-process sobre `firestorepb` (`Commit`, `BatchGetDocuments`, `RunQuery`).
- Cubiertos CRUD, paginación `GetAll`, soft delete + `Restore` + `GetInactivos` y subcolección `controles` sobre `*firestore.Client` real (81.9% cobertura en `pkg/firestore`). Commit `0660bf5`.

L26: [usuario] "haz el 1 y el 2" (`GET /pacientes/:id/controles/resumen` + consolidación documental, commit y push)
L26: [asistente] Implementado test-first `GET /pacientes/:id/controles/resumen` (`GetResumenControlesPaciente`, struct `ResumenControles`, helper `redondear2`) en `handlers/paciente.go` y 3 tests en `handlers/paciente_test.go` (`TestGetResumenControlesPacienteOK`, `TestGetResumenControlesPacienteSinControlesYUnControl`, `TestGetResumenControlesPacienteNoEncontradoYSoftDeleted`).
- `go test -count=1 ./...` (91 ejecuciones `--- PASS:` incluyendo subtests), `go vet ./...` y `go build ./...` verdes. Commit `e220d1c`.
- Documentación consolidada en `docs/CHANGELOG.md` (v1.0.14), `PROJECT_INDEX.md` (3.2, 4.2, 6.16 y dificultades 19-20), `AGENTS.md`, `odd/task/estado-actual-proyecto.md` y `todo`, commiteado y pusheado a `origin/main`.
- Planteado el estado actual y los TODOs para mañana (Opción 1: Exportación CSV de controles clínicos o inclusión de `altura` en `ExportPacientesCSV`; Opción 2: Paginación / filtro por rango de fechas en `GET /pacientes/:id/controles`; Opción 3: Configuración host RDD).

L27: [usuario] "haz todo 1" (Filtrado por rango de fechas `?desde=&hasta=` y paginación en `GET /pacientes/:id/controles`)
L27: [asistente] Implementado test-first en `handlers/paciente.go` (`parseFechaQuery` y actualización de `GetControlesPaciente`) y `handlers/paciente_test.go` (`TestGetControlesPacienteFiltroFechas`, `TestGetControlesPacienteFiltroFechasInvalidas`, `TestGetControlesPacientePaginacion`, v1.0.15):
- Soporte dual `RFC3339` y `YYYY-MM-DD` (ajustando `hasta` en `YYYY-MM-DD` a `23:59:59.999999999 UTC`), rechazo `400` ante formato inválido o rango `desde > hasta` y paginación posterior al filtrado vía `paginacionDesdeQuery` (`limit=10`, `offset=0` por defecto).
- `go test -count=1 ./...` (94 tests top-level en verde), `go vet ./...` y `go build ./...` verdes. Commit `8f4cad4`.

L28: [usuario] "Guarda todo el proceso en la documentacion con todas las dificultades encontradas y el porque se ha construido asi, commitealo y pushealo. y plantea el estado y los todo para mañana continuar con el proyecto."
L28: [asistente] Documentación consolidada con todas las dificultades y decisiones de diseño v1.0.15 (rango inclusivo `YYYY-MM-DD` hasta fin de día UTC, precedencia `404` vs `400`, paginación posterior al filtro sin afectar `GetResumenControlesPaciente`) en `docs/CHANGELOG.md`, `PROJECT_INDEX.md` (3.2, 6.17 y dificultad 21), `AGENTS.md`, `odd/task/estado-actual-proyecto.md` y `todo`.
- Commiteado y pusheado a `origin/main`.
- Planteado el estado actual y los TODOs para mañana (Opción 1: Exportación CSV de controles clínicos e inclusión de `altura` en `ExportPacientesCSV`; Opción 2: Eliminar/editar control clínico puntual `PUT/DELETE /pacientes/:id/controles/:controlId`; Opción 3: Configuración host RDD).

L29: [usuario] "haz el todo 1" (Exportación CSV de controles clínicos `GET /pacientes/:id/controles/export` e inclusión de `altura` en `ExportPacientesCSV`)
L29: [asistente] Implementado test-first en `handlers/paciente.go` (`obtenerControlesFiltrados`, `escribirCSV`, actualización de `ExportPacientesCSV`, `ExportControlesPacienteCSV` y rutas `/pacientes/:id/controles/export` + `{id}/controles/export`) y `handlers/paciente_test.go` (actualizados `TestExportPacientesCSVOK` y `TestExportPacientesCSVVacio`; agregados `TestExportControlesPacienteCSVOK`, `TestExportControlesPacienteCSVFiltroFechasYVacios` y `TestExportControlesPacienteCSVNoEncontradoYSoftDeleted`, v1.0.16).
- `go test -count=1 ./...` (97 tests top-level en verde), `go vet ./...` y `go build ./...` verdes. Commit `986116a`.

L30: [usuario] "Guarda todo el proceso en la documentacion con todas las dificultades encontradas y el porque se ha construido asi, commitealo y pushealo. y plantea el estado y los todo para mañana continuar con el proyecto."
L30: [asistente] Documentación consolidada con todas las dificultades y decisiones de diseño v1.0.16 (`altura` al final en CSV preservando índices `0..5`, cadena vacía `""` cuando falta `altura`/`imc`, extracción de `obtenerControlesFiltrados` y `escribirCSV`, filename estático `controles.csv` por seguridad en cabeceras HTTP y precedencia de ruta en Echo v4) en `docs/CHANGELOG.md`, `PROJECT_INDEX.md` (3.2, 6.18 y dificultad 22), `AGENTS.md`, `odd/task/estado-actual-proyecto.md` y `todo`.
- Commiteado y pusheado a `origin/main`.
- Planteado el estado actual y los TODOs para mañana (Opción 1: Edición y eliminación de controles clínicos puntuales `PUT/DELETE /pacientes/:id/controles/:controlId`; Opción 2: Pipeline CI en GitHub Actions con `go test -race ./...`, `go vet` y `go build`; Opción 3: Configuración host RDD).

L31: [usuario] "haz los 3"
L31: [asistente] Implementadas test-first la Opción 1 (`027f1af`) y la Opción 2 (`b3df81f`), y verificado el bloqueo externo de la Opción 3 (v1.0.17):
- Opción 1 (`027f1af`): agregado `ErrControlNoEncontrado`, `UpdateControl` y `DeleteControl` en `PacienteRepository`, `MockPacienteRepository` y `PacienteFirestoreRepo` (`obtenerControlExistente` con `codes.NotFound`/`!doc.Exists()`), `go mod tidy`, helpers `validarControlClinico`/`calcularIMCControl`, handlers `UpdateControlPaciente` (`200 OK` recalculando IMC y preservando `Fecha` si es cero) y `DeleteControlPaciente` (`204 No Content`), rutas en `SetupRoutes`, y 6 tests nuevos en `handlers/paciente_test.go`, `pkg/firestore/mock_test.go` y `pkg/firestore/repo_test.go` (`103` tests top-level en verde).
- Opción 2 (`b3df81f`): creado `.github/workflows/ci.yml` (`ubuntu-latest`, `permissions: contents: read`, `go vet`, `go test -race -count=1 ./...`, `go build ./...`); verificación independiente read-only `PASS` (riesgo ítem 5).
- Opción 3 (bloqueada / fuera de alcance): constatado en solo lectura que `C:\Users\alvar\.pi\agent\models.json` solo declara `providers` (`strata`, `ollama`) sin lentes RDD y que la configuración de agente/host es externa al repositorio.

L32: [usuario] "Guarda todo el proceso en la documentacion con todas las dificultades encontradas y el porque se ha construido asi, commitealo y pushealo. y plantea el estado y los todo para mañana continuar con el proyecto."
L32: [asistente] Documentación consolidada con todas las dificultades y decisiones de diseño v1.0.17 (`ErrControlNoEncontrado` vs errores internos `500`, preservación de `Fecha` histórica en `PUT` cuando se omite, eliminación de duplicación en helpers compartidos, limitación local de `go test -race` en Windows sin CGO/`gcc` vs runner Linux de CI, y bloqueo externo de `review-readability` en host) en `docs/CHANGELOG.md`, `PROJECT_INDEX.md` (2, 3.2, 4.1, 6, 6.19 y dificultades 23-26 en 6.20), `AGENTS.md`, `odd/task/estado-actual-proyecto.md` y `todo`.
- Planteado el estado actual y los TODOs para mañana (Opción 1: Hardening HTTP y ciclo de vida en `main.go` — `PORT` env var, middlewares `Logger`/`Recover`, timeouts y `graceful shutdown`; Opción 2: Consulta puntual `GET /pacientes/:id/controles/:controlId` o actualización parcial `PATCH /pacientes/:id/controles/:controlId`; Opción 3: Verificación del workflow CI `-race` en GitHub Actions y configuración manual de RDD en el host si se desea reactivar).

L33: [usuario] "haz los todo de hoy"
L33: [asistente] Implementadas test-first la Opción 1 (`161afbf`) y la Opción 2 (`db52369`), y verificada la corrida de CI en GitHub Actions (Opción 3, v1.0.18):
- Opción 1 (`161afbf`): refactorizado `main.go` en `resolveAddr`, `newEcho` (`RequestID` → `Logger` → `Recover`, `HideBanner = true`), `newServer` (timeouts `5s`/`15s`/`15s`/`60s`), `serve` y `runServer` con `signal.NotifyContext` y apagado ordenado de `10s`; agregados 4 tests en `main_test.go` (`50/50` iteraciones sin flakiness y verificación independiente read-only `PASS` por concurrencia / ítem 4).
- Opción 2 (`db52369`): agregado `GetControlByID` en `PacienteRepository`, `MockPacienteRepository` y `PacienteFirestoreRepo` (reutilizando `obtenerControlExistente`), handlers `GetControlPacienteByID` y `PatchControlPaciente` (validación estricta de tipos JSON y recálculo de `IMC`), registro en `SetupRoutes` respetando precedencia estática de `/resumen` y `/export`, y 6 tests nuevos (`113` tests top-level en verde).
- Opción 3: confirmada vía `gh run view 44898033376` la corrida exitosa de `.github/workflows/ci.yml` en `ubuntu-latest` (`1m26s`, incluyendo `go test -race`), detectando aviso de deprecación de Node.js 20 en `actions/checkout@v4` y `actions/setup-go@v5`.

L34: [usuario] "Guarda todo el proceso en la documentacion con todas las dificultades encontradas y el porque se ha construido asi, commitealo y pushealo. y plantea el estado y los todo para mañana continuar con el proyecto."
L34: [asistente] Documentación consolidada con todas las dificultades y decisiones de diseño v1.0.18 (testabilidad determinista de `graceful shutdown` con `net.Listener` efímero y `errCh` bufferizado, orden de middlewares `RequestID`→`Logger`→`Recover`, precedencia estática de `/resumen` y `/export` frente a `GET .../controles/:controlId`, validación estricta de tipos JSON en `PatchControlPaciente` y hallazgo de `notas.(string)` en `PatchPaciente`, más verificación de CI `44898033376` con aviso Node.js 20) en `docs/CHANGELOG.md`, `PROJECT_INDEX.md`, `AGENTS.md`, `odd/task/estado-actual-proyecto.md` y `todo`.
- Planteado el estado actual y los TODOs para mañana (Opción 1: `BodyLimit`, `CORS` configurable y endurecer `PatchPaciente` contra tipos JSON inválidos; Opción 2: actualizar acciones CI a Node.js 24 y sumar `-cover` en `.github/workflows/ci.yml`; Opción 3: endpoint `GET /pacientes/:id/controles/ultimo` para consulta rápida del último control clínico).






