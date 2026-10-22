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

### 4.1 Tests handlers rotos (preexistente)
`handlers/paciente_test.go` usa `c.Response().Body.Bytes()` — no existe en Echo v4.16.0. No es un bug de los cambios recientes. Para arreglar: usar `httptest.NewRecorder()` directamente y acceder a `rec.Body.Bytes()`.

### 4.2 Mock no persiste `ID` por omisión
`MockPacmanRepository` no setea `ID` al crear (Create devuelve el struct tal cual). Si creas un paciente y luego buscas por ID, no lo encontrarás. Para tests de GetByID/Delete, setea `p.ID` manualmente o actualiza el mock para setearlo.

### 4.3 Firestore GetByID falla con error genérico
El handler mapea cualquier error de `repo.GetByID` a 404. Si Firestore falla por otra razón (red, permisos), el cliente recibe 404 en lugar de 500. No es ideal, pero es la lógica actual.

### 4.4 Validación mínima
Solo se valida `nombre != ""`. `edad` y `peso` no se validan (RF-03 de spec antigua no implementada). Acepta `edad=0` o `peso=0`.

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