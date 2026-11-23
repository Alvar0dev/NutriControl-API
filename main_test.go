package main

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
