package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	"nutricontrol/models"
	"nutricontrol/pkg/firestore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestEcho crea una instancia de Echo para tests.
func newTestEcho() *echo.Echo {
	return echo.New()
}

func TestGetPacientesOK(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	req := httptest.NewRequest(http.MethodGet, "/pacientes", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())

	// Act
	err := GetPacientes(c, repo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	var pacientes []models.Paciente
	err = json.Unmarshal(c.Response().Body.Bytes(), &pacientes)
	require.NoError(t, err)
	assert.Len(t, pacientes, 2)
	assert.Equal(t, "Juan Perez", pacientes[0].Nombre)
	assert.Equal(t, "Maria Gomez", pacientes[1].Nombre)
}

func TestCreatePacienteOK(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	payload := `{"nombre":"Ana Lopez","edad":28,"peso":68.5,"porcentaje_grasa":24.0,"notas":"Nueva consulta"}`
	req := httptest.NewRequest(http.MethodPost, "/pacientes", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())

	// Act
	err := CreatePaciente(c, repo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, c.Response().Status)

	var creado models.Paciente
	err = json.Unmarshal(c.Response().Body.Bytes(), &creado)
	require.NoError(t, err)
	assert.Equal(t, "Ana Lopez", creado.Nombre)
	assert.Equal(t, 28, creado.Edad)
	assert.Equal(t, 68.5, creado.Peso)
	assert.Equal(t, 24.0, creado.PorcentajeGrasa)
	assert.Equal(t, "Nueva consulta", creado.Notas)
}

func TestCreatePacienteSinNombre(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	payload := `{"edad":28,"peso":68.5}`
	req := httptest.NewRequest(http.MethodPost, "/pacientes", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())

	// Act
	err := CreatePaciente(c, repo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, c.Response().Status)

	var resp map[string]string
	err = json.Unmarshal(c.Response().Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp["error"], "nombre")
}

func TestSetupRoutes(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()

	// Act
	SetupRoutes(e, repo)

	// Assert: las rutas deben estar registradas (no fallan al montar)
	assert.NotNil(t, e.Routes())
}