# CHANGELOG.md

## Versión 1.0.22 - 2026-11-27

### Rate limiting opcional por IP en `main.go` y hoja de ruta `plan.md`

- **Rate limiting opcional por IP (`main.go`, `main_test.go`, `go.mod`)**:
  - `resolveRateLimit(rpsEnv, burstEnv)`: lee `RATE_LIMIT_RPS` y `RATE_LIMIT_BURST` de forma segura; valores vacíos, `<= 0`, no numéricos o `NaN`/`Inf` dejan el limitador desactivado sin `panic`, mientras un burst ausente o `<= 0` usa `max(1, ceil(rps))`.
  - `newEchoWithConfig`: registra `middleware.RateLimiterWithConfig` + `middleware.NewRateLimiterMemoryStoreWithConfig` (`defaultRateLimitTTL = 3m` sobre `golang.org/x/time/rate`) en la cadena `RequestID` → `Logger` → `Recover` → `CORSWithConfig` (opcional) → `RateLimiterWithConfig` (opcional) → `BodyLimit`.
  - `Skipper` exime preflights `OPTIONS` y `GET /health` (`c.Path() == "/health"`); `DenyHandler` responde `echo.NewHTTPError(http.StatusTooManyRequests, map[string]string{"error": "Demasiadas peticiones, intente nuevamente más tarde"})` conservando cabeceras CORS en `429` y emitiendo `NoContent` en `HEAD`.
  - `go.mod`: `golang.org/x/time v0.16.0` promovido a `require` directo vía `go mod tidy` sin nuevas dependencias externas.
  - Tests en `main_test.go`: actualizado `TestLoadServerConfig` y añadidos `TestResolveRateLimit` (12 subcasos) y `TestNewEchoRateLimit` (`126` tests top-level en verde: `nutricontrol` `67.3%`, `handlers` `88.3%`, `models` `[no statements]`, `pkg/firestore` `85.0%`; `go vet ./...` y `go build ./...` OK; auditoría independiente read-only de seguridad `PASS`).
- **Hoja de ruta y verificación previa (`plan.md`, CI `v1.0.21`)**:
  - Creado `plan.md` en la raíz con el plan técnico para las próximas sesiones (Receta A autenticación HTTP `API_KEY`/`JWT`, Receta B paginación `StartAfter` en Firestore + `fakeFirestoreServer.RunQuery`, concurrencia opcional en mock y checklist ODD).
  - Confirmada vía `gh run list` / `gh run view 45799288074` la corrida en verde de CI `v1.0.21` (`564c1c3`, `39s` en `ubuntu-latest` con Go `1.26.1` y `-race`).
- Commit: `76fdb3b` (`feat: agregar rate limiting opcional por IP en main.go`).

**Dificultades encontradas y por qué se construyó así:**

1. **Orden `CORS` → `RateLimit` → `BodyLimit` y uso de `c.Path() == "/health"` en `Skipper`.** Al colocar `CORSWithConfig` antes de `RateLimiterWithConfig`, Echo escribe `Access-Control-Allow-Origin` antes de evaluar la cuota por IP, evitando que clientes web confundan un `429` con un bloqueo CORS; ubicar `BodyLimit` detrás evita leer el cuerpo en peticiones ya rechazadas. Además, como `e.Use(...)` se ejecuta tras `Router.Find`, `c.Path()` ya viene resuelto y permite eximir `/health` y `OPTIONS` sin abrir rutas inexistentes (`c.Path() == ""`).
2. **`echo.NewHTTPError(429, map[string]string{...})` frente a `c.JSON` directo en `DenyHandler`.** `RateLimiterWithConfig` delega el rechazo llamando a `c.Error(config.DenyHandler(...))`. Devolver `*echo.HTTPError` con un mapa evita que `DefaultHTTPErrorHandler` envuelva la respuesta en `{"message": ...}`, mantiene uniforme `{"error": ...}`, registra el error en `Logger` y delega a Echo enviar `NoContent(429)` ante `HEAD`.
3. **Configuración opt-in sin dependencias nuevas ni ruptura de tests.** Mantener el limitador apagado por defecto cuando falta `RATE_LIMIT_RPS` (igual que `CORS_ALLOWED_ORIGINS`) evita romper entornos locales o suites existentes, y promueve a directa la librería `golang.org/x/time/rate` que Echo ya incluía de forma transitiva.
4. **Supuestos operacionales detectados en la auditoría de seguridad (`c.RealIP()` y memoria por instancia).** Sin `e.IPExtractor` explícito, `c.RealIP()` confía en `X-Forwarded-For`/`X-Real-IP` (adecuado tras el balanceador de Cloud Run, pero susceptible a spoofing si la API se expusiera directamente sin proxy confiable), y `RateLimiterMemoryStore` mantiene cuotas por réplica con TTL de `3m` sin cabecera `Retry-After` (un `RATE_LIMIT_RPS` astronómico > `math.MaxInt` también desbordaría `int(math.Ceil(rps))`). Estas limitaciones quedan documentadas en `docs/arquitectura-y-defensa.md`.

---
## Versión 1.0.21 - 2026-11-26

### Filtrado temporal en resumen de controles (`?desde=&hasta=`), verificación CI v7 y `Dockerfile` multi-stage

- **Filtrado temporal opcional en `GET /pacientes/:id/controles/resumen` (`handlers/paciente.go`, `handlers/paciente_test.go`)**:
  - `GetResumenControlesPaciente` reutiliza `obtenerControlesFiltrados(c, repo)` para aceptar `?desde=&hasta=` (`YYYY-MM-DD` inclusivo hasta `23:59:59.999999999 UTC` o `RFC3339`), manteniendo retrocompatibilidad total cuando no se envían query params.
  - Rango sin coincidencias devuelve `200 OK` con `total_controles: 0` (omitiendo `fecha_inicio` y `fecha_fin` vía `omitempty`), rango de un control deja deltas en `0`, paciente inexistente o soft-deleted responde `404` y fechas inválidas o rango invertido (`desde > hasta`) responden `400`.
  - Tests nuevos: `TestGetResumenControlesPacienteFiltroFechas` (5 subcasos) y `TestGetResumenControlesPacienteFiltroFechasInvalidas` (3 subcasos).
- **Verificación de CI v7 y contenerización minimalista (`Dockerfile`, `.dockerignore`)**:
  - Confirmada vía `gh run view 45597881616` la corrida verde de `.github/workflows/ci.yml` sobre `v1.0.20` (`bd622f1`, `1m24s` en `ubuntu-latest`, Go `1.26.1`, `actions/checkout@v7` y `actions/setup-go@v7`, `go test -race -cover`, sin anotaciones de deprecación de Node.js 20).
  - Creado `Dockerfile` multi-stage: compilación estática Linux (`golang:1.26-alpine`, `CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w"`) sobre runtime `gcr.io/distroless/static-debian12:nonroot` (`USER nonroot:nonroot`, `ENV PORT=8080`, `EXPOSE 8080`, `ENTRYPOINT ["/nutricontrol"]`).
  - Creado `.dockerignore` excluyendo `.git/`, `.github/`, `.gitignore`, `.atl/`, `odd/`, `docs/`, `specs/`, `todo`, `*.md`, `*.exe`, `*.dll`, `Dockerfile` y `.dockerignore`.
  - Validados suite Go (`124` tests top-level en verde: `nutricontrol` `59.1%`, `handlers` `88.3%`, `models` `[no statements]`, `pkg/firestore` `85.0%`; `go vet ./...` y `go build ./...` OK), `docker build` + smoke test de contenedor (`22.049.372 bytes`, `User: nonroot:nonroot`, `GET /health` `200`, `GET /` `200`, `POST /pacientes` `201` y cierre limpio por `SIGTERM` con `"Servidor detenido correctamente"`) y verificación independiente read-only `PASS`.
- Commits: `92f26e3` (`feat: soportar filtro temporal desde/hasta en resumen de controles`) y `714f97a` (`build: agregar Dockerfile multi-stage y .dockerignore`).

**Dificultades encontradas y por qué se construyó así:**

1. **Reutilización directa de `obtenerControlesFiltrados` frente a duplicación en `GetResumenControlesPaciente`.** Antes `GetResumenControlesPaciente` repetía la validación de `:id`, `repo.GetByID` y `repo.GetControles`. Sustituir ese bloque por `obtenerControlesFiltrados` (ya compartido por `GetControlesPaciente` y `ExportControlesPacienteCSV`) unifica la precedencia `404` antes que `400`, el ajuste de `hasta` (`YYYY-MM-DD`) a fin de día UTC y el rechazo de rangos invertidos, sin aplicar paginación al resumen agregado.
2. **Semántica `200 OK` con `total_controles: 0` cuando un rango no devuelve controles.** A diferencia de `GET /pacientes/:id/controles/ultimo` (un recurso puntual que devuelve `404`), `/resumen` representa un agregado evolutivo; responder `200 OK` con `total_controles: 0` y sin `fecha_inicio`/`fecha_fin` mantiene coherencia con el caso base sin controles y evita que un filtro temporal vacío rompa consumidores gráficos.
3. **Elección de `gcr.io/distroless/static-debian12:nonroot` y `ENTRYPOINT` en forma exec.** Un binario Go `CGO_ENABLED=0` podría correr sobre `scratch`, pero el adaptador de Cloud Firestore requiere certificados CA (`/etc/ssl/certs`) para TLS/gRPC y datos de zona horaria además de ejecución sin privilegios `root`. `distroless/static-debian12:nonroot` aporta esa base en ~22 MB sin shell ni utilidades innecesarias, y `ENTRYPOINT ["/nutricontrol"]` en forma exec garantiza que `docker stop` entregue `SIGTERM` directamente al proceso PID 1 (`signal.NotifyContext` en `main.go`).
4. **Aislamiento de contexto Docker y separación del commit `build:` (`Risk: item 5`).** Excluir `.git/`, documentación y binarios Windows en `.dockerignore` evita invalidar la caché de `go mod download` al editar docs y mantiene fuera del contexto de build archivos ajenos al runtime; commitear `Dockerfile`/`.dockerignore` aparte del cambio funcional facilita auditoría y rollback independiente de infraestructura.

---
## Versión 1.0.20 - 2026-11-25

### Hardening de entrada HTTP (`BodyLimit` y `CORS`), endurecimiento de tipos en `PatchPaciente` y propagación estricta de `413`

- **Endurecimiento de `PatchPaciente` y unificación de helpers JSON (`handlers/paciente.go`)**:
  - Extraídos `manejarBindError`, `bindJSONStrict`, `leerBodyPatchJSON`, `extraerStringJSON` y `extraerFloatJSON`.
  - `PatchPaciente`: rechaza payload `null` (`400`), exige tipos JSON exactos en todos los campos (`nombre`, `edad`, `peso`, `porcentaje_grasa`, `altura`, `notas`), rechaza decimales en `edad` (`e != math.Trunc(e)`) y valores `NaN`/`Inf`, eliminando el panic de `notas.(string)` sin `, ok` y los `200 OK` silenciosos ante tipos incorrectos.
  - `bindJSONStrict` sustituye `c.Bind` en `CreatePaciente`, `CreatePacientesBatch`, `UpdatePaciente`, `CreateControlPaciente` y `UpdateControlPaciente` para leer con `io.ReadAll` antes de decodificar JSON y propagar `413` con `manejarBindError`.
