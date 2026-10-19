package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"

	"nutricontrol/handlers"
	"nutricontrol/pkg/firestore"
)

func main() {
	e := echo.New()

	var repo firestore.PacienteRepository
	var err error

	// Determinar si usar repositorio real (Firestore) o mock
	// Para desarrollo, se puede usar la variable de entorno NUTRICONTROL_USE_MOCK
	// Para producción, se requiere configuración de GCP con credenciales válidas
	useMock := os.Getenv("NUTRICONTROL_USE_MOCK") == "true"

	if !useMock {
		// Intentar crear repositorio real con Firestore
		// Esto requiere:
		// 1. Proyecto GCP configurado
		// 2. Credenciales de servicio (GOOGLE_APPLICATION_CREDENTIALS)
		// 3. Dependencia cloud.google.com/go/firestore
		projectID := os.Getenv("GCP_PROJECT_ID")
		if projectID != "" {
			repo, err = firestore.NewPacienteFirestoreRepo(context.Background(), projectID)
			if err == nil {
				log.Println("Usando repositorio Firestore real")
			} else {
				log.Printf("Error al inicializar Firestore: %v. Cambiando a mock.", err)
				useMock = true
			}
		} else {
			log.Println("GCP_PROJECT_ID no está definido. Cambiando a mock.")
			useMock = true
		}
	}

	if useMock {
		repo = firestore.NewMockPacienteRepository()
		log.Println("Usando repositorio mock (para desarrollo y testing)")
	}

	// Configurar rutas con inyección de dependencia
	handlers.SetupRoutes(e, repo)

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "API NutriControl Operativa")
	})

	// Información del servidor
	port := ":8080"
	log.Printf("Iniciando servidor en %s", port)
	e.Logger.Fatal(e.Start(port))
}