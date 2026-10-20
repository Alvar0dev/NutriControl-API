# CHANGELOG.md

## Versión 1.0.0 - 2026-10-20

### [A] Añadido endpoint `PUT /pacientes/{id}`
- Implementación del handler `UpdatePaciente` que permite actualizar un paciente por ID.
- Ruta `PUT /pacientes/{id}` ahora funcional con validación del ID.
- Se añadió el campo `ID string` al modelo `Paciente` para mantener la referencia única.
- Se implementó el método `Update` en `MockPacienteRepository` (mock) y se preparó la implementación real en `PacienteFirestoreRepository`.

### [B] Mejoras en el modelo `Paciente`
- Se agregó el campo `ID string` con tag JSON `"id,omitempty"` para asegurar que cada paciente tenga un identificador único.

## Pendiente
- Implementar `Update` real en `pkg/firestore/repo.go` (integración con Firestore).
- Implementar `Update` en `pkg/firestore/mock.go` (ya hecho).
- Agregar pruebas unitarias para el endpoint `PUT /pacientes/{id}`.
- Verificar persistencia tras reinicio del servidor.