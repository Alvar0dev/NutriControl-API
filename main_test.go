package main

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nutricontrol/models"
	"nutricontrol/pkg/firestore"
)

func TestResolveAddr(t *testing.T) {
	casos := []struct {
		name     string
		portEnv  string
		expected string
	}{
		{name: "vacio usa 8080 por defecto", portEnv: "", expected: ":8080"},
		{name: "espacios usa 8080 por defecto", portEnv: "   ", expected: ":8080"},
		{name: "puerto numerico agrega dos puntos", portEnv: "9090", expected: ":9090"},
		{name: "puerto con dos puntos se preserva", portEnv: ":3000", expected: ":3000"},
		{name: "recorta espacios alrededor", portEnv: " 8081 ", expected: ":8081"},
	}

	for _, tc := range casos {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, resolveAddr(tc.portEnv))
		})
	}
}

func TestNewEchoMiddlewaresYRoutes(t *testing.T) {
	repo := firestore.NewMockPacienteRepository()
	e := newEcho(repo)

	t.Run("ruta raiz responde 200 y emite X-Request-ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "API NutriControl Operativa", rec.Body.String())
		assert.NotEmpty(t, rec.Header().Get(echo.HeaderXRequestID))
	})

	t.Run("ruta health registrada y preserva X-Request-ID entrante", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		req.Header.Set(echo.HeaderXRequestID, "req-test-123")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "req-test-123", rec.Header().Get(echo.HeaderXRequestID))
		assert.Contains(t, rec.Body.String(), `"storage":"mock"`)
	})

	t.Run("middleware Recover intercepta panic y devuelve 500", func(t *testing.T) {
		e.GET("/panic-test", func(c echo.Context) error {
			panic("fallo simulado")
		})
		req := httptest.NewRequest(http.MethodGet, "/panic-test", nil)
		rec := httptest.NewRecorder()

		assert.NotPanics(t, func() {
			e.ServeHTTP(rec, req)
		})
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestNewServerTimeouts(t *testing.T) {
	repo := firestore.NewMockPacienteRepository()
	srv := newServer(":9090", repo)

	assert.Equal(t, ":9090", srv.Addr)
	assert.NotNil(t, srv.Handler)
	assert.Equal(t, 5*time.Second, srv.ReadHeaderTimeout)
	assert.Equal(t, 15*time.Second, srv.ReadTimeout)
	assert.Equal(t, 15*time.Second, srv.WriteTimeout)
	assert.Equal(t, 60*time.Second, srv.IdleTimeout)
}

func TestServeGracefulShutdown(t *testing.T) {
	repo := firestore.NewMockPacienteRepository()
	srv := newServer("127.0.0.1:0", repo)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- serve(ctx, srv, ln, 2*time.Second)
	}()

	// Verificar que el servidor atiende peticiones HTTP reales
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://" + ln.Addr().String() + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, resp.Header.Get(echo.HeaderXRequestID))

	// Cancelar contexto y comprobar apagado ordenado sin error
	cancel()

	select {
	case serveErr := <-errCh:
		assert.NoError(t, serveErr)
	case <-time.After(3 * time.Second):
		t.Fatal("el servidor no completó el graceful shutdown dentro del tiempo esperado")
	}
}

func TestResolveBodyLimit(t *testing.T) {
	casos := []struct {
		name     string
		envVal   string
		expected string
	}{
		{name: "vacio usa 1M por defecto", envVal: "", expected: "1M"},
		{name: "espacios usa 1M por defecto", envVal: "   ", expected: "1M"},
		{name: "valor valido en KB", envVal: " 512K ", expected: "512K"},
		{name: "valor valido en MB", envVal: "2M", expected: "2M"},
		{name: "cero vuelve al defecto seguro", envVal: "0", expected: "1M"},
		{name: "negativo vuelve al defecto seguro", envVal: "-1M", expected: "1M"},
		{name: "valor invalido vuelve al defecto seguro", envVal: "no-valido", expected: "1M"},
	}

	for _, tc := range casos {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, resolveBodyLimit(tc.envVal))
		})
	}
}

func TestParseCORSOrigins(t *testing.T) {
	casos := []struct {
		name     string
		envVal   string
		expected []string
	}{
		{name: "vacio devuelve nil (CORS desactivado)", envVal: "", expected: nil},
		{name: "solo espacios o comas devuelve nil", envVal: " , , ", expected: nil},
		{name: "un origen se recorta", envVal: " http://localhost:3000 ", expected: []string{"http://localhost:3000"}},
		{name: "multiples origenes separados por coma", envVal: "http://localhost:3000, https://app.nutricontrol.com ,", expected: []string{"http://localhost:3000", "https://app.nutricontrol.com"}},
	}

	for _, tc := range casos {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, parseCORSOrigins(tc.envVal))
		})
	}
}

