# Organización de documentación — NutriControl API

Este documento explica cómo está organizada la documentación del proyecto y **por qué** se organizó así, siguiendo las buenas prácticas de **Organic Driven Development (ODD)**.

## Principio rector: ODD no exige burocracia de documentos

ODD distingue por tamaño de tarea:

| Tamaño | Criterio | Documentos requeridos |
|--------|----------|----------------------|
| **Small** | Entendido, riesgo contenido, resumible desde `git diff` | **Ninguno obligatorio**. `todo` opcional. |
| **Large** | Varias sesiones, entrega separada, requisitos que no caben en el diff | **Un solo documento de feature**: `odd/tasks/<feature>.md` + espejo en Engram. |

De esto se deriva que **no todos los documentos son necesarios siempre**. Solo se crea lo que el tamaño de la tarea realmente requiere.

## Mapa actual de documentos

| Documento | Función | ¿Editar por sesión? |
|-----------|---------|---------------------|
| `PROJECT_INDEX.md` | **Único documento vivo**: navegación de carpetas, comandos de build/test/run, contratos de API, estado actual del proyecto. | Solo cuando cambia la estructura o los comandos. |
| `AGENTS.md` | Contexto operativo breve para arranque rápido de nuevas sesiones (stack, estructura, comandos clave, límites). | Solo cuando cambia el setup o las convenciones. |
| `README.md` | Bitácora histórica día a día de lo hecho, dificultades y soluciones. | **No editar por sesión.** Es memoria histórica, no documentación operativa. |
| `docs/constitution.md` | Principios innegociables del proyecto (hexagonal, testing purista, minimalismo). | Solo cuando cambian los principios. |
| `docs/CHANGELOG.md` | Historia de cambios contractuales. | Solo al cerrar versiones o hitos. |
| `odd/task/estado-actual-proyecto.md` | **Único archivo de definición de estado actual** del proyecto (contrato ODD S3). Punto de partida para la siguiente sesión. | Solo cuando cambia el estado base del proyecto. |
| `odd/tasks/<feature>.md` | Documento de **feature tracking** para trabajos grandes. Se crea antes del primer write y se cierra al terminar la feature. | Solo durante una feature activa y grande. |
| `odd/tasks/archive/` | Feature trackings ya cerrados (ej. `000-reestructuracion-documental.md`). Se mueven aquí al terminar; no se editan. | Nunca. |
| `todo` (raíz) | Checklist ligero de la **sesión actual**. | Sí, se actualiza a medida que avanzan las tareas de la sesión. |
| `.gitignore` | Ignora binarios compilados (`*.exe`, `*.dll`) y config local de Gentle AI (`.atl/`). | Solo cuando haya que ignorar algo nuevo. |

## ¿Por qué se eliminó o fusionó cada documento?

### 1. `DEV_RUNBOOK.md` → fusionado en `PROJECT_INDEX.md`
Antes teníamos dos documentos que repetían comandos de build, test, variables de entorno y estado actual. Eso generaba desincronización: uno decía una cosa y el otro otra.  
**Decisión:** se fusionaron en un solo `PROJECT_INDEX.md` que es la fuente única de navegación + comandos + estado. El antiguo `DEV_RUNBOOK.md` se movió a `docs/archive/DEV_RUNBOOK.md` por si se necesita consulta histórica.

### 2. `odd/tasks/000-reestructuracion-documental.md` → movido a `odd/tasks/archive/`
Era el feature tracking de una tarea ya cumplida (reestructuración documental ODD). Ya no es necesario seguirlo activo.  
**Decisión:** se archivó en `odd/tasks/archive/`. No se elimina (se conserva el historial), pero deja de ser un documento vivo.

### 3. `README.md` dejado como histórico, no operativo
Tenía una bitácora día a día mezclada con instrucciones de uso. Eso lo hacía confuso como referencia operativa.  
**Decisión:** se conserva como memoria histórica del proyecto, pero no se usa como documento de trabajo. Para consultar cómo se builda o testa, se usa `PROJECT_INDEX.md`.

### 4. `AGENTS.md` simplificado a contexto operativo breve
Antes mezclaba contexto por día con comandos y límites. Ahora se limita a lo que un nuevo agente o sesión necesita saber en los primeros minutos: stack, estructura, comandos clave, convenciones y límites. El detalle de estado vive en `PROJECT_INDEX.md` y `odd/task/estado-actual-proyecto.md`.

## ¿Cuándo se crea un documento nuevo?

Solo dos casos justifican crear documentación nueva:

1. **Arranca una feature grande (large):** creás `odd/tasks/<feature-name>.md` con `## Specs`, `## Tasks` y `## Log`, más su espejo en Engram. Ese documento es el specification que leen los subagentes. Se cierra cuando la feature termina.
2. **Cambian los principios o la estructura del proyecto:** se actualiza `docs/constitution.md` o `PROJECT_INDEX.md`.

En todo lo demás (trabajo small), no se crea documento ninguno: se toca código, se corre `go test ./...`, y se commitea.

## Flujo de trabajo resumido

```
Nueva sesión
    │
    ▼
Leer odd/task/estado-actual-proyecto.md  (estado base)
    │
    ▼
¿Trabajo small o large?
    │
    ├─ small ──► editar código, go test ./..., commitear. Sin documentos nuevos.
    │
    └─ large ──► crear odd/tasks/<feature>.md (Specs, Tasks, Log)
                   │
                   ▹ trabajar tarea por tarea
                   │
                   ▼
                 cerrar feature, mover a odd/tasks/archive/
```

## Resumen de la simplificación

- **Antes:** 3 documentos de navegación/estado (`PROJECT_INDEX.md`, `DEV_RUNBOOK.md`, `AGENTS.md`) + 1 feature tracking activo + bitácora en `README.md`.
- **Ahora:** 1 documento vivo de navegación/estado (`PROJECT_INDEX.md`) + 1 contexto breve (`AGENTS.md`) + 1 estado ODD (`odd/task/estado-actual-proyecto.md`) + 1 histórico (`README.md`) + feature trackings solo cuando hay una feature grande activa.

Esto reduce la superficie de documentación que hay que mantener sincronizada y evita que el equipo duplique información en dos o tres sitios.