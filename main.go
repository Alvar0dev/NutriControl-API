package main

import (
	"net/http"
	"nutricontrol/handlers"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "API NutriControl Operativa")
	})

	e.GET("/pacientes", handlers.GetPacientes)
	e.POST("/pacientes", handlers.CreatePaciente)

	e.Logger.Fatal(e.Start(":8080"))
}