func TestLoadServerConfig(t *testing.T) {
	t.Setenv("PORT", "9091")
	t.Setenv("BODY_LIMIT", "2M")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173, https://admin.example.com")
	t.Setenv("RATE_LIMIT_RPS", "15")
	t.Setenv("RATE_LIMIT_BURST", "30")

	cfg := loadServerConfig()
	assert.Equal(t, ":9091", cfg.Addr)
	assert.Equal(t, "2M", cfg.BodyLimit)
	assert.Equal(t, []string{"http://localhost:5173", "https://admin.example.com"}, cfg.CORSAllowedOrigins)
	assert.Equal(t, 15.0, cfg.RateLimitRPS)
	assert.Equal(t, 30, cfg.RateLimitBurst)
}

func TestResolveRateLimit(t *testing.T) {
	casos := []struct {
		name          string
		rpsEnv        string
		burstEnv      string
		expectedRPS   float64
		expectedBurst int
	}{
		{name: "vacio desactiva rate limiting", rpsEnv: "", burstEnv: "", expectedRPS: 0, expectedBurst: 0},
		{name: "solo espacios desactiva rate limiting", rpsEnv: "   ", burstEnv: "10", expectedRPS: 0, expectedBurst: 0},
		{name: "cero desactiva rate limiting", rpsEnv: "0", burstEnv: "10", expectedRPS: 0, expectedBurst: 0},
		{name: "negativo desactiva rate limiting", rpsEnv: "-5", burstEnv: "10", expectedRPS: 0, expectedBurst: 0},
		{name: "texto invalido desactiva rate limiting", rpsEnv: "abc", burstEnv: "10", expectedRPS: 0, expectedBurst: 0},
		{name: "NaN e Inf desactivan rate limiting", rpsEnv: "NaN", burstEnv: "10", expectedRPS: 0, expectedBurst: 0},
		{name: "RPS entero sin burst usa ceil(RPS)", rpsEnv: "10", burstEnv: "", expectedRPS: 10.0, expectedBurst: 10},
		{name: "RPS decimal sin burst redondea hacia arriba", rpsEnv: "1.5", burstEnv: "", expectedRPS: 1.5, expectedBurst: 2},
		{name: "RPS menor a 1 sin burst fija burst minimo 1", rpsEnv: "0.5", burstEnv: "", expectedRPS: 0.5, expectedBurst: 1},
		{name: "RPS y burst explicitos con espacios", rpsEnv: " 20 ", burstEnv: " 50 ", expectedRPS: 20.0, expectedBurst: 50},
		{name: "burst invalido o <= 0 recurre al burst por defecto", rpsEnv: "10", burstEnv: "no-int", expectedRPS: 10.0, expectedBurst: 10},
		{name: "burst cero recurre al burst por defecto", rpsEnv: "10", burstEnv: "0", expectedRPS: 10.0, expectedBurst: 10},
	}

	for _, tc := range casos {
		t.Run(tc.name, func(t *testing.T) {
			rps, burst := resolveRateLimit(tc.rpsEnv, tc.burstEnv)
			assert.Equal(t, tc.expectedRPS, rps)
			assert.Equal(t, tc.expectedBurst, burst)
		})
	}
}

