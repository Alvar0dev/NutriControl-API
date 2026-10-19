# Cuaderno de Bitácora — NutriControl API

Fecha: 19 oct 2026

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
