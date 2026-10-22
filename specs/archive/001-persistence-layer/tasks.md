# Tareas — Capa de Persistencia con Firestore

## Dependencias: RF-01 → RF-02 → RF-03 → RF-04. Las tareas siguen orden secuencial.

- [ ] **T1** [Base] — Configurar archivo `go.mod` con `cloud.google.com/go/firestore`. (Cubre RF-01/02) | ~20 min
- [ ] **T2** [Repo] — Crear `pkg/firestore/client.go`: inicializar cliente Firestore (modo archivo local / emulador). (Cubre RF-01/04) | ~25 min
- [ ] **T3** [Repo] — Crear `pkg/firestore/paciente_repo.go`: interfaz `PacienteRepository` con `Create()` y `GetAll()`. (Cubre RF-01/02) | ~25 min
- [ ] **T4** [Repo] — Implementar `Create()`: guardar documento, devolver `id` generado automáticamente. (Cubre RF-01, RF-04) | ~20 min
- [ ] **T5** [Repo] — Implementar `GetAll()`: recuperar todos los documentos, devolver `[]models.Paciente`. (Cubre RF-02, RF-04) | ~20 min
- [ ] **T6** [Mod] — Modificar `handlers/paciente.go`: eliminar datos mockeados, recibir repositorio, usar `repo.Create()` / `repo.GetAll()`. (Cubre RF-01/02) | ~20 min
- [ ] **T7** [Validación] — Añadir validación básica en `CreatePaciente`: comprobar campos obligatorios, devolver `400` con mensaje. (Cubre RF-03) | ~15 min
- [ ] **T8** [Error] — Añadir captura de errores: `try/catch`-equivalente en Go, devolver `500` con descripción. (Cubre RF-04) | ~15 min
- [ ] **T9** [Integración] — Modificar `main.go`: inicializar repositorio, pasarlo a los handlers (inyección). (Cubre RF-01/02) | ~15 min
- [ ] **T10** [Verificación] — Prueba manual con `curl`: `POST /pacientes`, luego `GET /pacientes`; verificar persistencia tras reinicio. (Cubre RF-01/02/03/04) | ~25 min
