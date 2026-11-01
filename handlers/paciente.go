package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"nutricontrol/models"
	"nutricontrol/pkg/firestore"
)

// GetPacientes obtiene todos los pacientes mediante el repositorio.
// Soporta filtro opcional por nombre (?nombre=, coincidencia parcial, sin distinción de mayúsculas).
func GetPacientes(c echo.Context, repo firestore.PacienteRepository) error {
	pacientes, err := repo.GetAll(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error al obtener pacientes"})
	}
	if nombre := c.QueryParam("nombre"); nombre != "" {
		buscado := strings.ToLower(nombre)
		filtrados := pacientes[:0]
		for _, p := range pacientes {
			if strings.Contains(strings.ToLower(p.Nombre), buscado) {
				filtrados = append(filtrados, p)
			}
		}
		pacientes = filtrados
	}
	return c.JSON(http.StatusOK, pacientes)
}

// SearchPacientes busca pacientes por múltiples criterios con paginación.
// Query params:
//   - nombre: coincidencia parcial, sin distinción de mayúsculas/minúsculas
//   - edad_min, edad_max: rango inclusivo de edad
//   - peso_min, peso_max: rango inclusivo de peso
//   - limit: máximo resultados (default 10)
//   - offset: salto de resultados (default 0)
func SearchPacientes(c echo.Context, repo firestore.PacienteRepository) error {
	pacientes, err := repo.GetAll(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error al obtener pacientes"})
	}

	// Filtro por nombre (coincidencia parcial, case-insensitive)
	if nombre := c.QueryParam("nombre"); nombre != "" {
		buscado := strings.ToLower(nombre)
		filtrados := pacientes[:0]
		for _, p := range pacientes {
			if strings.Contains(strings.ToLower(p.Nombre), buscado) {
				filtrados = append(filtrados, p)
			}
		}
		pacientes = filtrados
	}

	// Filtro por rango de edad
	if edadMinStr := c.QueryParam("edad_min"); edadMinStr != "" {
		if edadMin, err := strconv.Atoi(edadMinStr); err == nil {
			filtrados := pacientes[:0]
			for _, p := range pacientes {
				if p.Edad >= edadMin {
					filtrados = append(filtrados, p)
				}
			}
			pacientes = filtrados
		}
	}
	if edadMaxStr := c.QueryParam("edad_max"); edadMaxStr != "" {
		if edadMax, err := strconv.Atoi(edadMaxStr); err == nil {
			filtrados := pacientes[:0]
			for _, p := range pacientes {
				if p.Edad <= edadMax {
					filtrados = append(filtrados, p)
				}
			}
			pacientes = filtrados
		}
	}

	// Filtro por rango de peso
	if pesoMinStr := c.QueryParam("peso_min"); pesoMinStr != "" {
		if pesoMin, err := strconv.ParseFloat(pesoMinStr, 64); err == nil {
			filtrados := pacientes[:0]
			for _, p := range pacientes {
				if p.Peso >= pesoMin {
					filtrados = append(filtrados, p)
				}
			}
			pacientes = filtrados
		}
	}
	if pesoMaxStr := c.QueryParam("peso_max"); pesoMaxStr != "" {
		if pesoMax, err := strconv.ParseFloat(pesoMaxStr, 64); err == nil {
			filtrados := pacientes[:0]
			for _, p := range pacientes {
				if p.Peso <= pesoMax {
					filtrados = append(filtrados, p)
				}
			}
			pacientes = filtrados
		}
	}

	// Paginación: limit (default 10) y offset (default 0)
	limit := 10
	offset := 0
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Aplicar paginación
	if offset >= len(pacientes) {
		pacientes = []models.Paciente{}
	} else {
		end := offset + limit
		if end > len(pacientes) {
			end = len(pacientes)
		}
		pacientes = pacientes[offset:end]
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
	e.GET("/pacientes/search", func(c echo.Context) error {
		return SearchPacientes(c, repo)
	})
}
