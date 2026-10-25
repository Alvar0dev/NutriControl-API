package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"nutricontrol/models"
	"nutricontrol/pkg/firestore"
)

// GetPacientes obtiene todos los pacientes mediante el repositorio.
func GetPacientes(c echo.Context, repo firestore.PacienteRepository) error {
	pacientes, err := repo.GetAll(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error al obtener pacientes"})
	}
	return c.JSON(http.StatusOK, pacientes)
}

// validarPaciente verifica las reglas de negocio mínimas: nombre, edad y peso > 0.
func validarPaciente(p models.Paciente) error {
	if p.Nombre == "" {
		return fmt.Errorf("El campo 'nombre' es requerido")
	}
	if p.Edad <= 0 {
		return fmt.Errorf("El campo 'edad' debe ser mayor que 0")
	}
	if p.Peso <= 0 {
		return fmt.Errorf("El campo 'peso' debe ser mayor que 0")
	}
	return nil
}

// CreatePaciente crea un nuevo paciente mediante el repositorio.
func CreatePaciente(c echo.Context, repo firestore.PacienteRepository) error {
	paciente := new(models.Paciente)
	if err := c.Bind(paciente); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	if err := validarPaciente(*paciente); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	pacienteCreado, err := repo.Create(c.Request().Context(), *paciente)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error al crear paciente"})
	}
	return c.JSON(http.StatusCreated, pacienteCreado)
}

// UpdatePaciente actualiza un paciente existente por ID.
func UpdatePaciente(c echo.Context, repo firestore.PacienteRepository) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "El campo 'id' es requerido"})
	}
	p := new(models.Paciente)
	if err := c.Bind(p); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	if err := validarPaciente(*p); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	p.ID = id
	pacienteActualizado, err := repo.Update(c.Request().Context(), id, *p)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "No se encontró paciente con ID " + id})
	}
	return c.JSON(http.StatusOK, pacienteActualizado)
}

// GetPacienteByID obtiene un paciente por su ID.
func GetPacienteByID(c echo.Context, repo firestore.PacienteRepository) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "El campo 'id' es requerido"})
	}
	paciente, err := repo.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "No se encontró paciente con ID " + id})
	}
	return c.JSON(http.StatusOK, paciente)
}

// DeletePaciente elimina un paciente por su ID.
func DeletePaciente(c echo.Context, repo firestore.PacienteRepository) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "El campo 'id' es requerido"})
	}
	if err := repo.Delete(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "No se encontró paciente con ID " + id})
	}
	return c.NoContent(http.StatusNoContent)
}

// SetupRoutes configura las rutas de pacientes con inyección de dependencia.
func SetupRoutes(e *echo.Echo, repo firestore.PacienteRepository) {
	e.GET("/pacientes", func(c echo.Context) error {
		return GetPacientes(c, repo)
	})
	e.POST("/pacientes", func(c echo.Context) error {
		return CreatePaciente(c, repo)
	})
	e.GET("/pacientes/{id}", func(c echo.Context) error {
		return GetPacienteByID(c, repo)
	})
	e.PUT("/pacientes/{id}", func(c echo.Context) error {
		return UpdatePaciente(c, repo)
	})
	e.DELETE("/pacientes/{id}", func(c echo.Context) error {
		return DeletePaciente(c, repo)
	})
}
