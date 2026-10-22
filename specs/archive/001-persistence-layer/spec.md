# Spec 001 — Capa de Persistencia con Firestore

## Contexto y Objetivo
El proyecto NutriControl API se encuentra en fase inicial (mocks en memoria). El objetivo de esta especificación es reemplazar los datos mockeados actuales con una capa de persistencia real basada en Firestore (NoSQL, archivo de base de datos local). Se mantiene el enfoque minimalista (Go + Echo) y la separación de responsabilidades definida en la Constitución.

## Historias de Usuario
1. **Como usuario del sistema**, quiero que los pacientes creados mediante `POST /pacientes` se guarden en Firestore para que la información persista entre ejecuciones.
2. **Como usuario del sistema**, quiero consultar los pacientes almacenados mediante `GET /pacientes` para ver los datos reales guardados.

## Requisitos Funcionales (EARS)

- **RF-01** (Creación persistente): Cuando se recibe una petición `POST /pacientes` con datos válidos del paciente, el sistema deberá guardar el paciente en Firestore, generar automáticamente un `id`, y devolver `201 Created` con el paciente completo (incluido `id`).
- **RF-02** (Consulta persistente): Cuando se recibe una petición `GET /pacientes`, el sistema deberá recuperar todos los pacientes almacenados en Firestore y devolver `200 OK` con la lista completa; si no existen pacientes, devolver una lista vacía.
- **RF-03** (Validación básica): Cuando el cuerpo de la petición `POST /pacientes` incluya datos incompletos (nombre vacío, edad no numérica, peso no numérico), el sistema deberá rechazar la petición con `400 Bad Request` y un mensaje de error descriptivo.
- **RF-04** (Captura de errores): Cuando ocurra un error interno al interactuar con Firestore (fallo de lectura, escritura o conexión), el sistema deberá capturar el error con descripción y devolver `500 Internal Server Error` con un mensaje descriptivo.

## Casos Límite y Edge Cases
- Paciente con `edad` = 0 o `peso` = 0: aceptado si es un valor numérico válido.
- Colección `pacientes` vacía en `GET /pacientes`: devolver `[]` vacío con `200 OK`.
- Creación duplicada: cada `POST` genera un nuevo documento con ID automático (no se impide duplicación de datos en esta fase).
- Fallo de archivo de base de datos de Firestore: capturado con mensaje descriptivo (`500`).

## Definition of Done
- [ ] `POST /pacientes` persiste los datos en Firestore y devuelve `201` con `id` generado automáticamente.
- [ ] `GET /pacientes` recupera datos reales de Firestore, no datos mockeados.
- [ ] La validación básica rechaza datos inválidos con `400`.
- [ ] Cualquier error de Firestore se captura con descripción y devuelve `500`.
- [ ] El código existente (`main.go`, `models/`, `handlers/`) se mantiene funcional; no se rompen contratos actuales.
- [ ] No se toca ninguna carpeta fuera de `specs/001-persistence-layer/`, `models/`, `handlers/`, ni archivos de código sin autorización explícita.

## Fuera del Alcance (MVP)
- Endpoints de `PUT`, `DELETE` y `GET` por ID específico.
- Configuración de variables de entorno externas (archivo `.env` o `config/` estructurado).
- Logging avanzado o trazabilidad estructurada.
- Tests automatizados completos (se definen tareas para test básico en esta spec, pero sin infraestructura completa de testing).
- Migraciones de datos o manejo de versiones entre cambios de estructura.