- **Middleware `BodyLimit` y `CORS` configurable en `main.go`**:
  - `serverConfig`, `loadServerConfig`, `resolveBodyLimit` (validando `BODY_LIMIT` con `github.com/labstack/gommon/bytes.Parse`, default `"1M"`, revirtiendo valores `<= 0` o inválidos sin panic) y `parseCORSOrigins` (`CORS_ALLOWED_ORIGINS` separado por comas).
  - `newEchoWithConfig`: registra `RequestID` → `Logger` → `Recover` → `CORSWithConfig` (solo si `len(CORSAllowedOrigins) > 0`, `AllowCredentials: false`, seguro por defecto) → `BodyLimit`, e instala `strictJSONSerializer` en `e.JSONSerializer`.
  - `go.mod`: `github.com/labstack/gommon v0.5.0` promovido a `require` directo vía `go mod tidy`.
- **Tests (`handlers/paciente_test.go`, `main_test.go`)**:
  - Agregados `TestPatchPacienteTiposInvalidos`, `TestResolveBodyLimit`, `TestParseCORSOrigins`, `TestLoadServerConfig`, `TestNewEchoBodyLimit` (cubre `Content-Length` y streaming `ContentLength = -1` en los 7 endpoints de escritura), `TestNewEchoCORS` y `TestStrictJSONSerializerDeserializePropaga413` (`122` tests top-level en verde: `nutricontrol` `59.1%`, `handlers` `88.1%`, `pkg/firestore` `85.0%`; `go test -count=50 .` OK; verificación independiente read-only `PASS`).
- **Documentación de arquitectura y defensa (`docs/arquitectura-y-defensa.md`)**:
  - Creada la guía técnica con mapa de capas, decisiones/tradeoffs, preguntas probables de defensa con evidencia exacta y deuda técnica vigente, enlazada en `PROJECT_INDEX.md` y `docs/organizacion-documentacion.md`.
- Commits: `32e92d3` (`fix: endurecer tipos JSON en PatchPaciente y lectura estricta bindJSONStrict`) y `86e94d9` (`feat: agregar BodyLimit, CORS configurable y strictJSONSerializer en main.go`).

**Dificultades encontradas y por qué se construyó así:**

1. **Sustitución de `json.NewDecoder` por lectura estricta (`strictJSONSerializer` y `bindJSONStrict`) para impedir bypass de `BodyLimit` en Go `<= 1.26`.** Durante la auditoría independiente de seguridad, se detectó que `echo.DefaultJSONSerializer` usa `json.NewDecoder(req.Body).Decode`, el cual en `encoding/json` v1 (`Go <= 1.26`, versión declarada en `go.mod` y usada por el runner de CI) escanea el buffer antes de comprobar `err != nil`: si `middleware.BodyLimit` (`limitedReader.Read`) devuelve `(n, echo.ErrStatusRequestEntityTooLarge)` junto con el cierre `}`/`]` de un payload streaming (`ContentLength = -1`), `json.Decoder` podía retornar `nil` descartando el `413`. Al leer localmente con Go `1.27.2` (`json/v2` por defecto) el test pasaba igual, lo que habría ocultado la diferencia con CI. Con `strictJSONSerializer` en `main.go` y `bindJSONStrict` en `handlers/paciente.go` (más `TestStrictJSONSerializerDeserializePropaga413` usando `partialLimitReader`), se garantiza `413` determinista en ambas versiones de Go.
2. **Validación previa de `BODY_LIMIT` para evitar `panic` al iniciar Echo.** `middleware.BodyLimit` lanza `panic` si recibe un string inválido. Validar `BODY_LIMIT` en `resolveBodyLimit` con la misma función `bytes.Parse` de Echo (y rechazar `<= 0`) asegura que una variable mal configurada registre un aviso y caiga de forma segura en `"1M"`.
3. **CORS desactivado por defecto en lugar del `*` predeterminado de Echo.** `middleware.CORSWithConfig` con `AllowOrigins` vacío aplica por defecto `["*"]`. Condicionar su registro a `len(cfg.CORSAllowedOrigins) > 0` con `AllowCredentials: false` mantiene política restrictiva same-origin cuando `CORS_ALLOWED_ORIGINS` no está definida.
4. **Corrección integral de `PatchPaciente` al resolver `notas.(string)`.** En la fase RED de `TestPatchPacienteTiposInvalidos` se confirmó el `panic` al recibir `{"notas": 123}`/`{"notas": null}`, junto con respuestas `200 OK` silenciosas para `null` o tipos inválidos en el resto de campos y truncamiento silencioso de decimales en `edad`. Extraer `extraerStringJSON` y `extraerFloatJSON` unificó las guardas estrictas entre `PatchPaciente` y `PatchControlPaciente`.

---
## Versión 1.0.19 - 2026-11-24

### Actualización de CI a Node.js 24 con reporte de cobertura y endpoint `GET /pacientes/:id/controles/ultimo`

- **Pipeline CI en GitHub Actions (`.github/workflows/ci.yml`)**:
  - Actualizados `actions/checkout@v4` → `actions/checkout@v7` (`v7.0.1`) y `actions/setup-go@v5` → `actions/setup-go@v7` (`v7.1.0`), tras verificar vía `gh api` sus `action.yml` confirmando runtime `node24` y compatibilidad con `go-version-file: go.mod`, eliminando el aviso de deprecación de Node.js 20.
  - Añadido `-cover` al paso de tests: `go test -race -cover -count=1 ./...`.
  - Commit: `894c816 ci: actualizar checkout y setup-go a v7 (Node 24) y sumar cobertura`.
- **Nuevo endpoint `GET /pacientes/:id/controles/ultimo` (`handlers/paciente.go`)**:
  - Handler `GetUltimoControlPaciente`: valida `id` (`400`), verifica paciente activo (`404`, excluyendo registros soft-deleted), obtiene controles cronológicos vía `repo.GetControles` (`500` ante fallo del repositorio), responde `404 Not Found` explícito si el paciente aún no tiene controles registrados (`"No se encontraron controles clínicos para el paciente con ID " + id`) y devuelve `200 OK` con `controles[len(controles)-1]`.
  - Rutas `/pacientes/:id/controles/ultimo` y `/pacientes/{id}/controles/ultimo` registradas en `SetupRoutes` antes de `/pacientes/:id/controles/:controlId` y `/pacientes/{id}/controles/{controlId}`.
  - Tests nuevos en `handlers/paciente_test.go`: `TestGetUltimoControlPacienteOK` (orden cronológico independiente de inserción, actualización tras borrar el más reciente y no colisión con `:controlId`) y `TestGetUltimoControlPacienteErrores` (table-driven: paciente sin controles, inexistente y soft-deleted) — `115` tests top-level en verde (`nutricontrol` `45.0%`, `handlers` `87.7%`, `pkg/firestore` `85.0%`).
  - Commit: `c89e0f5 feat: endpoint GET /pacientes/:id/controles/ultimo`.

**Dificultades encontradas y por qué se construyó así:**

1. **Colisión de Echo v4 entre `/controles/ultimo` y `/controles/:controlId` durante la fase RED.** Al correr `TestGetUltimoControlPaciente` antes de registrar la ruta estática `/ultimo`, Echo v4 capturó `"ultimo"` como parámetro `:controlId` llamando a `GetControlPacienteByID` y devolviendo `404` con `"No se encontró control clínico con ID ultimo"`. Ubicar `/pacientes/:id/controles/ultimo` inmediatamente después de `/export` y antes de `/:controlId` garantizó la prioridad estática del router.
2. **Respuesta `404` frente a struct cero cuando un paciente activo no tiene controles.** A diferencia de `GET /pacientes/:id/controles/resumen` (un agregado estadístico donde `total_controles: 0` es válido), `/ultimo` representa un único recurso `ControlClinico`. Devolver `200 OK` con un struct vacío (`0001-01-01T00:00:00Z`, `peso: 0`) induciría datos clínicos falsos en consumidores; responder `404 Not Found` con mensaje explícito mantiene coherencia REST con `GET /controles/:controlId`.
3. **Verificación contractual de runtime Node.js 24 en acciones de GitHub antes de actualizar el workflow.** En lugar de adivinar tags mayores de `actions/checkout` y `actions/setup-go`, se consultaron los últimos releases (`v7.0.1` y `v7.1.0`) y sus `action.yml` reales vía `gh api`, confirmando `runs.using: node24` y preservación de `go-version-file: go.mod`, validado además por verificación independiente read-only (`PASS`).

---
## Versión 1.0.18 - 2026-11-23

### Hardening HTTP y ciclo de vida en `main.go`, consulta/edición parcial de controles (`GET/PATCH /pacientes/:id/controles/:controlId`) y verificación de CI

- **Hardening HTTP y graceful shutdown (`main.go`, `main_test.go`)**:
  - `resolveAddr(portEnv)`: soporte de variable de entorno `PORT` con normalización (`"9090"` → `":9090"`, `":3000"` preservado, default `":8080"`).
  - `newEcho(repo)`: registra middlewares operativos de Echo v4 en orden `RequestID` → `Logger` → `Recover` (`HideBanner = true`), rutas de dominio vía `handlers.SetupRoutes` y ruta raíz `GET /`.
  - `newServer(addr, repo)`: configura `*http.Server` con timeouts explícitos (`ReadHeaderTimeout: 5s`, `ReadTimeout: 15s`, `WriteTimeout: 15s`, `IdleTimeout: 60s`).
  - `serve` y `runServer`: ciclo de vida con `signal.NotifyContext` (`os.Interrupt`, `syscall.SIGTERM`), apagado ordenado `srv.Shutdown` con timeout de `10s` y drenado garantizado de `errCh` (buffer `1`) para evitar fugas de goroutines.
  - Tests nuevos en `main_test.go`: `TestResolveAddr`, `TestNewEchoMiddlewaresYRoutes` (verificando `X-Request-ID` y captura de `panic` → `500`), `TestNewServerTimeouts` y `TestServeGracefulShutdown` (petición HTTP real sobre `127.0.0.1:0` y apagado ordenado).
  - Commit: `161afbf feat: hardening HTTP, PORT por env var, middlewares y graceful shutdown en main.go`.
- **Consulta puntual y actualización parcial de control clínico (`GET/PATCH /pacientes/:id/controles/:controlId`)**:
  - Puerto y adaptadores (`pkg/firestore/paciente_repo.go`, `mock.go`, `repo.go`): agregado `GetControlByID(ctx, pacienteID, controlID)` con paridad entre `MockPacienteRepository` (`nil`-safe, copia defensiva, exclusión de soft-deleted, `ErrControlNoEncontrado`) y `PacienteFirestoreRepo` (reutilizando `obtenerControlExistente` y normalizando `ID`/`PacienteID`).
  - Handlers (`handlers/paciente.go`):
    - `GetControlPacienteByID` (`GET /pacientes/:id/controles/:controlId`): devuelve `200 OK` con el control, `404` si paciente o control no existen (o pertenecen a otro paciente) y `500` ante fallos internos.
    - `PatchControlPaciente` (`PATCH /pacientes/:id/controles/:controlId`): actualiza únicamente los campos enviados (`fecha` en `YYYY-MM-DD`/`RFC3339`, `peso`, `porcentaje_grasa`, `notas`) con validación estricta de tipos JSON y rangos clínicos (`400`), recalcula `IMC` según `paciente.Altura` vigente y persiste mediante `UpdateControl`.
  - Rutas registradas en `SetupRoutes` ubicando `GET /pacientes/:id/controles/:controlId` explícitamente después de `/resumen` y `/export` para preservar la precedencia estática en Echo v4.
  - Tests nuevos (`handlers/paciente_test.go`, `pkg/firestore/mock_test.go`, `pkg/firestore/repo_test.go`): `TestGetControlPacienteByIDOK`, `TestGetControlPacienteByIDErrores`, `TestPatchControlPacienteOK`, `TestPatchControlPacienteErrores`, `TestMockGetControlByID` y `TestFirestoreRepoGetControlByID` (113 tests top-level en verde).
  - Commit: `db52369 feat: consulta puntual y actualización parcial de controles clínicos (GET/PATCH)`.
