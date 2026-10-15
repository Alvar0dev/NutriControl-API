# NutriControl API - Bitácora de Proyecto

## Día 1: Setup y Fundamentos
- **Objetivo**: Inicializar el proyecto con bases sólidas y probar el enrutamiento básico.
- **Acciones Realizadas**:
  - Verificación del entorno de ejecución de Go.
  - Inicialización del módulo de Go (`go mod init nutricontrol`).
  - Instalación del framework web **Echo** (`github.com/labstack/echo/v4`).
  - Creación de `main.go` con un servidor básico y un endpoint `GET /` de chequeo de estado ("API NutriControl Operativa").
- **Conceptos Aplicados**: *Keep it simple.* Usar herramientas minimalistas que te den el control total en lugar de frameworks mágicos que te ocultan la complejidad.

## Día 2: Modelado y Primeros Endpoints (Caso de Uso: Pacientes)
- **Objetivo**: Estructurar la lógica para la gestión de pacientes y separar las capas de la aplicación.
- **Acciones Realizadas**:
  - Definición de los datos del dominio: creación del struct `Paciente` en `models/paciente.go`.
  - Creación de controladores (puertos de entrada): endpoints mockeados para `GET /pacientes` y `POST /pacientes` en `handlers/paciente.go`.
  - Refactorización del entrypoint: actualización de `main.go` para integrar los handlers externos sin ensuciar la función principal.
  - **Separation of Concerns** (Separación de responsabilidades): Los modelos de dominio no se mezclan con los controladores HTTP. Esto es el primer paso hacia una Arquitectura Limpia/Hexagonal.

## Día 3: Organización SDD y Planificación
- **Objetivo**: Aplicar Spec Driven Development (SDD) para estructurar la capa de persistencia con Firestore.
- **Acciones Realizadas**:
  - Creación de `AGENTS.md` con contexto operativo principal, comandos y convenciones.
  - Creación de `docs/constitution.md` con 8 principios innegociables.
  - Creación de `specs/001-persistence-layer/spec.md` con requisitos funcionales y criterios de aceptación.
  - Creación de `specs/001-persistence-layer/plan.md` con plan técnico detallado.
  - Creación de `specs/001-persistence-layer/tasks.md` con desglose de tareas pequeñas y dependencias.
- **Estructura de carpetas actual**:
  - `AGENTS.md` - Contexto operativo principal
  - `docs/constitution.md` - Principios innegociables
  - `specs/001-persistence-layer/` - Especificación, plan y tareas para la capa de persistencia
  - `main.go` - Punto de entrada principal
  - `handlers/paciente.go` - Endpoints HTTP con datos mockeados
  - `models/paciente.go` - Modelo de dominio
  - `go.mod` - Dependencias del proyecto

## Día 4-5: Implementación de Capa de Persistencia
- **Objetivo**: Implementar la capa de persistencia con Firestore, reemplazando los datos mockeados.
- **Acciones Planificadas**:
  - Implementar repositorio Firestore (`pkg/firestore/`)
  - Modificar handlers para usar el repositorio
  - Añadir validación básica y manejo de errores
  - Probar con `curl` y verificar persistencia

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