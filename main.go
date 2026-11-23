package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"nutricontrol/handlers"
	"nutricontrol/pkg/firestore"
)

const (
	defaultAddr       = ":8080"
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 15 * time.Second
	idleTimeout       = 60 * time.Second
	shutdownTimeout   = 10 * time.Second
)

// resolveAddr normaliza el valor de la variable de entorno PORT al formato ":<puerto>".
func resolveAddr(portEnv string) string {
	port := strings.TrimSpace(portEnv)
	if port == "" {
		return defaultAddr
	}
	if strings.HasPrefix(port, ":") {
		return port
	}
	return ":" + port
}

// newEcho construye la instancia de Echo con middlewares operativos y rutas registradas.
func newEcho(repo firestore.PacienteRepository) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	handlers.SetupRoutes(e, repo)

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "API NutriControl Operativa")
	})

	return e
}

// newServer configura un *http.Server con timeouts explícitos sobre el router Echo.
func newServer(addr string, repo firestore.PacienteRepository) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           newEcho(repo),
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
}

// serve arranca srv sobre ln y ejecuta un apagado ordenado (graceful shutdown) al cancelarse ctx.
func serve(ctx context.Context, srv *http.Server, ln net.Listener, timeout time.Duration) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(ln)
	}()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		shutdownErr := srv.Shutdown(shutdownCtx)
		serveErr := <-errCh
		if shutdownErr != nil {
			return shutdownErr
		}
		if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			return serveErr
		}
		return nil
	}
}

// runServer abre el listener TCP en srv.Addr y delega el ciclo de vida a serve.
func runServer(ctx context.Context, srv *http.Server, timeout time.Duration) error {
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return err
	}
	return serve(ctx, srv, ln, timeout)
}

func main() {
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

	addr := resolveAddr(os.Getenv("PORT"))
	srv := newServer(addr, repo)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("Iniciando servidor en %s", addr)
	if err := runServer(ctx, srv, shutdownTimeout); err != nil {
		log.Fatalf("Error en servidor HTTP: %v", err)
	}
	log.Println("Servidor detenido correctamente")
}