- **Verificación de CI en GitHub Actions**:
  - Comprobada vía `gh run view 44898033376` la primera ejecución exitosa de `.github/workflows/ci.yml` sobre `main` (`46e9d59`, `Go Vet, Test (race) & Build` en `1m26s`).
  - Detectada anotación informativa de deprecación de Node.js 20 en `actions/checkout@v4` y `actions/setup-go@v5` (ejecutados en Node.js 24 por el runner).

**Dificultades encontradas y por qué se construyó así:**

1. **Testabilidad determinista del `graceful shutdown` sin puertos fijos ni fugas de goroutines.** `e.Start(":8080")` acopla el enlace TCP y el bucle de servicio dentro de `main()`, impidiendo probar el apagado ordenado en paralelo o sin colisiones de puerto. Separar `newEcho`, `newServer`, `serve(ctx, srv, ln, timeout)` y `runServer` permitió inyectar un `net.Listener` efímero (`127.0.0.1:0`), enviar una petición real a `/health`, cancelar el contexto y drenar `errCh` (buffer `1`) tras `srv.Shutdown`, verificado en `50/50` ejecuciones consecutivas sin flakiness.
2. **Orden de middlewares en `newEcho`.** Registrar `RequestID` → `Logger` → `Recover` asegura que el `X-Request-ID` exista antes de loguear cada petición y que `Recover` capture panics dentro de los handlers devolviendo `500` en lugar de derribar el proceso.
3. **Precedencia de rutas en Echo v4 al introducir `GET /pacientes/:id/controles/:controlId`.** Esta ruta comparte profundidad y método `GET` con `/pacientes/:id/controles/resumen` y `/pacientes/:id/controles/export`. Registrándola después de ambas rutas estáticas en `SetupRoutes` y validando explícitamente las tres rutas dentro de `TestGetControlPacienteByIDOK` se garantiza contractualmente ausencia de colisiones.
4. **Validación estricta de tipos JSON en `PatchControlPaciente`.** Al deserializar en `map[string]any` para distinguir campos omitidos de valores cero, un valor con tipo incorrecto (p. ej. `"notas": 123` o `"peso": "alto"`) causaría panic con un type assertion directo sin coma-ok (como ocurre aún en `PatchPaciente` para `notas`, identificado como mejora pendiente). En `PatchControlPaciente` todos los campos se validan con `ok` devolviendo `400 Bad Request` descriptivo, y `fecha` reutiliza `parseFechaQuery` para aceptar tanto `YYYY-MM-DD` como `RFC3339`.
5. **Validación real de `-race` en GitHub Actions y detección de aviso Node.js 20.** Aunque el host Windows carece de `gcc`/CGO para correr `-race` en local, la corrida `44898033376` confirmó en `ubuntu-latest` que `go test -race -count=1 ./...` pasa sin carreras y sacó a la luz el aviso de deprecación de Node.js 20 en `actions/checkout@v4` y `actions/setup-go@v5`.

---
## Versión 1.0.17 - 2026-11-21

### Edición y eliminación de controles clínicos puntuales (`PUT/DELETE /pacientes/:id/controles/:controlId`) y pipeline CI en GitHub Actions

- **Puerto y adaptadores (`pkg/firestore/`)**:
  - Definido el error centinela `ErrControlNoEncontrado` y los métodos `UpdateControl` y `DeleteControl` en `PacienteRepository` (`pkg/firestore/paciente_repo.go`).
  - `MockPacienteRepository` (`pkg/firestore/mock.go`): valida paciente activo (`GetByID`), aplica protección `nil`-safe sobre `m.controles`, aísla controles por `pacienteID`, conserva la `Fecha` original cuando se envía fecha cero y actualiza o elimina el elemento correspondiente.
  - `PacienteFirestoreRepo` (`pkg/firestore/repo.go`): extrae el helper `obtenerControlExistente` para verificar paciente activo y existencia del documento en `pacientes/{id}/controles/{controlId}`, mapeando `codes.NotFound` / `!doc.Exists()` a `ErrControlNoEncontrado` y ejecutando `Set` o `Delete` sobre la referencia.
  - `go mod tidy` promovió `google.golang.org/grpc` y `google.golang.org/protobuf` a `require` directo en `go.mod` por el uso directo de `grpc/codes` y `grpc/status` en `repo.go`.
- **Handlers HTTP y rutas (`handlers/paciente.go`)**:
  - Extraídos los helpers compartidos `validarControlClinico` y `calcularIMCControl` (reutilizados por `CreateControlPaciente` y `UpdateControlPaciente`).
  - `UpdateControlPaciente` (`PUT /pacientes/:id/controles/:controlId`): valida paciente activo (`404`), valida peso (`>0 && <=500`) y porcentaje de grasa (`0-100`) (`400`), recalcula el `IMC` usando `paciente.Altura` actual y devuelve `200 OK` (`404` si el control no existe o pertenece a otro paciente).
  - `DeleteControlPaciente` (`DELETE /pacientes/:id/controles/:controlId`): valida paciente activo y existencia del control, devolviendo `204 No Content` (`404` si no existe o está aislado en otro paciente).
  - Rutas registradas en `SetupRoutes` con ambas sintaxis (`:id`/`:controlId` y `{id}`/`{controlId}`) antes de `/pacientes/:id`.
- **Pipeline CI (`.github/workflows/ci.yml`)**:
  - Workflow en `ubuntu-latest` activado en `push` y `pull_request` sobre `main`, con permisos mínimos (`contents: read`), `actions/checkout@v4`, `actions/setup-go@v5` (`go-version-file: go.mod`) y ejecución secuencial de `go vet ./...`, `go test -race -count=1 ./...` y `go build ./...`.
- **Tests (`handlers/paciente_test.go`, `pkg/firestore/mock_test.go`, `pkg/firestore/repo_test.go`)**:
  - Agregados `TestUpdateControlPacienteOK`, `TestUpdateControlPacienteErrores`, `TestDeleteControlPacienteOK`, `TestDeleteControlPacienteErrores`, `TestMockUpdateYDeleteControl` y `TestFirestoreRepoUpdateYDeleteControl` (103 tests top-level en verde).
- **Opción 3 (lente `review-readability` en `models.json`)**:
  - Verificado en solo lectura que `C:\Users\alvar\.pi\agent\models.json` únicamente declara `providers` (`strata` y `ollama`) sin configuración de lentes RDD; además, editar configuración externa de agente/host queda fuera del alcance del repositorio (acción manual en el host RDD si se decide reactivar).
- Commits: `027f1af` (`feat: edición y eliminación de controles clínicos puntuales`) y `b3df81f` (`ci: agregar workflow de GitHub Actions con vet, test -race y build`).

**Dificultades encontradas y por qué se construyó así:**

1. **Diferenciación precisa entre `404 Not Found` y errores internos `500` en Firestore.** Si el handler interpretara cualquier error devuelto por `UpdateControl` o `DeleteControl` como `404`, un fallo transitorio de red o permisos en Firestore se ocultaría como "control no encontrado". Al introducir `ErrControlNoEncontrado` en `pkg/firestore` y traducir únicamente `codes.NotFound` o `!doc.Exists()` en `obtenerControlExistente`, los handlers usan `errors.Is(err, firestore.ErrControlNoEncontrado)` para responder `404` y propagan cualquier otro error como `500`.
2. **Preservación de la fecha histórica cuando `PUT` omite el campo `fecha`.** En `ControlClinico`, `Fecha` es `time.Time` (valor cero `0001-01-01` si se omite en JSON). A diferencia de `POST` (donde omitir `fecha` asigna `time.Now().UTC()`), sobrescribir un control existente con la fecha cero o con la fecha actual al corregir un peso histórico distorsionaría el orden cronológico y el resumen evolutivo. Por ello, tanto `MockPacienteRepository` como `PacienteFirestoreRepo` leen el control previo y conservan `existing.Fecha` cuando `control.Fecha.IsZero()`, mientras recalculan siempre `IMC` con la altura vigente del paciente.
3. **Eliminación de duplicación entre alta y edición de controles (`validarControlClinico`, `calcularIMCControl` y `obtenerControlExistente`).** Para evitar divergencias en las cotas clínicas o el redondeo del IMC entre `POST` y `PUT`, y entre `UpdateControl` y `DeleteControl` en `repo.go`, se centralizó la lógica compartida en helpers privados manteniendo el contrato público inalterado.
4. **Limitación local de `go test -race` en Windows y descarga al runner Linux de CI.** Al ejecutar `go test -race -count=1 ./...` localmente en Windows falló por falta de `gcc`/CGO en `PATH` (`go: -race requires cgo`). El workflow `.github/workflows/ci.yml` sobre `ubuntu-latest` resuelve esa cobertura con detector de carreras en cada `push` y `pull_request`, complementado por un bloqueo `sync.Mutex` verificado en `fakeFirestoreServer` y ausencia de estado compartido mutable en `handlers` y tests del mock.
5. **Alcance de configuración externa RDD (`review-readability`).** La clave no existe dentro del proyecto ni en el esquema de proveedores de `models.json` del agente Pi local (`providers.strata`/`ollama`), por lo que queda documentada como ajuste manual exclusivo del entorno RDD del host para respetar los límites del repositorio.

---
## Versión 1.0.16 - 2026-11-20

### Exportación CSV de controles clínicos (`GET /pacientes/:id/controles/export`) e inclusión de `altura` en `ExportPacientesCSV`

- **Helpers compartidos en `handlers/paciente.go`**:
  - `obtenerControlesFiltrados(c echo.Context, repo firestore.PacienteRepository)`: unifica la validación de ID (`400`), existencia de paciente activo (`404`), parseo y orden de fechas `?desde=&hasta=` (`400`) y filtrado cronológico inclusivo para `GetControlesPaciente` y `ExportControlesPacienteCSV`.
  - `escribirCSV(c echo.Context, filename string, header []string, rows [][]string)`: centraliza la serialización RFC 4180 vía `encoding/csv`, las cabeceras `Content-Type: text/csv; charset=utf-8` y `Content-Disposition: attachment; filename="<filename>"`.
- **Actualización de `ExportPacientesCSV` (`GET /pacientes/export`)**:
  - Agrega `altura` como 7.ª columna (`id,nombre,edad,peso,porcentaje_grasa,notas,altura`), emitiendo cadena vacía `""` cuando `Altura <= 0` y formato decimal exacto (`strconv.FormatFloat(..., 'f', -1, 64)`) cuando existe.
