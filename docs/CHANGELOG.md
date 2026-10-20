# CHANGELOG.md

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