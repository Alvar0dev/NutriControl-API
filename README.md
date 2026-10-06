# NutriControl API - Bitácora de Proyecto

## Día 1: Setup y Fundamentos
- **Objetivo**: Inicializar el proyecto con bases sólidas y probar el enrutamiento básico.
- **Acciones Realizadas**:
  - Verificación del entorno de ejecución de Go.
  - Inicialización del módulo de Go (`go mod init nutricontrol`).
  - Instalación del framework web **Echo** (`github.com/labstack/echo/v4`).
  - Creación de `main.go` con un servidor básico y un endpoint `GET /` de chequeo de estado ("API NutriControl Operativa").
- **Conceptos Aplicados**: *Keep it simple.* Usar herramientas minimalistas que te den el control total en lugar de frameworks mágicos que te ocultan la complejidad.

---

## Día 2: Modelado y Primeros Endpoints (Caso de Uso: Pacientes)
- **Objetivo**: Estructurar la lógica para la gestión de pacientes y separar las capas de la aplicación.
- **Acciones Realizadas**:
  - Definición de los datos del dominio: creación del struct `Paciente` en `models/paciente.go`.
  - Creación de controladores (puertos de entrada): endpoints mockeados para `GET /pacientes` y `POST /pacientes` en `handlers/paciente.go`.
  - Refactorización del entrypoint: actualización de `main.go` para integrar los handlers externos sin ensuciar la función principal.
- **Conceptos Aplicados**: *Separation of Concerns* (Separación de responsabilidades). Los modelos de dominio no se mezclan con los controladores HTTP. Esto es el primer paso hacia una Arquitectura Limpia/Hexagonal.