- **Nuevo endpoint `GET /pacientes/:id/controles/export` (`ExportControlesPacienteCSV`)**:
  - Exporta en CSV RFC 4180 (`attachment; filename="controles.csv"`) el historial cronológico completo o filtrado por fechas (`?desde=&hasta=`) con columnas `id,paciente_id,fecha,peso,porcentaje_grasa,imc,notas` (`fecha` en `RFC3339` UTC e `imc` vacío cuando `<= 0`).
  - Rutas `/pacientes/:id/controles/export` y `/pacientes/{id}/controles/export` registradas en `SetupRoutes` antes de `/pacientes/:id`.
- **Tests en `handlers/paciente_test.go` (97 tests top-level en verde)**:
  - Actualizados `TestExportPacientesCSVOK` y `TestExportPacientesCSVVacio` verificando la columna `altura` (vacía en seed y `"1.75"` cuando existe).
  - Agregados `TestExportControlesPacienteCSVOK`, `TestExportControlesPacienteCSVFiltroFechasYVacios` y `TestExportControlesPacienteCSVNoEncontradoYSoftDeleted`.
- `go test -count=1 ./...`, `go vet ./...`, `go build ./...` todos verdes.
- Commit: `986116a feat: exportación CSV de controles clínicos y altura en ExportPacientesCSV`.

**Dificultades encontradas y por qué se construyó así:**

1. **Preservación de contratos posicionales al añadir `altura` en `ExportPacientesCSV`.** Insertar `altura` entre `peso` y `porcentaje_grasa` habría desplazado los índices de columna `4` y `5` en hojas de cálculo o consumidores CSV existentes. Al ubicar `altura` como 7.ª columna al final y emitir `""` cuando `Altura <= 0` (en coherencia con `omitempty` en JSON y evitando falsos `"0"` metros), se mantiene retrocompatibilidad posicional y claridad semántica.
2. **Eliminación de duplicación entre listados y exportadores (`obtenerControlesFiltrados` y `escribirCSV`).** En lugar de copiar la lógica de `parseFechaQuery`, validación `400`/`404`, filtrado temporal y generación de `bytes.Buffer` + `csv.NewWriter`, se extrajeron ambos helpers compartidos. `obtenerControlesFiltrados` devuelve `(controles, status, errMsg)` sin escribir en el contexto HTTP para mantener explícito el flujo de errores en cada handler.
3. **Seguridad en `Content-Disposition` y formato temporal determinista en CSV.** Interpolar el parámetro de ruta `id` dentro de `attachment; filename="controles-<id>.csv"` expondría la cabecera HTTP a manipulación si el cliente envía caracteres especiales; por ello se fijó el nombre estático `controles.csv` igual que `pacientes.csv`, formateando las fechas en `UTC().Format(time.RFC3339)` e `IMC == 0` como cadena vacía `""`.
4. **Precedencia de ruta en Echo v4.** `/pacientes/:id/controles/export` se registró junto a `/pacientes/:id/controles/resumen` antes de `/pacientes/:id` para asegurar el match de sub-recursos profundos.

---
## Versión 1.0.15 - 2026-11-19

### Filtrado por rango de fechas y paginación en `GET /pacientes/:id/controles`

- **Helper `parseFechaQuery(valor string, finDeDia bool)` en `handlers/paciente.go`**:
  - Acepta fechas en formato `RFC3339` (`2026-10-15T15:00:00Z`) o `YYYY-MM-DD` (`2026-10-15`).
  - Cuando `finDeDia` es `true` y la entrada viene en formato `YYYY-MM-DD`, ajusta el límite al cierre del día UTC (`23:59:59.999999999`).
  - Rechaza formatos inválidos con `400 Bad Request`.
- **Actualización de `GetControlesPaciente` (`GET /pacientes/:id/controles`)**:
  - Valida paciente activo (`404` si no existe o fue dado de baja lógica).
  - Lee `desde` y `hasta`; si `desde` es posterior a `hasta`, devuelve `400 Bad Request` (`"El parámetro 'desde' no puede ser posterior a 'hasta'"`).
  - Filtra el historial cronológico de forma inclusiva (`!Fecha.Before(*desde)` y `!Fecha.After(*hasta)`).
  - Aplica paginación sobre los resultados filtrados reutilizando `paginacionDesdeQuery(c)` (`limit = 10` y `offset = 0` por defecto), devolviendo `[]` vacío si `offset >= total`.
- **Tests agregados en `handlers/paciente_test.go` (3 nuevos tests / 11 subtests, 94 tests top-level en verde)**:
  - `TestGetControlesPacienteFiltroFechas`: table-driven cubriendo rangos `YYYY-MM-DD` inclusivos hasta fin de día, `RFC3339` exacto, límite superior único (`?hasta=`) y rango sin coincidencias.
  - `TestGetControlesPacienteFiltroFechasInvalidas`: table-driven rechazando `desde` inválido, `hasta` inválido y rango invertido (`desde > hasta`) con `400 Bad Request`.
  - `TestGetControlesPacientePaginacion`: verifica `limit` por defecto en `10`, `limit`/`offset` explícitos, `offset` superior al total y combinación de filtro de fechas + paginación.
- `go test -count=1 ./...`, `go vet ./...`, `go build ./...` todos verdes.
- Commit: `8f4cad4 feat: filtrado por fechas y paginación en GET /pacientes/:id/controles`.

**Dificultades encontradas y por qué se construyó así:**

1. **Rango inclusivo al usar `YYYY-MM-DD` sin hora (`finDeDia`).** `time.Parse("2006-01-02", "2026-11-01")` produce `2026-11-01 00:00:00 UTC`. Si `?hasta=2026-11-01` se comparara directamente con ese instante, cualquier control cargado durante ese día (por ejemplo a las `10:00:00Z`) quedaría excluido. Ajustar `hasta` en formato `YYYY-MM-DD` a `23:59:59.999999999 UTC` hace el extremo superior naturalmente inclusivo sin obligar al cliente a enviar marcas horarias completas.
2. **Precedencia entre existencia del paciente (`404`) y validación de query params (`400`).** `GetControlesPaciente` verifica primero `repo.GetByID` para responder `404` ante un paciente inexistente o eliminado lógicamente, y valida después la sintaxis y orden (`desde <= hasta`) de los parámetros temporales para devolver `400 Bad Request` explícito en lugar de retornar silenciosamente un array vacío ante rangos invertidos.
3. **Paginación posterior al filtrado y preservación del resumen evolutivo.** Al filtrar primero por fechas en `GetControlesPaciente` y aplicar después `paginacionDesdeQuery` (`limit=10`, `offset=0`), las páginas reflejan exactamente el subconjunto buscado; por su parte, `GetResumenControlesPaciente` sigue consultando directamente `repo.GetControles(ctx, id)` sin pasar por la capa HTTP de paginación, garantizando que el resumen evolutivo considere siempre el historial completo aunque el paciente supere 10 controles.

---
## Versión 1.0.14 - 2026-11-18

### Cobertura real de `PacienteFirestoreRepo` sobre gRPC / Emulator y endpoint evolutivo `GET /pacientes/:id/controles/resumen`

- **Tests reales sobre `*firestore.Client` (`pkg/firestore/repo.go` y `pkg/firestore/repo_test.go`)**:
  - `NewPacienteFirestoreRepoWithClient(client *firestore.Client)`: permite inyectar clientes Firestore en tests sin alterar la firma pública existente.
  - `NewPacienteFirestoreRepo(ctx, projectID)`: omite `option.WithEndpoint("firestore.googleapis.com:443")` cuando `FIRESTORE_EMULATOR_HOST` está definido, permitiendo conectar al emulador oficial automáticamente.
  - `fakeFirestoreServer` in-process sobre `cloud.google.com/go/firestore/apiv1/firestorepb` (`Commit`, `BatchGetDocuments`, `RunQuery`) en `pkg/firestore/repo_test.go`: ejercita las llamadas reales del SDK de Firestore sobre loopback `127.0.0.1:0` (o contra `FIRESTORE_EMULATOR_HOST` si está configurado).
  - Tests agregados: `TestFirestoreRepoCRUDYGetAllPaginado`, `TestFirestoreRepoSoftDeleteYRestore`, `TestFirestoreRepoControlesClinicos` (elevando `pkg/firestore` al **81.9%** de cobertura de sentencias y cerrando `Risk: item 6`).
  - Commit: `0660bf5 test: cubrir PacienteFirestoreRepo sobre gRPC in-process y Emulator`.
- **Endpoint evolutivo `GET /pacientes/:id/controles/resumen` (`handlers/paciente.go`)**:
  - Struct `ResumenControles`: expone `paciente_id`, `total_controles`, `fecha_inicio` (`omitempty`), `fecha_fin` (`omitempty`), `peso_inicial`, `peso_actual`, `delta_peso`, `porcentaje_grasa_inicial`, `porcentaje_grasa_actual`, `delta_porcentaje_grasa`, `imc_inicial`, `imc_actual` y `delta_imc`.
  - Handler `GetResumenControlesPaciente` y helper `redondear2`: valida paciente activo (`404` si no existe o fue dado de baja lógica), recupera el historial ordenado cronológicamente vía `repo.GetControles` y calcula la evolución entre el primer y último control registrado.
  - Rutas `/pacientes/:id/controles/resumen` y `/pacientes/{id}/controles/resumen` registradas en `SetupRoutes` antes de `/pacientes/:id`.
  - Tests agregados en `handlers/paciente_test.go` (3 nuevos, **91** ejecuciones `--- PASS:` totales incluyendo subtests): `TestGetResumenControlesPacienteOK`, `TestGetResumenControlesPacienteSinControlesYUnControl`, `TestGetResumenControlesPacienteNoEncontradoYSoftDeleted`.
  - Commit: `e220d1c feat: endpoint GET /pacientes/:id/controles/resumen evolutivo`.
- `go test -count=1 ./...`, `go vet ./...`, `go build ./...` todos verdes.

**Dificultades encontradas y por qué se construyó así:**

1. **Verificación de `PacienteFirestoreRepo` sin depender de Firebase CLI ni Docker en ejecución.** Aunque el código ya utilizaba el SDK oficial `cloud.google.com/go/firestore`, la suite local solo cubría `MockPacienteRepository`. Como el proyecto ya incorpora `google.golang.org/grpc` y `cloud.google.com/go/firestore/apiv1/firestorepb` en `go.mod`, se implementó `fakeFirestoreServer` en memoria con soporte de precondiciones `Exists`, respuestas `Found`/`Missing` y consultas por prefijo de colección/subcolección en `RunQuery`, evitando dependencias externas y logrando 81.9% de cobertura en `pkg/firestore` en milisegundos.
2. **Endpoint hardcodeado en `NewPacienteFirestoreRepo` frente a `FIRESTORE_EMULATOR_HOST`.** Forzar siempre `option.WithEndpoint("firestore.googleapis.com:443")` impide que el SDK oficial redirija tráfico cuando existe `FIRESTORE_EMULATOR_HOST`. Condicionar esa opción a `os.Getenv("FIRESTORE_EMULATOR_HOST") == ""` mantiene el comportamiento de producción y habilita pruebas reales contra el emulador sin cambios adicionales.
3. **Deriva de punto flotante y cálculo condicional de `delta_imc` en `ResumenControles`.** Al restar métricas decimales (`ultimo - primero`) pueden aparecer artefactos IEEE 754, por lo que todos los deltas se redondean con `redondear2` (`math.Round(val*100)/100`). Además, como `Altura` es opcional, un paciente puede tener controles con `IMC == 0`; en ese caso `delta_imc` solo se calcula cuando tanto `primero.IMC > 0` como `ultimo.IMC > 0`, evitando deltas artificiales negativos o positivos cuando falta la altura.
4. **Representación limpia de `0` controles sin serializar fechas cero.** Si el paciente activo aún no tiene controles clínicos, devolver `200 OK` con `total_controles: 0` mantiene consistencia con `GET /pacientes/:id/controles` (`[]`) y `GET /pacientes/reporte` (`total: 0`). Usar `*time.Time` con `omitempty` para `fecha_inicio` y `fecha_fin` evita emitir `"0001-01-01T00:00:00Z"` cuando no hay controles.

