package handlers

import (
	"net/http"

	"nutricontrol/models"
	"nutricontrol/pkg/firestore"
)

// GetPacientes obtiene todos los pacientes mediante el repositorio.
func GetPacientes(c echo.Context, repo *firestore.PacienteRepository) error {
	pacientes, err := repo.GetAll(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error al obtener pacientes"})
	}
	return c.JSON(http.StatusOK, pacientes)
}

// CreatePaciente crea un nuevo paciente mediante el repositorio.
func CreatePaciente(c echo.Context, repo *firestore.PacienteRepository) error {
	paciente := new(models.Paciente)
	if err := c.Bind(paciente); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	// Validación mínima: nombre requerido
	if paciente.Nombre == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "El campo 'nombre' es requerido"})
	}
	pacienteCreado, err := repo.Create(c.Request().Context(), *paciente)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error al crear paciente"})
	}
	return c.JSON(http.StatusCreated, pacienteCreado)
}

// SetupRoutes configura las rutas de pacientes con inyección de dependencia.
func SetupRoutes(e *echo.Echo, repo *firestore.PacienteRepository) {
	e.GET("/pacientes", func(c echo.Context) error {
		return GetPacientes(c, repo)
	})
	e.POST("/pacientes", func(c echo.Context) error {
		return CreatePaciente(c, repo)
	})
}