# Plan de Continuación — NutriControl API

Hoja de ruta operativa para retomar las siguientes sesiones sin reprocesar análisis ni arquitectura. Parte del estado posterior a **v1.0.23** (`API_KEY` en commit `de5c6e5`, CI v1.0.22 run `45982386942` PASS) tras completar **Receta A / Variante 3A (`API_KEY`)**.

---

## 1. Orden recomendado para los próximos días

| Prioridad | Qué hacer | Riesgo ODD / Verificador | Notas clave |
|---|---|---|---|
| **Completada (`de5c6e5`)** | **Autenticación HTTP en `main.go` (`API_KEY`)** | `Risk: item 2` (seguridad) — auditoría `PASS` | Cubre toda la API clínica (`/pacientes*`), exime `OPTIONS`, `GET /health` y `GET /`. *(Si en el futuro hay roles/multi-usuario, evaluar Variante 3B JWT abajo).* |
| **Día siguiente (1.º)** | **Paginación con cursores `StartAfter` en Firestore** (`pkg/firestore/` + `handlers/`) | `Risk: item 3` (ampliación de puerto y contrato HTTP) → verificador read-only | Mayor impacto: exige extender `PacienteRepository`, `MockPacienteRepository` y `fakeFirestoreServer.RunQuery` (`pkg/firestore/repo_test.go`). |
| **Opcional / higiene** | **Concurrencia en `MockPacienteRepository`** (`sync.RWMutex` en `pkg/firestore/mock.go`) y/o **`IPExtractor` + tope `RPS`** | `Risk: item 4` / `item 2` → verificador read-only | Cierra la deuda técnica #3 de `docs/arquitectura-y-defensa.md` y endurece detección de IP real si corre sin proxy confiable. |

Cadena activa de middlewares en `newEchoWithConfig` (`main.go`):

```text
RequestID → Logger → Recover → CORS (opcional) → RateLimit (opcional) → KeyAuth (opcional) → BodyLimit
```

---

## 2. Receta A — Autenticación HTTP (`API_KEY` ✅ implementada en `de5c6e5`; JWT opcional futuro)

- **Implementado (`v1.0.23`, `de5c6e5`):** `resolveAPIKey`, `KeyAuthWithConfig` opt-in (`Authorization: Bearer` y `X-API-Key`), `ConstantTimeCompare`, `"X-API-Key"` en `CORSConfig.AllowHeaders`, `ErrorHandler` `401` JSON y exención únicamente de `OPTIONS`, `GET /health` y `GET /`.
- **Evolución futura (solo si aparecen múltiples usuarios/roles):** migrar o convivir con **3B — JWT** (`golang-jwt/jwt/v5`, `HS256` o JWKS, claims `sub`/`role`), o comparar digest SHA-256 antes de `subtle.ConstantTimeCompare` si se requiere ocultar incluso la longitud de la clave.

---

## 3. Receta B — Paginación con cursores `StartAfter` en Firestore

### Decisión previa de contrato HTTP

- Mantener `GetAll(ctx, limit, offset)` intacto para no romper `ReportePacientes`, `StatsPacientes`, `SearchPacientes`, `CountPacientes`, `ExportPacientesCSV` ni `GetPacientesPorEdad`.
- Añadir soporte de cursor de forma **aditiva** para evitar breaking changes:
  - **En repositorio (`pkg/firestore/paciente_repo.go`):** `GetPage(ctx context.Context, limit int, cursor string) ([]models.Paciente, string, error)`.
  - **En HTTP (`handlers/paciente.go`):** elegir al arrancar la sesión entre:
    - **Opción HTTP 1 (compatible 100%):** `GET /pacientes?cursor=<id>&limit=<n>` devolviendo el array `[Paciente]` y exponiendo el siguiente cursor en cabecera `X-Next-Cursor` (o solo cuando se pasa `cursor`).
    - **Opción HTTP 2 (endpoint dedicado):** `GET /pacientes/page?cursor=<id>&limit=<n>` devolviendo `{"items": [...], "next_cursor": "..."}`. *(Evita cambiar el esquema JSON de `GET /pacientes`.)*

### Invariantes de implementación en Firestore y Mock

1. **Orden e índice sin índices compuestos:**
   - Usar `r.col.OrderBy(firestore.DocumentID, firestore.Asc)` en `PacienteFirestoreRepo`. Esto usa el índice automático de `__name__` en Firestore sin requerir despliegue previo de índices compuestos.