---
## Versión 1.0.13 - 2026-11-17

### Soft delete en pacientes e historial de controles clínicos (`GET /pacientes/inactivos`, `POST /pacientes/:id/restore`, `GET/POST /pacientes/:id/controles`)

- **Extensión del modelo de dominio (`models/paciente.go`)**:
  - `Paciente`: incorporación del campo `DeletedAt *time.Time` (`json:"deleted_at,omitempty"`). Cuando es `nil`, el paciente está activo y el campo se omite del JSON.
  - `ControlClinico`: nuevo struct con `ID`, `PacienteID`, `Fecha`, `Peso`, `PorcentajeGrasa`, `IMC` (`omitempty`) y `Notas` (`omitempty`).
- **Extensión del puerto `PacienteRepository` y adaptadores (`pkg/firestore/`)**:
  - `Delete(ctx, id)`: pasa a realizar baja lógica marcando `DeletedAt` con `time.Now().UTC()` (tanto en `MockPacienteRepository` como en `PacienteFirestoreRepo`); devuelve error si el registro no existe o ya está dado de baja.
  - `GetAll(ctx, limit, offset)`, `GetByID(ctx, id)` y `Update(ctx, id, p)`: excluyen automáticamente los registros con `DeletedAt != nil`. En `GetAll`, la paginación (`limit`/`offset`) se aplica sobre los pacientes activos ya filtrados.
  - `Restore(ctx, id)`: limpia `DeletedAt = nil` en pacientes dados de baja; falla si no existe o ya está activo.
  - `GetInactivos(ctx)`: devuelve exclusivamente los pacientes con `DeletedAt != nil`.
  - `AddControl(ctx, pacienteID, control)` y `GetControles(ctx, pacienteID)`: persisten y recuperan controles clínicos ordenados cronológicamente (`Fecha` ascendente) validando previamente que el paciente exista y siga activo (`GetByID`). En Firestore se almacenan en la subcolección `pacientes/{id}/controles`.
- **Nuevos handlers y protección de escritura (`handlers/paciente.go`)**:
  - `GetPacientesInactivos` (`GET /pacientes/inactivos`): lista pacientes dados de baja lógica.
  - `RestorePaciente` (`POST /pacientes/:id/restore` y `{id}/restore`): reactiva un paciente inactivo (`200 OK` o `404 Not Found`).
  - `CreateControlPaciente` (`POST /pacientes/:id/controles` y `{id}/controles`): valida paciente activo (`404`), `peso` (`> 0 y <= 500`) y `porcentaje_grasa` (`0..100`) (`400`), asigna `Fecha` UTC si se omite, calcula automáticamente `IMC` redondeado a 2 decimales si el paciente tiene `Altura > 0` y responde `201 Created`.
  - `GetControlesPaciente` (`GET /pacientes/:id/controles` y `{id}/controles`): devuelve el historial cronológico (`200 OK` o `404`).
  - `CreatePaciente`, `CreatePacientesBatch` y `UpdatePaciente` fuerzan `DeletedAt = nil` para impedir que clientes manipulen la marca de baja en peticiones de alta o actualización.
- **Tests agregados (85 ejecuciones `--- PASS:` incluyendo subtests)**:
  - `models/paciente_test.go`: `TestPacienteJSONSerializacionConDeletedAt`, `TestControlClinicoJSONSerializacion`.
  - `pkg/firestore/mock_test.go`: `TestMockSoftDeleteYRestore`, `TestMockGetAllPaginacionExcluyeSoftDeleted`, `TestMockAddYGetControles`.
  - `handlers/paciente_test.go`: `TestSoftDeleteExcluyeDeListadosYEndpoints`, `TestGetPacientesInactivosOK`, `TestRestorePacienteOK`, `TestRestorePacienteErrores`, `TestCreateControlPacienteOK`, `TestCreateControlPacienteValidaciones` (table-driven), `TestGetControlesPacienteOK`, `TestControlesPacienteNoEncontradoYSoftDeleted`.
- `go test -count=1 ./...`, `go vet ./...`, `go build ./...` todos verdes + verificación independiente read-only (`PASS`).
- Commit: `be4cc9a feat: soft delete en pacientes e historial de controles clínicos`.

**Dificultades encontradas y por qué se construyó así:**

1. **Elección de `DeletedAt *time.Time` con `omitempty` frente a `Activo bool`.** En Go, el valor cero de un booleano es `false`. Introducir `Activo bool` habría provocado que cualquier literal `models.Paciente{Nombre: "X", Edad: 30, Peso: 70}` en los 62 tests existentes o un JSON sin ese campo se interpretara como inactivo. Usar un puntero `*time.Time` (`nil` = activo) mantiene 100% de compatibilidad con la suite previa, conserva la fecha exacta de baja para auditoría y se omite del JSON en pacientes activos.
2. **Sesgo de paginación en `GetAll` tras soft delete.** Originalmente `MockPacienteRepository.GetAll` aplicaba `offset` y `limit` directamente sobre `m.store`, y `PacienteFirestoreRepo.GetAll` pasaba `.Limit`/`.Offset` antes de filtrar. Si el primer registro está dado de baja lógica, `limit=2` habría devuelto solo 1 elemento y en Firestore los documentos históricos sin campo `deleted_at` no coinciden con filtros `== null`. Por eso ambas implementaciones filtran primero los registros con `DeletedAt == nil` y aplican después `offset` y `limit` sobre el conjunto activo.
3. **Seguridad frente a `nil` map en instanciaciones literales `&MockPacienteRepository{}`.** Varios tests previos del repositorio crean el mock mediante `&firestore.MockPacienteRepository{}` sin invocar `NewMockPacienteRepository()`. Al añadir el mapa `controles map[string][]models.ControlClinico`, se inicializó de forma perezosa (`if m.controles == nil`) en `AddControl` y se validó su nulidad en `GetControles` para evitar `panic: assignment to entry in nil map`.
4. **Coherencia de serialización en `PacienteFirestoreRepo` sin etiquetas `firestore:`.** `models.Paciente` solo declara etiquetas `json:"..."`, por lo que el SDK de Firestore persiste las propiedades con el nombre del campo Go (`DeletedAt`). Si `Delete` hubiera usado `docRef.Update` con path `"deleted_at"`, habría escrito un campo distinto al que lee `doc.DataTo(&p)`. Al usar `docRef.Set(ctx, p)` tras mutar `p.DeletedAt` (tanto en `Delete` como en `Restore`), la escritura y la lectura mantienen total simetría.
5. **Precedencia de `/pacientes/inactivos` y blindaje de `DeletedAt` en endpoints de escritura.** `/pacientes/inactivos` se registró en `SetupRoutes` antes de `/pacientes/:id` para evitar que Echo interprete `"inactivos"` como un ID; además, los endpoints de creación y actualización (`POST`, `POST /batch`, `PUT`) limpian `DeletedAt = nil` para evitar bajas simuladas vía payload.
6. **Cálculo condicional de `IMC` y ordenamiento cronológico en controles clínicos.** Dado que `Altura` es opcional en `models.Paciente`, `CreateControlPaciente` calcula y guarda `IMC` redondeado a 2 decimales solo cuando `paciente.Altura > 0` (evitando divisiones por cero) y asigna `time.Now().UTC()` si `Fecha` viene vacía; `GetControles` ordena con `sort.SliceStable` por `Fecha.Before` para asegurar una línea temporal consistente independientemente del orden de inserción.
7. **Brecha de cobertura en `PacienteFirestoreRepo` sin emulador en vivo.** Los tests unitarios ejercitan `MockPacienteRepository`, por lo que la lógica agregada en `pkg/firestore/repo.go` queda verificada por compilación, `go vet` e inspección independiente de paridad contra el mock, quedando la batería de integración contra Firestore Emulator como TODO explícito.

---
## Versión 1.0.12 - 2026-11-16

### Endpoint `GET /pacientes/reporte` (Analítica demográfica y nutricional de la población)

- **Estructuras de respuesta en `handlers/paciente.go`**:
  - `GruposEtarios`: desglosa la población en `menores` (`Edad < 18`), `adultos` (`18 <= Edad <= 65`) y `mayores` (`Edad > 65`).
  - `ClasificacionIMCPoblacion`: contabiliza pacientes por categoría OMS (`bajo_peso`, `normal`, `sobrepeso`, `obesidad`) más `sin_datos` para pacientes sin altura o peso válido registrado.
  - `PromediosReporte`: promedios poblacionales de `edad`, `peso`, `porcentaje_grasa` e `imc` redondeados a 2 decimales (`math.Round(val*100)/100`).
  - `RespuestaReporte`: estructura contenedora con `total`, `grupos_etarios`, `clasificacion_imc` y `promedios`.
- **Handler `ReportePacientes` (`GET /pacientes/reporte`)**:
  - Consulta el universo completo de pacientes mediante `repo.GetAll(c.Request().Context(), 0, 0)`.
  - Reutiliza `clasificarIMC` sobre el IMC redondeado a 2 decimales cuando el paciente posee `Altura > 0` y `Peso > 0`; en caso contrario contabiliza el registro en `sin_datos`.
  - Si el repositorio está vacío (`total == 0`), devuelve `200 OK` con todos los contadores y promedios en `0`.
- **Registro en `SetupRoutes`**: ruta estática `GET /pacientes/reporte` ubicada junto a `/pacientes/stats` y antes de `/pacientes/:id` y `/pacientes/{id}`.
- **Tests agregados en `handlers/paciente_test.go` (3 nuevos, 72 ejecuciones `--- PASS:` incluyendo subtests)**:
  - `TestReportePacientesOK`: verifica el conteo exacto en las 3 franjas etarias, las 4 categorías OMS más `sin_datos` (pacientes del seed sin altura) y los 4 promedios redondeados a 2 decimales.
  - `TestReportePacientesVacio`: verifica respuesta `200 OK` con todos los contadores y promedios en `0` sobre repositorio vacío.
  - `TestReportePacientesRutaRegistrada`: verifica a través de `e.ServeHTTP` que `/pacientes/reporte` se resuelva en el router de Echo antes de `/pacientes/:id`.
- `go test -count=1 ./...`, `go vet ./...`, `go build ./...` todos verdes.
- Commit: `79d9c53 feat: endpoint GET /pacientes/reporte para analítica poblacional`.

**Dificultades encontradas y por qué se construyó así:**

