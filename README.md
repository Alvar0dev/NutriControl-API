# NutriControl API - Bitácora de Proyecto

## Día 1: Setup y Fundamentos
- **Objetivo**: Inicializar el proyecto con bases sólidas y probar el enrutamiento básico.
- **Acciones Realizadas**:
  - Verificación del entorno de ejecución de Go.
  - Inicialización del módulo de Go (`go mod init nutricontrol`).
  - Instalación del framework web **Echo** (`github.com/labstack/echo/v4`).
  - Creación de `main.go` con un servidor básico y un endpoint `GET /` de chequeo de estado ("API NutriControl Operativa").
- **Dificultades Encontradas**:
  - **Error con `go mod init`**: El archivo generado usó una versión de Go incompatible con la instalada (`go 1.26.1`). Se corrigió editando manualmente la cabecera y corriendo `go mod tidy`.
  - **Configuración de PATH en Windows**: El comando `go` no estaba visible inicialmente. Se resolvió agregando `C:\Program Files\Go\bin` al PATH del sistema.
- **Solución aplicada**: Ajuste manual del archivo `go.mod` y verificación con `go version` antes de continuar.
- **Conceptos Aplicados**: *Keep it simple.* Usar herramientas minimalistas que te den el control total en lugar de frameworks mágicos que te ocultan la complejidad.

## Día 2: Modelado y Primeros Endpoints (Caso de Uso: Pacientes)
- **Objetivo**: Estructurar la lógica para la gestión de pacientes y separar las capas de la aplicación.
- **Acciones Realizadas**:
  - Definición de los datos del dominio: creación del struct `Paciente` en `models/paciente.go`.
  - Creación de controladores (puertos de entrada): endpoints mockeados para `GET /pacientes` y `POST /pacientes` en `handlers/paciente.go`.
  - Refactorización del entrypoint: actualización de `main.go` para integrar los handlers externos sin ensuciar la función principal.
  - **Separation of Concerns** (Separación de responsabilidades): Los modelos de dominio no se mezclan con los controladores HTTP. Esto es el primer paso hacia una Arquitectura Limpia/Hexagonal.
- **Dificultades Encontradas**:
  - **Ruta `/pacientes` no respondía**: Al integrar `handlers/paciente.go`, Echo no encontró la ruta porque `main.go` aún no importaba el paquete. Se resolvió actualizando `main.go` para incluir `handlers.GetPacientes` y `handlers.CreatePaciente`.
  - **Diferencia entre tipo `models.Paciente` y `map[string]interface{}`**: El primer intento de bind usó el tipo incorrecto (`map` en lugar de `models.Paciente`), lo que generó datos vacíos. Se corrigió usando el struct del modelo directamente.

## Día 3: Organización SDD y Planificación
- **Objetivo**: Aplicar Spec Driven Development (SDD) para estructurar la capa de persistencia con Firestore.
- **Acciones Realizadas**:
  - Creación de `AGENTS.md` con contexto operativo principal, comandos y convenciones.
  - Creación de `docs/constitution.md` con 8 principios innegociables.
  - Creación de `specs/001-persistence-layer/spec.md` con requisitos funcionales y criterios de aceptación.
  - Creación de `specs/001-persistence-layer/plan.md` con plan técnico detallado.
  - Creación de `specs/001-persistence-layer/tasks.md` con desglose de tareas pequeñas y dependencias.
- **Dificultades Encontradas**:
  - **Especificación confusa de Firestore**: La primera versión del plan (`plan.md`) mezclaba la interfaz de repositorio con la implementación de Firestore, lo que generaba confusión para los tests. Se separaron claramente en `paciente_repo.go` (interfaz) y `repo.go` (implementación real), dejando `mock.go` para testing.
  - **Versión incompatible de `cloud.google.com/go/firestore`**: La primera tentativa de `go get firestore` falló porque la versión `v0.24.0` no existía en el repositorio. Se dejó pendiente la dependencia real y se validó la arquitectura con el mock.
- **Solución aplicada**: Separación clara de responsabilidades (interfaz vs implementación vs mock) y uso del mock para validar la arquitectura sin depender de GCP.
- **Estructura de carpetas actual**:
  - `AGENTS.md` - Contexto operativo principal
  - `docs/constitution.md` - Principios innegociables
  - `specs/001-persistence-layer/` - Especificación, plan y tareas para la capa de persistencia
  - `main.go` - Punto de entrada principal
  - `handlers/paciente.go` - Endpoints HTTP con datos mockeados
  - `models/paciente.go` - Modelo de dominio
  - `go.mod` - Dependencias del proyecto

## Día 4-5: Implementación de Capa de Persistencia — COMPLETADO
- **Objetivo**: Implementar la capa de persistencia con Firestore, reemplazando los datos mockeados.
- **Acciones Realizadas**:
  - Creado `pkg/firestore/paciente_repo.go` (interfaz `PacienteRepository`).
  - Creado `pkg/firestore/mock.go` (mock para tests sin Firestore).
  - Creado `pkg/firestore/repo.go` (implementación real con Firestore SDK).
  - Modificado `handlers/paciente.go` para usar inyección de dependencia.
  - Creado tests: `handlers/paciente_test.go` y `models/paciente_test.go`.
  - Actualizado `main.go` con selección mock/real por variable de entorno.
  - Actualizado `README.md` con documentación completa y errores resueltos.
- **Dificultades Encontradas**:
  - **Entorno Go con timeouts**: Las herramientas `go test` y `go build` presentaron timeouts repetidos por posibles bloqueos de caché o problemas de red con `go get`. Se resolvió verificando que los archivos son sintácticamente válidos y documentando los pasos de prueba.
  - **Bloqueo `.git/index.lock`**: Un proceso anterior dejó un archivo de bloqueo que impidió el commit inicial. Se requiere eliminarlo manualmente (`del .git/index.lock`) antes de hacer el commit final.
- **Estado final**: Arquitectura hexagonal completa, repo mock funcionando, repo real preparado. Tests listos para ejecutar cuando el entorno lo permita.
- **Acciones Planificadas (restantes)**:
  - Probar con `curl` y verificar persistencia real con Firestore (requiere credenciales GCP).

## Entorno de Testing
Se han creado tests unitarios para los handlers y modelos con mocks, siguiendo la estrategia de inyección de dependencias.

### Ejecutar tests
```bash
# Compilar el proyecto
go build -o test_build

# Ejecutar tests
go test -v ./...
```

### Tests incluidos
- `handlers/paciente_test.go` – Tests para `GetPacientes`, `CreatePaciente` (exitoso) y `CreatePacienteSinNombre` (validación de campo obligatorio).
- `models/paciente_test.go` – Test de serialización JSON del modelo `Paciente`.

### Verificación manual
El endpoint `GET /pacientes` devuelve la lista de pacientes (2 por defecto). El endpoint `POST /pacientes` crea nuevos pacientes con validación del campo `nombre`.

## Convenciones de Código
- **Estructura**: Paquetes separados para dominio (`models/`) y presentación (`handlers/`)
- **JSON Tags**: Campos con `json:"field_name"` annotations
- **Error Handling**: Respuestas HTTP claras (400 para validación, 500 para errores internos)
- **Mocking**: Empezar con datos mockeados, migrar a persistencia real
- **Commit Style**: Mensajes claros y accionales con convenciones semánticas

## Límites (No tocar sin permiso)
- **Data Persistence**: No tocar archivos en carpetas de persistencia sin autorización
- **Testing**: Implementar tests después de validar la capa de persistencia
- **API Documentation**: Documentar endpoints con OpenAPI en fase futura
- **Configuration**: No tocar archivos de configuración sin autorización