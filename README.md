# NutriControl API 🍏

> Sistema de gestión de historiales clínicos y analíticas, diseñado para profesionales de la nutrición deportiva y evaluación antropométrica.

Este proyecto nace de la necesidad real de digitalizar y optimizar el seguimiento de pacientes, combinando mi trayectoria previa de más de 15 años en asesoría nutricional y atención al cliente con el desarrollo de software moderno. 

Actualmente lo desarrollo como proyecto práctico integral durante mi formación en el Grado Superior en Desarrollo de Aplicaciones Multiplataforma (DAM). El objetivo técnico es implementar una arquitectura backend robusta y prepararla para producción utilizando despliegues optimizados.

---

## 🚀 Stack Tecnológico

El ecosistema de la aplicación está construido con las siguientes tecnologías:

*   **Backend:** Go (Golang) utilizando el framework **Echo** para el enrutamiento.
*   **Base de Datos:** Google Cloud **Firestore** (NoSQL).
*   **Frontend:** HTML5, CSS3 y Vanilla JavaScript (sin frameworks externos).
*   **Infraestructura (DevOps):** Contenerización mediante **Docker** (Multi-stage build) preparada para despliegue en Google Cloud Platform (Cloud Run).
*   **Control de Versiones:** Git y GitHub.

---

## ⚙️ Endpoints Principales

La API RESTful gestiona las siguientes operaciones clave:

*   `GET /api/pacientes` - Obtiene el listado completo de pacientes.
*   `POST /api/pacientes` - Registra un nuevo paciente (Nombre, Edad, Peso inicial, Objetivos).
*   `GET /api/pacientes/:id` - Recupera el historial y evolución de un paciente específico.
*   `POST /api/pacientes/:id/analiticas` - Añade una nueva revisión (Ej: % Grasa, Peso actual, Notas de prescripción).

---

## 🛠️ Instalación y Ejecución Local

Sigue estos pasos para desplegar el proyecto en tu máquina local:

### 1. Clonar el repositorio
```bash
git clone [https://github.com/Alvarodev/nutricontrol-api.git](https://github.com/Alvarodev/nutricontrol-api.git)
cd nutricontrol-api