1. **Contabilidad poblacional exacta ante pacientes sin altura (`sin_datos` en `ClasificacionIMCPoblacion`).** Dado que `Altura` se definió como opcional (`json:"altura,omitempty"`) en v1.0.11 para preservar retrocompatibilidad con registros históricos y con el seed inicial de `MockPacienteRepository` (`Juan Perez` y `Maria Gomez` tienen `Altura: 0`), una parte del padrón puede carecer de altura. Si simplemente se omitieran del bloque `clasificacion_imc`, la suma de categorías nutricionales no coincidiría con `total`. Por eso se incorporó el campo explícito `sin_datos`, preservando el invariante `bajo_peso + normal + sobrepeso + obesidad + sin_datos == total`, y el promedio `promedios.imc` se divide únicamente entre los pacientes con IMC calculable (`conIMC > 0`) para no distorsionar la media clínica con ceros ni dividir por cero cuando ningún paciente tiene altura.
2. **Consulta del padrón completo (`repo.GetAll(ctx, 0, 0)`) frente a la paginación por defecto.** A diferencia de `GetPacientes` o `SearchPacientes` que aplican `paginacionDesdeQuery` (`limit = 10` por defecto), un reporte analítico demográfico y nutricional debe reflejar el estado global de la población registrada y no solo la primera página. Por ello `ReportePacientes` invoca `repo.GetAll(c.Request().Context(), 0, 0)`, manteniendo el mismo criterio que `CountPacientes` y `ExportPacientesCSV`.
3. **Consistencia clínica entre el cálculo individual (`GetPacienteIMC`) y el reporte agregado (`ReportePacientes`).** En lugar de reescribir umbrales de comparación en el reporte, se calcula el IMC individual redondeado primero a 2 decimales (`math.Round((p.Peso/(p.Altura*p.Altura))*100) / 100`) y luego se invoca la función existente `clasificarIMC(imc)`. Esto evita discrepancias de punto flotante en valores límite (como `24.999`) y garantiza que la categoría de un paciente en `GET /pacientes/imc/:id` sea idéntica a la que suma en `GET /pacientes/reporte`.
4. **Precedencia de la ruta estática `/pacientes/reporte` en Echo v4.** Siguiendo la convención de rutas estáticas del proyecto, `/pacientes/reporte` se registró en `SetupRoutes` antes de `/pacientes/:id` y `/pacientes/{id}`, y se validó con `TestReportePacientesRutaRegistrada` mediante `e.ServeHTTP` para asegurar que el router no interprete `"reporte"` como un identificador de paciente.

---
## Versión 1.0.11 - 2026-11-14

### Extensión de Dominio con `Altura` y endpoints de cálculo de IMC (`GET /pacientes/imc/{id}` y `GET /pacientes/{id}/imc`)

- **Modelo de dominio (`models.Paciente`)**: incorporación del campo `Altura float64` con etiqueta `json:"altura,omitempty"`. Al utilizar `omitempty`, valores `0` (pacientes creados sin altura previa) no rompen clientes ni esquemas serializados existentes.
- **Validaciones de negocio clínicas en `validarPaciente`**:
  - Constantes `alturaMinima = 0.30` y `alturaMaxima = 2.50` metros.
  - Validación opcional: solo se evalúa si `p.Altura != 0`. Si se envía, debe estar estrictamente en el rango `[0.30, 2.50]`, rechazando valores menores, mayores o negativos con `400 Bad Request` (`"El campo 'altura' debe estar entre 0.30 y 2.50 metros"`).
  - Aplicado automáticamente en `CreatePaciente` (`POST /pacientes`), `CreatePacientesBatch` (`POST /pacientes/batch`), `UpdatePaciente` (`PUT /pacientes/:id`) y `PatchPaciente` (`PATCH /pacientes/:id`).
- **Adaptador Mock**: actualizado `MockPacienteRepository.Update` para persistir `m.store[i].Altura = p.Altura`.
- **Ordenamiento dinámico**: `ordenarPacientes` ampliado con soporte de ordenamiento por `"altura"` en orden ascendente y descendente (`?sort=altura&order=asc|desc`).
- **Endpoints de cálculo clínico de IMC**:
  - `GET /pacientes/imc/{id}` y `GET /pacientes/{id}/imc`: invocan `GetPacienteIMC`.
  - Verifica existencia del paciente (404 Not Found si no existe).
  - Verifica que el paciente posea altura registrada (`altura <= 0` retorna `400 Bad Request` indicando que se requiere registrar la altura) y peso válido (`peso <= 0` retorna 400).
  - Calcula el Índice de Masa Corporal $IMC = \frac{peso}{altura^2}$, redondeado a 2 decimales (`math.Round(imc*100)/100`).
  - Categorización según estándares de la OMS: `< 18.5` "Bajo peso", `18.5 - 24.9` "Normal", `25.0 - 29.9` "Sobrepeso", `>= 30.0` "Obesidad".
  - Retorna `200 OK` con struct `RespuestaIMC` (`id`, `nombre`, `peso`, `altura`, `imc`, `clasificacion`).
- **SetupRoutes y corrección de sintaxis de router Echo**:
  - Mapeo dual para IMC: `/pacientes/imc/:id` y `/pacientes/:id/imc`.
  - Normalización de rutas parametrizadas registradas con sintaxis nativa de Echo (`:id` y `:edad`) además de `{id}` para asegurar el match en peticiones HTTP reales a través del router.
- **Tests agregados (10 nuevos, suite completa en 59 tests)**:
  - `TestPacienteJSONSerializacionConAltura`: verifica serialización de `altura` en `models`.
  - `TestGetPacienteIMCOK`: cálculo correcto de IMC 24.65 ("Normal") con peso 75.5 y altura 1.75.
  - `TestGetPacienteIMCRutasRegistradas`: router Echo respondiendo 200 tanto en `/pacientes/imc/:id` como en `/pacientes/:id/imc`.
  - `TestGetPacienteIMCNoEncontrado`: 404 Not Found con ID inexistente.
  - `TestGetPacienteIMCSinAltura`: 400 Bad Request cuando el paciente no tiene altura registrada (`altura == 0`).
  - `TestGetPacienteIMCClasificaciones`: table-driven cubriendo las 4 categorías OMS (Bajo peso, Normal, Sobrepeso, Obesidad).
  - `TestCreatePacienteAlturaValida`: creación exitosa de paciente con altura informada.
  - `TestCreatePacienteAlturaFueraDeRango`: table-driven rechazando alturas menores a 0.30 m, mayores a 2.50 m y negativas con 400.
  - `TestPatchPacienteAltura`: actualización parcial de altura vía PATCH.
  - `TestPatchPacienteAlturaInvalida`: rechazo 400 ante altura fuera de rango en PATCH.
- `go test ./...`, `go vet ./...`, `go build ./...` todos verdes (59 tests pasando).

**Dificultades encontradas y por qué se construyó así:**

1. **Compatibilidad hacia atrás del campo `Altura` en `models.Paciente`.** Se utilizó la etiqueta `json:"altura,omitempty"`. Al ser `Altura` de tipo `float64`, un valor `0` (paciente sin altura cargada) no se emite en el JSON si no se desea alterar contratos preexistentes, y los registros anteriores en Firestore o en tests unitarios no fallan ni requieren migración forzosa.
2. **Sintaxis de parámetros en el router de Echo v4 (`:id` vs `{id}`).** Históricamente las rutas parametrizadas se habían registrado con `{id}` en `SetupRoutes`, lo cual funcionaba en tests unitarios directos gracias al helper de inyección con reflexión `setParam`, pero al invocar `e.ServeHTTP` a través del router de Echo, la sintaxis `{id}` era tratada como un segmento estático literal. Se solucionó registrando las rutas parametrizadas con la sintaxis nativa de Echo (`:id` y `:edad`) garantizando el ruteo real en HTTP sin romper llamadas legacy.
3. **Ruta dual para el cálculo de IMC (`/pacientes/imc/:id` y `/pacientes/:id/imc`).** Se detectó que en especificaciones y notas previas se mencionaba `GET /pacientes/imc/{id}`, mientras que el estándar REST de diseño de APIs prescribe sub-recursos bajo la entidad (`GET /pacientes/{id}/imc`). En lugar de forzar una única convención, `SetupRoutes` mapea limpiamente ambas alternativas hacia el mismo handler `GetPacienteIMC`, maximizando la ergonomía para clientes frontend y consumidores externos.
4. **Validación condicional de altura en `validarPaciente`.** A diferencia del peso y la edad que son de carga obligatoria para crear un paciente, la altura puede no estar disponible en la primera consulta de triage. Por ello, la regla de validación clínica de altura solo se dispara cuando `Altura != 0`. Si se envía, debe estar estrictamente dentro del rango biológico plausible (0.30 a 2.50 metros).

---
## Versión 1.0.10 - 2026-11-13

### Endpoint `POST /pacientes/batch` (Alta masiva de pacientes con validación atómica)

- Handler `CreatePacientesBatch` (`POST /pacientes/batch`) en `handlers/paciente.go`: permite dar de alta una colección de pacientes en una sola llamada HTTP.
- Validación atómica en dos fases:
  1. Si el payload no es un JSON válido o la lista está vacía (`[]`), responde `400 Bad Request` con mensaje explicativo (`"El lote de pacientes no puede estar vacío"`).
  2. Normaliza (`normalizarNombre`) y valida (`validarPaciente`) cada elemento. Si cualquier paciente incumple las reglas clínicas (edad 1–120, peso 0–500, grasa 0–100, nombre requerido), la petición se rechaza de inmediato con `400 Bad Request` señalando el índice fallido (`"Paciente en posición %d inválido: %s"`).
  3. Solo cuando el 100% del lote es válido, se procede con la persistencia secuencial en el repositorio (`repo.Create`), garantizando que no queden pacientes persistidos parcialmente ante errores de validación.
- Devuelve `201 Created` con el slice completo de pacientes y sus identificadores únicos asignados.
- Ruta registrada en `SetupRoutes`.
- Tests agregados en `handlers/paciente_test.go`:
  - `TestCreatePacientesBatchOK`: verifica creación de lote de 2 pacientes válidos con espacios en nombres, confirmando normalización, asignación de IDs y persistencia en el repositorio (total 4 pacientes).
  - `TestCreatePacientesBatchVacio`: verifica que un slice `[]` retorne `400 Bad Request`.
  - `TestCreatePacientesBatchAtomicidadInvalido`: verifica que ante un paciente válido seguido de uno inválido, se retorne `400 Bad Request` y el repositorio conserve estrictamente los registros previos (cero escrituras parciales).
  - `TestCreatePacientesBatchJSONInvalido`: verifica respuesta `400 Bad Request` ante cuerpos no serializables en array de pacientes.
  - `TestCreatePacientesBatchRutaRegistrada`: test router-level que verifica que `POST /pacientes/batch` responda `201 Created` a través del router de Echo.
- `go test ./...`, `go vet ./...`, `go build ./...` todos verdes (49 tests totales).

**Dificultades encontradas y por qué se construyó así:**

1. **Garantía de atomicidad sin transacciones distribuidas en storage.** Sin imponer requisitos de soporte transaccional ACID en la interfaz abstracta `PacienteRepository` (que debe soportar tanto Firestore como el Mock en memoria o futuras bases SQL/NoSQL), la atomicidad frente a datos inválidos se resolvió separando la lógica en dos pasadas: (1) validación estricta de todo el lote en memoria, y (2) persistencia efectiva solo tras haber aprobado todos los registros. Si el segundo registro falla, no se llega a insertar el primero.
2. **Compatibilidad de tipos con la firma de `repo.Create`.** `PacienteRepository.Create` retorna `(*models.Paciente, error)`. Al acumular las respuestas para la respuesta HTTP, el slice `creados` se definió como `[]*models.Paciente` para admitir directamente las referencias devueltas por el adaptador sin conversiones redundantes y serializar limpiamente como array JSON en Echo.
3. **Manejo explícito de lotes vacíos.** Enviar un array vacío `[]` no debe generar operaciones vacías ni devolver `201 Created`. Se rechaza explícitamente con `400 Bad Request` y mensaje `"El lote de pacientes no puede estar vacío"`.