2. **Soft delete retrocompatible:**
   - No filtrar `where("deleted_at", "==", nil)` en Firestore porque los documentos creados antes de v1.0.13 no tienen el campo `DeletedAt` y Firestore los excluiría de la consulta.
   - Leer por páginas internas acotadas (ej. `StartAfter` desde el último documento leído cuando `cursor != ""`), filtrar `p.DeletedAt != nil` en Go hasta reunir `limit + 1` activos, devolver hasta `limit` elementos y usar el `ID` del último elemento devuelto como `nextCursor` solo si hubo un `(limit + 1)`-ésimo activo (evita falsos positivos cuando quedan 0 activos después).
3. **Paridad con `MockPacienteRepository` (`pkg/firestore/mock.go`):**
   - Ordenar activos por `p.ID` ascendente (`sort.Slice`), ubicar el primer ID `>` `cursor` (o devolver error/vacío definido si el cursor no existe — definir simetría con Firestore), aplicar `limit` y retornar `nextCursor`.
4. **Extensión de `fakeFirestoreServer.RunQuery` (`pkg/firestore/repo_test.go`):**
   - Implementar soporte para `StructuredQuery.OrderBy` (`__name__`), `StructuredQuery.StartAt` (`Before == false` equivalente a `StartAfter`) y `StructuredQuery.Limit` en el servidor gRPC in-process.
5. **Tests (`pkg/firestore/mock_test.go`, `pkg/firestore/repo_test.go`, `handlers/paciente_test.go`):**
   - Primera página con `next_cursor`, segunda página sin duplicar el cursor, última página con `next_cursor == ""`, saltos sobre pacientes soft-deleted en medio de la página, y `limit` por defecto/inválido.
6. **Commit sugerido:** `feat: paginación con cursor StartAfter en Firestore y API`.

---

## 4. Cómo reflejar cada cierre en la documentación (Checklist ODD)

Siguiendo `docs/organizacion-documentacion.md`, **primero commitear las unidades de código** (para capturar hashes y cobertura reales), verificar, y luego hacer **un único commit `docs:`**:

- [ ] **Verificación previa al cierre:**
  - `go mod tidy` (si cambió `go.mod`), `go vet ./...`, `go build ./...`, `go test -cover -count=1 ./...`.
  - Si el cambio toca seguridad (`item 2`), contratos públicos (`item 3`), concurrencia (`item 4`) o entrega/deps (`item 5`): ejecutar subagente auditor read-only (`research`) y aislar su commit.
  - Recordar que `go test -race` no corre en Windows local sin `gcc`; validar en CI (`ubuntu-latest`) tras el push.
- [ ] **`docs/CHANGELOG.md`:** nueva sección de versión (`v1.0.22+`) con features, hashes exactos, cobertura/tests y bloque **"Dificultades encontradas y por qué se construyó así"**.
- [ ] **`PROJECT_INDEX.md`:** actualizar árbol (si aplica), §3.1/§3.2/§4.1 según corresponda, tabla §5.4 de variables de entorno, tabla §6 de estado, agregar subsección `§6.xx` de feature, desplazar/renumerar `### Dificultades acumuladas y decisiones de diseño` (actualmente `§6.26`) y añadir los nuevos aprendizajes.
- [ ] **`AGENTS.md`:** resumen operativo corto en `Current Architecture` (sin duplicar el detalle del changelog).
- [ ] **`docs/arquitectura-y-defensa.md`:** actualizar conteo total de tests, sección 4 (evolución), sección 5 (Q&A afectada) y **limpiar §6 Deuda técnica** retirando lo resuelto.
- [ ] **`odd/task/estado-actual-proyecto.md`:** añadir filas `Txx` con hashes en `## Tasks` y entradas `Lxx` en `## Log`.
- [ ] **`todo` y este `plan.md`:** marcar lo completado en `todo`, actualizar fecha/estado base y retirar o marcar como hechas las recetas ejecutadas en `plan.md`.
- [ ] **No tocar:** `README.md` (bitácora histórica congelada) ni crear `odd/tasks/<feature>.md` salvo tareas `large` que no quepan en un `git diff`.
- [ ] **Engram (`nutricontrol-api`):** `mem_save` con las decisiones clave, actualizar `nutricontrol-api/estado-para-manana` y ejecutar `mem_session_summary` antes de cerrar.
