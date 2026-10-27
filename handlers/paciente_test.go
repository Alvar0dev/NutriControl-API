package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/labstack/echo/v4"

	"nutricontrol/models"
	"nutricontrol/pkg/firestore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setParam establece un parámetro de ruta en el contexto de Echo sin pasar por el router.
// Útil para probar handlers que usan c.Param directamente (Update, Delete).
// Echo v4 almacena los parámetros en pnames []string y pvalues []string (campos no exportados).
func setParam(c echo.Context, key, value string) {
	v := reflect.ValueOf(c).Elem()
	pnames := v.FieldByName("pnames")
	pvalues := v.FieldByName("pvalues")
	if pnames.IsValid() && pvalues.IsValid() {
		reflect.NewAt(pnames.Type(), unsafe.Pointer(pnames.UnsafeAddr())).Elem().Set(reflect.ValueOf([]string{key}))
		reflect.NewAt(pvalues.Type(), unsafe.Pointer(pvalues.UnsafeAddr())).Elem().Set(reflect.ValueOf([]string{value}))
	}
}

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
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &pacientes)
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
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &creado)
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
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp["error"], "nombre")
}

func TestCreatePacienteEdadInvalida(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	payload := `{"nombre":"Ana Lopez","edad":0,"peso":68.5,"porcentaje_grasa":24.0,"notas":"Nueva consulta"}`
	req := httptest.NewRequest(http.MethodPost, "/pacientes", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())

	// Act
	err := CreatePaciente(c, repo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, c.Response().Status)

	var resp map[string]string
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp["error"], "edad")
}

func TestCreatePacientePesoInvalido(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	payload := `{"nombre":"Ana Lopez","edad":28,"peso":0,"porcentaje_grasa":24.0,"notas":"Nueva consulta"}`
	req := httptest.NewRequest(http.MethodPost, "/pacientes", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())

	// Act
	err := CreatePaciente(c, repo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, c.Response().Status)

	var resp map[string]string
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp["error"], "peso")
}

func TestUpdatePacienteEdadInvalida(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	payload := `{"nombre":"Ana Lopez","edad":-5,"peso":68.5,"porcentaje_grasa":24.0,"notas":"Nueva consulta"}`
	req := httptest.NewRequest(http.MethodPut, "/pacientes/1", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())
	setParam(c, "id", "1")

	// Act
	err := UpdatePaciente(c, repo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, c.Response().Status)

	var resp map[string]string
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp["error"], "edad")
}

func TestUpdatePacientePesoInvalido(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	payload := `{"nombre":"Ana Lopez","edad":28,"peso":-2,"porcentaje_grasa":24.0,"notas":"Nueva consulta"}`
	req := httptest.NewRequest(http.MethodPut, "/pacientes/1", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())
	setParam(c, "id", "1")

	// Act
	err := UpdatePaciente(c, repo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, c.Response().Status)

	var resp map[string]string
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp["error"], "peso")
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

func TestGetPacienteByIDOK(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	creado, err := repo.Create(context.Background(), models.Paciente{Nombre: "Ana Lopez", Edad: 28, Peso: 68.5})
	require.NoError(t, err)
	require.NotEmpty(t, creado.ID)

	req := httptest.NewRequest(http.MethodGet, "/pacientes/"+creado.ID, nil)
	c := e.NewContext(req, httptest.NewRecorder())
	setParam(c, "id", creado.ID)

	// Act
	err = GetPacienteByID(c, repo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	var p models.Paciente
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &p)
	require.NoError(t, err)
	assert.Equal(t, creado.ID, p.ID)
	assert.Equal(t, "Ana Lopez", p.Nombre)
	assert.Equal(t, 28, p.Edad)
	assert.Equal(t, 68.5, p.Peso)
}

func TestGetPacienteByIDNoEncontrado(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()

	req := httptest.NewRequest(http.MethodGet, "/pacientes/inexistente", nil)
	c := e.NewContext(req, httptest.NewRecorder())
	setParam(c, "id", "inexistente")

	// Act
	err := GetPacienteByID(c, repo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, c.Response().Status)

	var resp map[string]string
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp["error"], "No se encontró paciente")
}