---
## Versión 1.0.9 - 2026-11-12

### Ordenamiento dinámico en `GET /pacientes` y `GET /pacientes/search`, y nuevo endpoint `GET /pacientes/export` (CSV)

- Función `ordenarPacientes` en `handlers/paciente.go`: ordena en memoria el slice de pacientes según `sort` (`nombre`, `edad`, `peso`, `porcentaje_grasa`) y `order` (`asc` por defecto, o `desc`). Case-insensitive en nombres y soporte de campos numéricos.
- Integrado ordenamiento en `GetPacientes` y en `SearchPacientes` (aplicado tras los filtros multicriterio).
- Handler `ExportPacientesCSV` (`GET /pacientes/export`) en `handlers/paciente.go`: consulta `repo.GetAll(ctx, 0, 0)`, construye un archivo CSV compatible con RFC 4180 (`encoding/csv`), configura cabecera `Content-Type: text/csv; charset=utf-8` y `Content-Disposition: attachment; filename="pacientes.csv"`.
- Ruta `GET /pacientes/export` registrada en `SetupRoutes` antes de `/pacientes/{id}` para evitar colisiones con el parámetro `{id}`.
- Tests agregados en `handlers/paciente_test.go`:
  - `TestGetPacientesOrdenamiento`: table-driven probando orden por `edad` (asc/desc), `nombre` (desc), `peso` (asc) y `porcentaje_grasa` (desc).
  - `TestSearchPacientesOrdenamiento`: table-driven probando orden en `search` sin filtros y con filtros combinados (`edad_min=20&sort=peso&order=desc`).
  - `TestExportPacientesCSVOK`: verifica cabeceras HTTP, fila de títulos `id,nombre,edad,peso,porcentaje_grasa,notas` y datos de los pacientes.
  - `TestExportPacientesCSVVacio`: verifica generación exclusiva de cabecera cuando no hay registros.
  - `TestExportPacientesCSVRutaRegistrada`: verifica a través del router de Echo que `/pacientes/export` responda 200 CSV y no caiga en el handler de `{id}` (404).
- `go test ./...`, `go vet ./...`, `go build ./...` todos verdes.

**Dificultades encontradas y por qué se construyó así:**

1. **Prioridad de rutas en Echo v4 frente a rutas parametrizadas.** Si `/pacientes/export` se registrara después de `/pacientes/{id}`, cualquier petición a `/pacientes/export` sería capturada por `GetPacienteByID` interpretando `"export"` como el ID del paciente (devolviendo `404 Not Found`). Se registró antes y se agregó un test de integración router-level (`TestExportPacientesCSVRutaRegistrada`) que ejecuta `e.ServeHTTP` para garantizar contractualmente que el router respeta la prioridad estática.
2. **Serialización y escape estricto de CSV mediante `encoding/csv`.** En lugar de concatenar strings con comas manualmente (lo cual rompería si un paciente tiene comas o saltos de línea en el campo libre `Notas`), se utilizó `csv.NewWriter` sobre un `bytes.Buffer`. Esto garantiza cumplimiento con RFC 4180 y escape transparente con comillas dobles.
3. **Formateo de números decimales sin pérdida de precisión ni ceros espurios.** Para los campos `peso` y `porcentaje_grasa` se optó por `strconv.FormatFloat(val, 'f', -1, 64)` en vez de `fmt.Sprintf("%.2f", val)`. El flag `-1` produce la representación exacta más corta necesaria (ej. `75.5` en vez de `75.50` y `62` en vez de `62.00`), reduciendo el tamaño del payload y evitando redondeos indeseados.
4. **Estabilidad en el ordenamiento con `sort.SliceStable`.** Se utilizó `sort.SliceStable` para que ante valores idénticos en el campo seleccionado se conserve el ordenamiento original relativo de la colección, evitando alternancias aleatorias en peticiones sucesivas.
5. **Diagnóstico de configuración en la relay de revisión externa (RDD).** Al activar el flujo de revisión nativa con las lentes `review-readability` y `review-reliability`, la infraestructura de relay del host reportó la ausencia de modelo asignado para la clave `review-readability` en `models.json`. Conforme al contrato RDD, el bloqueo de configuración externa no afecta la suite funcional local (`go test ./...` 44/44 pasando) ni interrumpe la entrega ordinaria del repositorio.

---
## Versión 1.0.8 - 2026-11-10

### Endpoints `GET /health`, `GET /pacientes/count` y filtros `porcentaje_grasa_min`/`max` en `GET /pacientes/search`

- Handler `HealthCheck` (`GET /health`) en `handlers/paciente.go`: devuelve `200 OK` con `{"status": "ok", "storage": "mock"|"firestore"}` según la implementación activa de `PacienteRepository`.
- Handler `CountPacientes` (`GET /pacientes/count`) en `handlers/paciente.go`: consulta `repo.GetAll(ctx, 0, 0)` y devuelve `200 OK` con `{"total": <n>}`.
- `SearchPacientes` (`GET /pacientes/search`): agrega filtros por rango inclusivo `porcentaje_grasa_min` y `porcentaje_grasa_max` (`strconv.ParseFloat`).
- `SetupRoutes`: agrupa todas las rutas con segmento estático (`/pacientes/search`, `/pacientes/stats`, `/pacientes/count`, `/pacientes/edad/{edad}`) antes de las rutas parametrizadas `/pacientes/{id}`.
- Tests agregados en `handlers/paciente_test.go`: `TestHealthCheckOK`, `TestCountPacientesOK`, `TestCountPacientesVacio` y `TestSearchPacientesFiltrarPorPorcentajeGrasaMinMax` (table-driven).
- `go test ./...`, `go vet ./...`, `go build ./...` todos verdes.
- Commit: `8d32dd6 feat: agregar GET /health, GET /pacientes/count y filtro porcentaje_grasa en search`.

**Dificultades encontradas y por qué se construyó así:**

1. **Elección de `GET /pacientes/count` frente a `GET /pacientes/imc/{id}`.** El modelo de dominio `models.Paciente` no posee campo `Altura`, por lo que calcular el IMC habría exigido alterar la estructura de dominio o recibir parámetros externos en una ruta de lectura. `GET /pacientes/count` respeta la frontera de dominio actual y reutiliza `repo.GetAll(ctx, 0, 0)` sin tocar `pkg/firestore/`.
2. **Detección del modo de almacenamiento en `HealthCheck` sin modificar la interfaz `PacienteRepository`.** Mediante una aserción de tipo `_, isMock := repo.(*firestore.MockPacienteRepository)` se informa si el servidor corre en modo `"mock"` o `"firestore"` manteniendo intacto el contrato del puerto `PacienteRepository`.
3. **Datos seed del mock en `TestSearchPacientesFiltrarPorPorcentajeGrasaMinMax`.** En `NewMockPacienteRepository()`, `Juan Perez` tiene `PorcentajeGrasa: 18.5` y `Maria Gomez` tiene `22.0`; al ejecutar la primera corrida de los tests table-driven se detectó que los nombres esperados estaban invertidos respecto del seed real, corrigiéndose inmediatamente contra `pkg/firestore/mock.go`.
4. **Reordenamiento preventivo de rutas estáticas en `SetupRoutes`.** Al incorporar `/pacientes/count` junto a `/pacientes/search`, `/pacientes/stats` y `/pacientes/edad/{edad}`, todas las rutas con segmento estático bajo `/pacientes/...` se ubicaron explícitamente antes de `/pacientes/{id}` (`GET`, `PUT`, `PATCH`, `DELETE`) para que el orden de declaración refleje la prioridad real de enrutamiento y evite colisiones con el parámetro `{id}`.

---
## Versión 1.0.7 - 2026-11-09

### Validación de `porcentaje_grasa` (0–100), `nombre` en PATCH y cobertura de tests en `DeletePaciente` y `GetPacientesPorEdad`

- `validarPaciente` en `handlers/paciente.go`: agrega validación de `porcentaje_grasa` entre `0` y `100` (`porcentajeGrasaMaximo = 100`).
- `PatchPaciente` en `handlers/paciente.go`: valida que `porcentaje_grasa` esté entre `0` y `100` cuando se envía, usa aserción de tipo segura `nombre.(string)` y rechaza con `400 Bad Request` un `nombre` vacío o compuesto solo por espacios tras `normalizarNombre`.
- Tests agregados en `handlers/paciente_test.go`:
  - `TestCreatePacientePorcentajeGrasaFueraDeRango` (table-driven: negativo y > 100 → `400`).
  - `TestUpdatePacientePorcentajeGrasaFueraDeRango` (> 100 en `PUT` → `400`).
  - `TestPatchPacienteValidacionesExtra` (table-driven: `porcentaje_grasa` negativo, > 100 y `nombre` solo espacios en `PATCH` → `400`).
  - `TestDeletePacienteOK` (`204 No Content` y baja efectiva en el repositorio) y `TestDeletePacienteNoEncontrado` (`404 Not Found`).
  - `TestGetPacientesPorEdadInvalida` (`edad` no numérica → `400 Bad Request`).
- Saneamiento documental: restaurado `PROJECT_INDEX.md` completo, cerrado y archivado `odd/tasks/archive/feature-b-paginacion-firestore.md` y sincronizado `odd/task/estado-actual-proyecto.md` y `todo`.
- `go test ./...`, `go vet ./...`, `go build ./...` todos verdes.
- Commits: `a0ad130` (saneamiento documental) y `0fd27a6` (validaciones y tests v1.0.7).

**Dificultades encontradas y por qué se construyó así:**

1. **Pérdida accidental de secciones en `PROJECT_INDEX.md` en el working tree previo.** Al iniciar la sesión, `PROJECT_INDEX.md` tenía 269 líneas menos porque una edición inconclusa había sobrescrito las secciones `1` a `6.7` dejando solo un bloque `6.8` con tokens corruptos en otros idiomas (`keine`, `测试端点`, `接收`). Se recuperó la base desde `git show HEAD:PROJECT_INDEX.md`, se actualizó la firma del puerto `GetAll(ctx, limit, offset int)` y se separó la restauración documental en un commit propio (`a0ad130`) antes de tocar código funcional (`0fd27a6`) para mantener unidades de trabajo limpias y revisables.

2. **Asimetría de invariantes entre `POST`/`PUT` y `PATCH`.** `CreatePaciente` y `UpdatePaciente` pasaban por `normalizarNombre` y luego por `validarPaciente`, rechazando `"   "` como nombre vacío. En cambio, `PatchPaciente` aplicaba `normalizarNombre` sobre el mapa JSON sin verificar si el string resultante quedaba vacío, permitiendo dejar un paciente sin nombre mediante una actualización parcial. Además, `porcentaje_grasa` no tenía cota `[0, 100]`. Se unificó el invariante con la constante `porcentajeGrasaMaximo = 100` en ambos flujos y el chequeo `if existente.Nombre == ""` dentro de `PatchPaciente`, preservando la semántica parcial de `PATCH` (solo se validan los campos efectivamente enviados).

