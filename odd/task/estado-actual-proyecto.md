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
