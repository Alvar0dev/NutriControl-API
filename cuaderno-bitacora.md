# Cuaderno de Bitácora — NutriControl API

## Historial de commits (con fechas)
- 2026-10-19 `24a9284` — bitacora: fix missing Patient ID + Update interface (paso hoy)
- 2026-10-19 `1f8aa3d` — feat: add patient handler, firestore repo, tests and mock data
- 2026-10-16 `cb97ca8` — docs: organize SDD structure and persistence layer spec
- 2026-10-07 `fc03609` / `9e6ae4b` — Add files via upload
- 2026-10-07 `ff161c1` — Revise README.md
- 2026-10-06 `6274ee5` — Initial commit

## Avance de hoy (opción 1 — actualizaciones)
- Solo paso inicial de actualización de paciente.
- Agregué `ID string` al modelo `Paciente`.
- Agregué método `Update` a la interfaz `PacienteRepository`.

## Error encontrado y solucionado
- El modelo `Paciente` no tenía campo `ID`, pero `repo.go` asignaba `p.ID = docRef.ID`. Esto provocaba pérdida de referencia al actualizar registros.
- Solución: agregar `ID string` con `json:"id,omitempty"` en `models/paciente.go`.

## Pendiente para mañana
- Implementar `Update` real en `pkg/firestore/repo.go`.
- Implementar `Update` en `pkg/firestore/mock.go`.
- Agregar handler `UpdatePaciente` en `handlers/paciente.go` y ruta `PUT /pacientes/{id}`.
- Subir cambios del paso pequeño.
