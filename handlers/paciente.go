package handlers

import (
	"net/http"
	"nutricontrol/models"

	"github.com/labstack/echo/v4"
)

func GetPacientes(c echo.Context) error {
	pacientes := []models.Paciente{
		{Nombre: "Juan Perez", Edad: 30, Peso: 75.5, PorcentajeGrasa: 18.5, Notas: "Primera consulta"},
		{Nombre: "Maria Gomez", Edad: 25, Peso: 62.0, PorcentajeGrasa: 22.0, Notas: "Seguimiento mensual"},
	}
	return c.JSON(http.StatusOK, pacientes)
}

func CreatePaciente(c echo.Context) error {
	paciente := new(models.Paciente)
	if err := c.Bind(paciente); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	// Mock: retornar el mismo paciente simulando la creación exitosa
	return c.JSON(http.StatusCreated, paciente)
}
