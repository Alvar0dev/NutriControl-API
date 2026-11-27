package main

import (
	stdbytes "bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/bytes"
	"golang.org/x/time/rate"

	"nutricontrol/handlers"
	"nutricontrol/pkg/firestore"
)

const (
	defaultAddr         = ":8080"
	defaultBodyLimit    = "1M"
	defaultRateLimitTTL = 3 * time.Minute
	readHeaderTimeout   = 5 * time.Second
	readTimeout         = 15 * time.Second
	writeTimeout        = 15 * time.Second
	idleTimeout         = 60 * time.Second
	shutdownTimeout     = 10 * time.Second
)

type serverConfig struct {
	Addr               string
	BodyLimit          string
	CORSAllowedOrigins []string
	RateLimitRPS       float64
	RateLimitBurst     int
}

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

// resolveBodyLimit valida BODY_LIMIT con la gramática de Echo/gommon para evitar panics en arranque,
// recurriendo al límite por defecto cuando el valor falta, es <= 0 o tiene formato inválido.
func resolveBodyLimit(limitEnv string) string {
	limit := strings.TrimSpace(limitEnv)
	if limit == "" {
		return defaultBodyLimit
	}
	parsed, err := bytes.Parse(limit)
	if err != nil || parsed <= 0 {
		log.Printf("BODY_LIMIT inválido (%q), usando valor por defecto %s", limitEnv, defaultBodyLimit)
		return defaultBodyLimit
	}
	return limit
}

// parseCORSOrigins extrae los orígenes permitidos separados por coma; devuelve nil si no hay ninguno.
func parseCORSOrigins(originsEnv string) []string {
	parts := strings.Split(originsEnv, ",")
	var origins []string
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}

// resolveRateLimit valida RATE_LIMIT_RPS y RATE_LIMIT_BURST de forma segura sin provocar panics;
// si RATE_LIMIT_RPS falta o es <= 0/inválido devuelve (0, 0) dejando el rate limiting desactivado.
func resolveRateLimit(rpsEnv, burstEnv string) (float64, int) {
	rawRPS := strings.TrimSpace(rpsEnv)
	if rawRPS == "" {
		return 0, 0
	}
	rps, err := strconv.ParseFloat(rawRPS, 64)
	if err != nil || math.IsNaN(rps) || math.IsInf(rps, 0) || rps <= 0 {
		log.Printf("RATE_LIMIT_RPS inválido (%q), rate limiting desactivado", rpsEnv)
		return 0, 0
	}

	defaultBurst := int(math.Max(1, math.Ceil(rps)))
	rawBurst := strings.TrimSpace(burstEnv)
	if rawBurst == "" {
		return rps, defaultBurst
	}
	parsedBurst, err := strconv.Atoi(rawBurst)
	if err != nil || parsedBurst <= 0 {
		log.Printf("RATE_LIMIT_BURST inválido (%q), usando ráfaga por defecto %d", burstEnv, defaultBurst)
		return rps, defaultBurst
	}
	return rps, parsedBurst
}

func defaultServerConfig(addr string) serverConfig {
	return serverConfig{
		Addr:      resolveAddr(addr),
		BodyLimit: defaultBodyLimit,
	}
}

func loadServerConfig() serverConfig {
	rps, burst := resolveRateLimit(os.Getenv("RATE_LIMIT_RPS"), os.Getenv("RATE_LIMIT_BURST"))
	return serverConfig{
		Addr:               resolveAddr(os.Getenv("PORT")),
		BodyLimit:          resolveBodyLimit(os.Getenv("BODY_LIMIT")),
		CORSAllowedOrigins: parseCORSOrigins(os.Getenv("CORS_ALLOWED_ORIGINS")),
		RateLimitRPS:       rps,
		RateLimitBurst:     burst,
	}
}

type strictJSONSerializer struct {
	echo.DefaultJSONSerializer
}

// Deserialize lee el cuerpo completo antes de parsear JSON para evitar que json.Decoder (Go <=1.26)
// descarte echo.ErrStatusRequestEntityTooLarge cuando un valor JSON cierra dentro del mismo Read.
func (strictJSONSerializer) Deserialize(c echo.Context, i any) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}
	if len(stdbytes.TrimSpace(body)) == 0 {
		return io.EOF
	}
	return json.Unmarshal(body, i)
}

// newEchoWithConfig construye Echo con middlewares operativos, protección de entrada (CORS, RateLimit opcionales, BodyLimit) y rutas.
func newEchoWithConfig(repo firestore.PacienteRepository, cfg serverConfig) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.JSONSerializer = strictJSONSerializer{}

	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	if len(cfg.CORSAllowedOrigins) > 0 {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     cfg.CORSAllowedOrigins,
			AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
			AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, echo.HeaderXRequestID},
			AllowCredentials: false,
		}))
	}
	if cfg.RateLimitRPS > 0 && !math.IsNaN(cfg.RateLimitRPS) && !math.IsInf(cfg.RateLimitRPS, 0) {
		burst := cfg.RateLimitBurst
		if burst <= 0 {
			burst = int(math.Max(1, math.Ceil(cfg.RateLimitRPS)))
		}
		limiterStore := middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
			Rate:      rate.Limit(cfg.RateLimitRPS),
			Burst:     burst,
			ExpiresIn: defaultRateLimitTTL,
		})
		e.Use(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
			Skipper: func(c echo.Context) bool {
				return c.Request().Method == http.MethodOptions || c.Path() == "/health"
			},
			Store: limiterStore,
			DenyHandler: func(c echo.Context, identifier string, err error) error {
				return echo.NewHTTPError(http.StatusTooManyRequests, map[string]string{
					"error": "Demasiadas peticiones, intente nuevamente más tarde",
				})
			},
		}))
	}
	e.Use(middleware.BodyLimit(resolveBodyLimit(cfg.BodyLimit)))

	handlers.SetupRoutes(e, repo)

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "API NutriControl Operativa")
	})

	return e
}

// newEcho construye la instancia de Echo con la configuración segura por defecto.
func newEcho(repo firestore.PacienteRepository) *echo.Echo {
	return newEchoWithConfig(repo, defaultServerConfig(""))
}

// newServerWithConfig configura un *http.Server con timeouts explícitos y la configuración HTTP dada.
func newServerWithConfig(cfg serverConfig, repo firestore.PacienteRepository) *http.Server {
	return &http.Server{
		Addr:              cfg.Addr,
		Handler:           newEchoWithConfig(repo, cfg),
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
}

// newServer configura un *http.Server con timeouts explícitos sobre el router Echo.
func newServer(addr string, repo firestore.PacienteRepository) *http.Server {
	return newServerWithConfig(defaultServerConfig(addr), repo)
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

	cfg := loadServerConfig()
	srv := newServerWithConfig(cfg, repo)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("Iniciando servidor en %s (body_limit=%s, cors_origins=%d, rate_limit_rps=%g, rate_limit_burst=%d)", cfg.Addr, cfg.BodyLimit, len(cfg.CORSAllowedOrigins), cfg.RateLimitRPS, cfg.RateLimitBurst)
	if err := runServer(ctx, srv, shutdownTimeout); err != nil {
		log.Fatalf("Error en servidor HTTP: %v", err)
	}
	log.Println("Servidor detenido correctamente")
}