func TestNewEchoBodyLimit(t *testing.T) {
	repo := firestore.NewMockPacienteRepository()
	e := newEchoWithConfig(repo, serverConfig{
		Addr:      ":8080",
		BodyLimit: "100B",
	})

	t.Run("payload dentro del limite responde 201", func(t *testing.T) {
		payload := `{"nombre":"Ana","edad":28,"peso":60,"porcentaje_grasa":20}`
		req := httptest.NewRequest(http.MethodPost, "/pacientes", strings.NewReader(payload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("payload que supera Content-Length responde 413", func(t *testing.T) {
		payload := `{"nombre":"Ana","edad":28,"peso":60,"porcentaje_grasa":20,"notas":"` + strings.Repeat("x", 150) + `"}`
		req := httptest.NewRequest(http.MethodPost, "/pacientes", strings.NewReader(payload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
	})

	t.Run("payload streaming sin Content-Length que supera limite responde 413 en todos los endpoints de escritura", func(t *testing.T) {
		creado, err := repo.Create(context.Background(), models.Paciente{Nombre: "Stream Test", Edad: 30, Peso: 70})
		require.NoError(t, err)
		ctrl, err := repo.AddControl(context.Background(), creado.ID, models.ControlClinico{
			Fecha:           time.Now().UTC(),
			Peso:            70,
			PorcentajeGrasa: 20,
		})
		require.NoError(t, err)

		relleno := strings.Repeat("y", 150)
		casos := []struct {
			name    string
			method  string
			target  string
			payload string
		}{
			{
				name:    "POST /pacientes",
				method:  http.MethodPost,
				target:  "/pacientes",
				payload: `{"nombre":"Ana","edad":28,"peso":60,"porcentaje_grasa":20,"notas":"` + relleno + `"}`,
			},
			{
				name:    "POST /pacientes/batch",
				method:  http.MethodPost,
				target:  "/pacientes/batch",
				payload: `[{"nombre":"Ana","edad":28,"peso":60,"porcentaje_grasa":20,"notas":"` + relleno + `"}]`,
			},
			{
				name:    "PUT /pacientes/:id",
				method:  http.MethodPut,
				target:  "/pacientes/" + creado.ID,
				payload: `{"nombre":"Ana","edad":28,"peso":60,"porcentaje_grasa":20,"notas":"` + relleno + `"}`,
			},
			{
				name:    "PATCH /pacientes/:id",
				method:  http.MethodPatch,
				target:  "/pacientes/" + creado.ID,
				payload: `{"notas":"` + relleno + `"}`,
			},
			{
				name:    "POST /pacientes/:id/controles",
				method:  http.MethodPost,
				target:  "/pacientes/" + creado.ID + "/controles",
				payload: `{"peso":70,"porcentaje_grasa":20,"notas":"` + relleno + `"}`,
			},
			{
				name:    "PUT /pacientes/:id/controles/:controlId",
				method:  http.MethodPut,
				target:  "/pacientes/" + creado.ID + "/controles/" + ctrl.ID,
				payload: `{"peso":70,"porcentaje_grasa":20,"notas":"` + relleno + `"}`,
			},
			{
				name:    "PATCH /pacientes/:id/controles/:controlId",
				method:  http.MethodPatch,
				target:  "/pacientes/" + creado.ID + "/controles/" + ctrl.ID,
				payload: `{"notas":"` + relleno + `"}`,
			},
		}

		for _, tc := range casos {
			t.Run(tc.name, func(t *testing.T) {
				req := httptest.NewRequest(tc.method, tc.target, strings.NewReader(tc.payload+"\n"))
				req.ContentLength = -1
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				rec := httptest.NewRecorder()

				e.ServeHTTP(rec, req)
				assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
			})
		}
	})
}

func TestNewEchoCORS(t *testing.T) {
	t.Run("sin CORS_ALLOWED_ORIGINS no emite cabeceras CORS", func(t *testing.T) {
		repo := firestore.NewMockPacienteRepository()
		e := newEcho(repo)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		req.Header.Set(echo.HeaderOrigin, "https://evil.example.com")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Empty(t, rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
	})

	t.Run("con origen permitido responde preflight OPTIONS y cabecera CORS", func(t *testing.T) {
		repo := firestore.NewMockPacienteRepository()
		e := newEchoWithConfig(repo, serverConfig{
			Addr:               ":8080",
			BodyLimit:          defaultBodyLimit,
			CORSAllowedOrigins: []string{"http://localhost:3000"},
		})

		reqPre := httptest.NewRequest(http.MethodOptions, "/pacientes", nil)
		reqPre.Header.Set(echo.HeaderOrigin, "http://localhost:3000")
		reqPre.Header.Set(echo.HeaderAccessControlRequestMethod, http.MethodPost)
		recPre := httptest.NewRecorder()

		e.ServeHTTP(recPre, reqPre)
		assert.Equal(t, http.StatusNoContent, recPre.Code)
		assert.Equal(t, "http://localhost:3000", recPre.Header().Get(echo.HeaderAccessControlAllowOrigin))
		assert.Empty(t, recPre.Header().Get(echo.HeaderAccessControlAllowCredentials))

		// Origen no autorizado no recibe Access-Control-Allow-Origin
		reqDisallowed := httptest.NewRequest(http.MethodGet, "/health", nil)
		reqDisallowed.Header.Set(echo.HeaderOrigin, "https://otro.example.com")
		recDisallowed := httptest.NewRecorder()

		e.ServeHTTP(recDisallowed, reqDisallowed)
		assert.Equal(t, http.StatusOK, recDisallowed.Code)
		assert.Empty(t, recDisallowed.Header().Get(echo.HeaderAccessControlAllowOrigin))
	})
}

type partialLimitReader struct {
	data []byte
	done bool
}

func (r *partialLimitReader) Read(p []byte) (int, error) {
	if r.done {
		return 0, echo.ErrStatusRequestEntityTooLarge
	}
	n := copy(p, r.data)
	r.done = true
	return n, echo.ErrStatusRequestEntityTooLarge
}

func (r *partialLimitReader) Close() error { return nil }

func TestStrictJSONSerializerDeserializePropaga413(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/pacientes", nil)
	req.Body = &partialLimitReader{data: []byte(`{"nombre":"Ana"}` + "\n")}
	req.ContentLength = -1
	c := e.NewContext(req, httptest.NewRecorder())

	var p models.Paciente
	err := (strictJSONSerializer{}).Deserialize(c, &p)
	assert.True(t, errors.Is(err, echo.ErrStatusRequestEntityTooLarge))
	assert.Empty(t, p.Nombre)
}

func TestNewEchoRateLimit(t *testing.T) {
	t.Run("sin RATE_LIMIT_RPS no limita peticiones", func(t *testing.T) {
		repo := firestore.NewMockPacienteRepository()
		e := newEcho(repo)

		for i := 0; i < 10; i++ {
			req := httptest.NewRequest(http.MethodGet, "/pacientes", nil)
			req.RemoteAddr = "203.0.113.10:1234"
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})

	t.Run("con limite activo permite rafaga, devuelve 429 JSON, aisla por IP, omite health/OPTIONS y conserva CORS", func(t *testing.T) {
		repo := firestore.NewMockPacienteRepository()
		e := newEchoWithConfig(repo, serverConfig{
			Addr:               ":8080",
			BodyLimit:          defaultBodyLimit,
			CORSAllowedOrigins: []string{"http://localhost:3000"},
			RateLimitRPS:       0.001,
			RateLimitBurst:     2,
		})

		ip1 := "203.0.113.10:1234"
		ip2 := "203.0.113.20:5678"

		// Ráfaga inicial de 2 peticiones permitida para ip1
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest(http.MethodGet, "/pacientes", nil)
			req.RemoteAddr = ip1
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
		}

		// La tercera petición de ip1 excede la cuota → 429 JSON manteniendo cabecera CORS
		reqLimited := httptest.NewRequest(http.MethodGet, "/pacientes", nil)
		reqLimited.RemoteAddr = ip1
		reqLimited.Header.Set(echo.HeaderOrigin, "http://localhost:3000")
		recLimited := httptest.NewRecorder()
		e.ServeHTTP(recLimited, reqLimited)
		assert.Equal(t, http.StatusTooManyRequests, recLimited.Code)
		assert.Contains(t, recLimited.Body.String(), `"error"`)
		assert.Contains(t, recLimited.Body.String(), "Demasiadas peticiones")
		assert.Equal(t, "http://localhost:3000", recLimited.Header().Get(echo.HeaderAccessControlAllowOrigin))

		// Otra IP mantiene su propia cuota intacta
		reqIP2 := httptest.NewRequest(http.MethodGet, "/pacientes", nil)
		reqIP2.RemoteAddr = ip2
		recIP2 := httptest.NewRecorder()
		e.ServeHTTP(recIP2, reqIP2)
		assert.Equal(t, http.StatusOK, recIP2.Code)

		// /health y preflight OPTIONS siguen permitidos incluso para ip1 bloqueada
		for i := 0; i < 3; i++ {
			reqHealth := httptest.NewRequest(http.MethodGet, "/health", nil)
			reqHealth.RemoteAddr = ip1
			recHealth := httptest.NewRecorder()
			e.ServeHTTP(recHealth, reqHealth)
			assert.Equal(t, http.StatusOK, recHealth.Code)

			reqOpt := httptest.NewRequest(http.MethodOptions, "/pacientes", nil)
			reqOpt.RemoteAddr = ip1
			reqOpt.Header.Set(echo.HeaderOrigin, "http://localhost:3000")
			reqOpt.Header.Set(echo.HeaderAccessControlRequestMethod, http.MethodPost)
			recOpt := httptest.NewRecorder()
			e.ServeHTTP(recOpt, reqOpt)
			assert.Equal(t, http.StatusNoContent, recOpt.Code)
		}
	})
}

var _ io.ReadCloser = (*partialLimitReader)(nil)
