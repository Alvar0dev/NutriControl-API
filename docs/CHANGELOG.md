# CHANGELOG.md

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