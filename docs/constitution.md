# Constitución de NutriControl API

## Principios Innegociables

### 1. Simplicidad del Stack
- Usar herramientas minimalistas que den control total vs. frameworks "mágicos"
- Limitar framework HTTP a solo lo esencial para desarrollo rápido y mantenimiento
- Rechazar dependencias innecesarias que oculten complejidad

### 2. Separación de Responsabilidades
- Models (dominio) nunca se mezclan con handlers (HTTP/presentation)
- Uso consistente de estructuras de paquetes limpios y bien definidos
- Validación de entrada en handlers, lógica de negocio pura en modelos

### 3. Especificación Primaria
- Los contratos de API guían el diseño de código, no al revés
- Todo código servirá como implementación de especificaciones claras
- Documentación automática derivada de especificaciones, no al contrario

### 4. Persistencia Controlada
- Base de datos SQL por defecto (Go + sqlx/SQL package)
- Implementación progresiva: primero mocks, luego persistente, nunca ambos a la vez
- Separación clara entre interfaz (puertos) e implementación (adaptadores)

### 5. Testing Purista
- Testing de unidad para componentes aislados
- Testing de integración para dependencias externas (base de datos, HTTP)
- Testing de contrato para endpoints vs. implementación (mock vs. real)

### 6. Errores como Características
- Tratamiento explícito de errores en las fronteras del sistema
- Tipado contextual de errores para clientes HTTP/API
- Sin reintentos ocultos o cachés que enmascaren problemas

### 7. Código como Objeto de Aprendizaje
- Cada componente nuevo fundamentado en fundamentos Go sólidos
- Composición sobre herencia, composición sobre patrones complejos
- Elección consciente de minimalismo vs. complejidad

### 8. Iteración por Especificaciones
- Cada cambio de código originado de especificaciones actualizadas
- Versiónamiento semántico guiado por cambios contractuales
- Base de especificaciones como verdad única del sistemaActualizado ODD: estructura, tracking odd/tasks/, SDD archivado eliminado (specs/archive/). Contratos: Go+Echo, Firestore/mock, manual DI, no tocar persistencia sin acuerdo.