3. **Desajuste de estructura JSON durante la fase RED de los tests.** Al correr los tests nuevos en fase RED (antes de implementar la validación), los handlers devolvían `201 Created` o `200 OK` con el struct `Paciente` (donde `edad` es `int`), lo que hacía fallar el `json.Unmarshal` hacia `map[string]string` en el test. Esto confirmó que el test estaba probando la frontera real HTTP; una vez agregada la validación en fase GREEN, el handler responde `400 Bad Request` con `map[string]string{"error": "..."}` y tanto el código de estado como el mensaje pasan de forma consistente.

4. **Verificación de estado real en `TestDeletePacienteOK`.** Para no limitarse a comprobar el código HTTP `204 No Content`, el test consulta inmediatamente `repo.GetByID(ctx, creado.ID)` y verifica que devuelva error, garantizando que el borrado tuvo efecto real en el almacenamiento en memoria.

---
## Versión 1.0.6 - 2026-11-08

### Nuevo endpoint `GET /pacientes/edad/{edad}` (filtro por edad exacta)

- Handler `GetPacientesPorEdad` en `handlers/paciente.go`: lee `c.Param("edad")`, convierte a entero, consulta `repo.GetAll` y devuelve solo los pacientes cuya edad coincida exactamente. Responde `200 OK` con lista (vacía si no hay coincidencia).
- Ruta `GET /pacientes/edad/{edad}` registrada en `SetupRoutes`, antes de `/pacientes/{id}` para evitar ambigüedad con el parámetro de ruta.
- Validación: `edad` requerido; si no es un número entero válido → `400 Bad Request` con mensaje `"El campo 'edad' debe ser un número entero"`.
- Tests agregados en `handlers/paciente_test.go`: `TestGetPacientesPorEdadOK` (edad 30 → Juan Perez) y `TestGetPacientesPorEdadVacio` (edad 99 → lista vacía).
- `go test ./...`, `go vet ./...`, `go build ./...` todos verdes.

**Dificultades encontradas y por qué se construyó así:**

1. **`c.Param("edad")` vacío al invocar el handler directamente desde el test.** Como ocurre con todos los handlers que usan parámetros de ruta (`GetPacienteByID`, `UpdatePaciente`, `DeletePaciente`), Echo v4 almacena `pnames`/`pvalues` en campos no exportados del contexto y solo los llena al pasar por el router. Se reutilizó el helper `setParam` (reflect + unsafe.Pointer) para inyectar `edad` en el contexto de test, manteniendo consistencia con el patrón existente del proyecto.

2. **Orden de rutas en `SetupRoutes`.** Se registró `/pacientes/edad/{edad}` antes que `/pacientes/{id}` para que el segmento estático `edad` tenga prioridad sobre el parámetro `{id}` y el router no intente interpretar `edad` como un ID de paciente.

3. **Límite de paginación `limit=0, offset=0` para devolver todos.** `GetPacientesPorEdad` llama a `repo.GetAll(ctx, 0, 0)` para obtener todos los pacientes y filtrar en memoria, igual que `SearchPacientes` y `StatsPacientes`. El mock acepta `limit=0` (sin límite) y `offset=0`.

4. **Tipo `int` para la edad en el filtro.** La edad del modelo `Paciente` es `int`, por lo que `strconv.Atoi` es suficiente sin `ParseFloat`, manteniendo coherencia con el dominio.

---\n## Versión 1.0.5 - 2026-11-07

### Validaciones de negocio extra y normalización de nombres

- `validarPaciente` en `handlers/paciente.go`: `edad` debe estar entre 1 y 120 (antes solo > 0); `peso` debe ser mayor que 0 y menor o igual a 500 (antes solo > 0). Los mensajes de error incluyen el rango permitido.
- Función `normalizarNombre`: elimina espacios sobrantes (`strings.TrimSpace`) y deja cada palabra con primera letra mayúscula y el resto minúscula (`strings.ToUpper`/`strings.ToLower` + `strings.Join`).
- `CreatePaciente` y `UpdatePaciente`: normalizan el nombre antes de validar y guardar. `PatchPaciente`: normaliza el nombre cuando se envía en el body.
- Nombre compuesto solo por espacios → 400 (`"El campo 'nombre' es requerido"`).
- Edad fuera de rango (ej. 150) → 400 con mensaje que menciona `edad`. Peso fuera de rango (ej. 600) → 400 con mensaje que menciona `peso`.
- Tests agregados en `handlers/paciente_test.go`: `TestCreatePacienteEdadFueraDeRango`, `TestCreatePacientePesoFueraDeRango`, `TestCreatePacienteNombreConEspaciosSeNormaliza`, `TestCreatePacienteNombreSoloEspacios`.
- `go test ./...`, `go vet ./...`, `go build ./...` todos verdes.
- Commit: `f21a6a8 feat: validaciones de negocio extra y normalizar nombres en pacientes`.

**Dificultades encontradas y por qué se construyó así:**

1. **Rangos de edad y peso elegidos de forma deliberada.** La validación original solo rechazaba `<= 0`, lo que permitía valores absurdos (edad 999, peso 10000). Se eligió edad 1-120 y peso 0-500 como límites clínicos razonables para un sistema de nutrición, sin ser restrictivos para casos edge reales. El límite de peso usa `%d` (int) en el mensaje porque la constante `pesoMaximo` se declara como `int`, no `float64`.

2. **Normalización solo en crear y actualizar completa, no en PATCH parcial de otros campos.** El nombre se normaliza en `CreatePaciente`, `UpdatePaciente` y `PatchPaciente` (cuando se envía). No se normalizan otros campos (edad, peso) porque son numéricos y no requieren limpieza de texto.

3. **`fmt.Errorf` vs `fmt.Sprintf` en PATCH.** En `validarPaciente` se usa `fmt.Errorf` (retorna `error`); en `PatchPaciente` se usa `fmt.Sprintf` dentro de `c.JSON` porque el valor se interpola directamente en un `map[string]string`, no se retorna como `error`.

4. **El tipo de `pesoMaximo` causó un error de compilación.** Inicialmente se usó `%.0f` con `pesoMaximo` (`int`), lo que falló en `go vet`. Se corrigió a `%d` para coincidir con el tipo `int`.

## Versión 1.0.4 - 2026-11-02

### PATCH `PATCH /pacientes/{id}` (actualización parcial)

- `PatchPaciente` en `handlers/paciente.go`: aplica solo los campos enviados en el body (nombre, edad, peso, porcentaje_grasa, notas). Los campos no enviados se conservan.
- Requiere que el paciente exista (`GetByID`), si no existe devuelve 404.
- Valida `edad > 0` y `peso > 0` solo cuando esos campos se envían; 400 si inválidos.
- Ruta `PATCH /pacientes/{id}` registrada en `SetupRoutes`.
- Tests agregados: `TestPatchPacienteOK`, `TestPatchPacienteSoloPeso`, `TestPatchPacienteNoEncontrado`, `TestPatchPacienteEdadInvalida`.
- `go test ./...`, `go vet ./...`, `go build ./...` todos verdes.
- Commit: `9c6f605 feat: agregar PATCH /pacientes/{id} actualización parcial y archivar tracking C.4`.
- Tracking ODD archivado: `odd/tasks/feature-c4-search.md` movido a `odd/tasks/archive/`.
- Estado sincronizado en `odd/task/estado-actual-proyecto.md` (T11 + L10) y `PROJECT_INDEX.md`.

## Versión 1.0.3 - 2026-11-02

### [C.6] Completar `PUT /pacientes/{id}` con tests de éxito y no encontrado

**¿Qué se hizo y por qué:**
El handler `UpdatePaciente` y la ruta `PUT /pacientes/{id}` ya existían desde la versión 1.0.0, con validación de ID (`c.Param("id")`) y delegación al repositorio `Update`. Lo que faltaba para cerrar C.6 eran los **tests de éxito (200) y no encontrado (404)**, porque sin ellos el endpoint no tenía cobertura de su caso principal: actualizar un paciente existente y devolverlo con los datos nuevos, ni el caso de error cuando el ID no existe.

**Tests agregados en `handlers/paciente_test.go`:**
- `TestUpdatePacienteOK`: crea un paciente con `repo.Create` (el mock le asigna `paciente-<n>`), luego llama `UpdatePaciente` con payload nuevo y verifica `200 OK`, que `ID` se conserva y que `nombre`, `edad`, `peso`, `porcentaje_grasa` y `notas` se actualizan.
- `TestUpdatePacienteNoEncontrado`: llama `UpdatePaciente` con ID `inexistente` y verifica `404 Not Found` con mensaje `"No se encontró paciente"`.

**Tests preexistentes que cubren la validación de negocio:**
- `TestUpdatePacienteEdadInvalida`: `edad <= 0` → `400 Bad Request` con mensaje sobre `edad`.
- `TestUpdatePacientePesoInvalido`: `peso <= 0` → `400 Bad Request` con mensaje sobre `peso`.

**Diseño:**
- El mock `Update` modifica el paciente en `m.store` por índice y lo devuelve, así que el handler solo pasa el ID de `c.Param("id")` y el cuerpo parseado. Valida, delega al repo, devuelve el resultado.
- El helper `setParam` (reflect + unsafe.Pointer) inyecta `pnames`/`pvalues` en el contexto de Echo v4 para que `c.Param("id")` funcione cuando el test invoca el handler directo, sin pasar por el router.
- Commit: `85ed67f feat: completar C.6 PUT /pacientes/{id} con tests de éxito y no encontrado`.
- `go test ./...`, `go vet ./...`, `go build ./...` todos verdes.

## Versión 1.0.2 - 2026-11-02

### [C.5] Endpoint `GET /pacientes/stats` (resumen de solo lectura)
- `StatsPacientes` devuelve `{total, edad_promedio, peso_promedio, porcentaje_grasa_promedio, edades[], pesos[]}` agregando en memoria sobre `repo.GetAll`. Solo lectura, no modifica datos.
- Ruta `GET /pacientes/stats` registrada en `SetupRoutes`.
- Tests agregados: `TestStatsPacientesOK` y `TestStatsPacientesListaVacia`.
- `go test ./...`, `go vet ./...`, `go build ./...` todos verdes.

## Versión 1.0.1 - 2026-10-26

### [C.2] Mock asigne ID autoincremental al crear + GetByID/Delete
- `MockPacienteRepository.Create` ahora asigna `paciente-<n>` automáticamente cuando `p.ID` viene vacío, usando `atomic.AddInt64(&m.nextID, 1)`. Respeta el ID si ya viene seteado.
- Se implementaron `GetByID`, `Update` y `Delete` en el mock para permitir operaciones sobre datos creados en memoria.
- Handler `GetPacienteByID` (GET `/pacientes/{id}`): devuelve 200 con el paciente o 404 si no existe.
- Handler `DeletePaciente` (DELETE `/pacientes/{id}`): devuelve 204 o 404 si no existe.
- Tests agregados: `TestMockCreateAsignaIDCuandoEstaVacio`, `TestMockCreateRespetaIDExistente`, `TestGetPacienteByIDOK`, `TestGetPacienteByIDNoEncontrado`.
- Commit: `65b514d feat: tests de GetPacienteByID y mock asigna ID al crear`.

### [C.3] Endpoint `GET /pacientes?nombre=` (filtro por nombre)
- `GetPacientes` filtra por `c.QueryParam("nombre")` con coincidencia parcial y sin distinción de mayúsculas/minúsculas.
- Test agregado: `TestGetPacientesFiltrarPorNombre`.
- Commit: `9102ca9 feat: filtrar pacientes por nombre en GET /pacientes`.

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