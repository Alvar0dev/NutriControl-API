# DEV_RUNBOOK — NutriControl API

> Comandos exactos para compilar, correr, testear y linter. Requisitos de entorno y gotchas.

## 1. Requisitos previos
- **Go 1.26.1+** (`go version`).
- **Internet** (primera vez): `go get` descarga deps en `go.sum`.
- **Opcional:** `GCP_PROJECT_ID` + credenciales GCP para Firestore real (ver §2).

## 2. Variables de entorno

| Variable | Obligatorio | Valor | Propósito |
|----------|-------------|-------|-----------|
| `NUTRICONTROL_USE_MOCK` | No | `true` | Forzar mock en memoria (sin Firestore). |
| `GCP_PROJECT_ID` | No | `<id>` | ID proyecto GCP para Firestore real. |
| `GOOGLE_APPLICATION_CREDENTIALS` | No | `/path/cred.json` | Credenciales de servicio GCP. |

**Comportamiento de arranque (`main.go`):**
- Si `NUTRICONTROL_USE_MOCK=true` → siempre mock.
- Si no, y `GCP_PROJECT_ID` vacío → fallback a mock con warning.
- Si no, y `GCP_PROJECT_ID` seteado → intenta Firestore real; si falla, fallback a mock con warning.

## 3. Comandos

### 3.1 Dependencias
```bash
go mod tidy            # Limpia go.sum/go.mod
go get cloud.google.com/go/firestore google.golang.org/api/iterator google.golang.org/api/option
go get github.com/stretchr/testify/assert github.com/stretchr/testify/require
```

### 3.2 Build / Run
```bash
go build -o nutricontrol.exe ./main.go   # Compila binario
go run ./...                              # Ejecuta main.go directo
```
Puerto por defecto: **:8080**.

### 3.3 Testing
```bash
go test ./...                    # Todos los paquetes
go test -v ./handlers/           # Solo handlers (verbose)
go test -v ./models/             # Solo models
go test -cover ./...             # Con cobertura
```
**Estado actual:** `models/` OK. `handlers/` **no compila** (Echo v4 `Response.Body` eliminado). Ver §4.

### 3.4 Lint (opcional, no configurado)
```bash
go vet ./...                     # Verificación estándar
golangci-lint run ./...         # Si se instala
```

## 4. Gotchas y peculiaridades

### 4.1 Tests handlers (Echo v4) — resuelto en `b3754f1`
`handlers/paciente_test.go` usaba `c.Response().Body.Bytes()` (eliminado en Echo v4). Se corrigió a `c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes()` en commit `b3754f1`. `go test ./...` pasa: `handlers` y `models` verdes.

### 4.2 Mock asigne ID automático al crear — resuelto en `65b514d`
`MockPacienteRepository.Create` ahora asigna `paciente-<n>` automáticamente cuando `p.ID` viene vacío, usando `atomic.AddInt64(&m.nextID, 1)`. Respeta el ID si ya viene seteado. Si creas un paciente sin ID y luego buscas por ID, lo encontrarás. Se implementaron además `GetByID`, `Update` y `Delete` en el mock.

### 4.3 Firestore GetByID falla con error genérico
El handler mapea cualquier error de `repo.GetByID` a 404. Si Firestore falla por otra razón (red, permisos), el cliente recibe 404 en lugar de 500. No es ideal, pero es la lógica actual.

### 4.4 Validación mínima
Se valida `nombre != ""`, `edad > 0` y `peso > 0` en `handlers.validarPaciente` (commit `8a1a47f`). 400 con `{"error": "..."}` si falta o es inválido. Ya no se acepta `edad=0` ni `peso=0`.

### 4.5 `go.sum` grande
Primera vez que se agrega `cloud.google.com/go/firestore`, `go.sum` crece ~30 líneas. Asegúrate de hacer `go mod tidy` antes de commit.

### 4.6 Binario local
`nutricontrol.exe` (Windows) puede quedar en raíz tras `go build`. No es parte del repo (ver `.gitignore` si existe).

## 5. Verificación rápida
```bash
NUTRICONTROL_USE_MOCK=true go run ./... &
curl -s http://localhost:8080/pacientes | jq
curl -s -X POST http://localhost:8080/pacientes -H 'Content-Type: application/json' -d '{"nombre":"Test","edad":30,"peso":70}' | jq
```

## 6. Flujo de commits y convenciones
- Estilo: **Conventional Commits** (`feat:`, `fix:`, `refactor:`, `docs:`, `build:`, `chore:`).
- Rama de trabajo: crear feature branch desde `main` antes del primer cambio de fuente; `main` se mantiene estable y verde.
- Antes de commitear: `go mod tidy` (ajusta `go.sum`), `go build ./...` y `go test ./...`.
- Push y creación de PR quedan a criterio del usuario bajo política ordinaria del repositorio.

## 7. Próximos pasos (según `todo` en raíz)
- `B`) Completar `PROJECT_INDEX.md` / `DEV_RUNBOOK.md` (en curso — estado actual documentado en §7 de `PROJECT_INDEX.md`).
- `C`) Nueva feature: extender repo Firestore/mock o agregar un endpoint nuevo (pendiente de elección en próxima sesión).