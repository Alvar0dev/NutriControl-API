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
