package handlers

import (
	"context"
	"encoding/json"
	"fmt"
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

func TestGetPacientesFiltrarPorNombre(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	req := httptest.NewRequest(http.MethodGet, "/pacientes?nombre=juan", nil)
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
	assert.Len(t, pacientes, 1)
	assert.Equal(t, "Juan Perez", pacientes[0].Nombre)
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

func TestSearchPacientesSinFiltros(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	req := httptest.NewRequest(http.MethodGet, "/pacientes/search", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())

	// Act
	err := SearchPacientes(c, repo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	var pacientes []models.Paciente
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &pacientes)
	require.NoError(t, err)
	// Default limit=10, offset=0 -> todos los 2 pacientes seed
	assert.Len(t, pacientes, 2)
}

func TestSearchPacientesFiltrarPorNombre(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	req := httptest.NewRequest(http.MethodGet, "/pacientes/search?nombre=juan", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())

	// Act
	err := SearchPacientes(c, repo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	var pacientes []models.Paciente
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &pacientes)
	require.NoError(t, err)
	assert.Len(t, pacientes, 1)
	assert.Equal(t, "Juan Perez", pacientes[0].Nombre)
}

func TestSearchPacientesFiltrarPorEdadMinMax(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	// Juan: 30, Maria: 25
	req := httptest.NewRequest(http.MethodGet, "/pacientes/search?edad_min=28&edad_max=35", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())

	// Act
	err := SearchPacientes(c, repo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	var pacientes []models.Paciente
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &pacientes)
	require.NoError(t, err)
	assert.Len(t, pacientes, 1)
	assert.Equal(t, "Juan Perez", pacientes[0].Nombre)
}

func TestSearchPacientesFiltrarPorPesoMinMax(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	// Juan: 75.5, Maria: 62.0
	req := httptest.NewRequest(http.MethodGet, "/pacientes/search?peso_min=70&peso_max=80", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())

	// Act
	err := SearchPacientes(c, repo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	var pacientes []models.Paciente
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &pacientes)
	require.NoError(t, err)
	assert.Len(t, pacientes, 1)
	assert.Equal(t, "Juan Perez", pacientes[0].Nombre)
}

func TestSearchPacientesPaginacionLimitOffset(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	// Crear 3 pacientes más para tener 5 total
	for i := 0; i < 3; i++ {
		_, _ = repo.Create(context.Background(), models.Paciente{
			Nombre: fmt.Sprintf("Paciente Extra %d", i),
			Edad:  20 + i,
			Peso:  60.0 + float64(i),
		})
	}
	// limit=2, offset=1 -> debe devolver 2 pacientes (segundo y tercero)
	req := httptest.NewRequest(http.MethodGet, "/pacientes/search?limit=2&offset=1", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())

	// Act
	err := SearchPacientes(c, repo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	var pacientes []models.Paciente
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &pacientes)
	require.NoError(t, err)
	assert.Len(t, pacientes, 2)
}

func TestSearchPacientesValoresPorDefectoLimitOffset(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	// Crear 15 pacientes total (2 seed + 13 nuevos)
	for i := 0; i < 13; i++ {
		_, _ = repo.Create(context.Background(), models.Paciente{
			Nombre: fmt.Sprintf("Paciente %d", i),
			Edad:  20,
			Peso:  60.0,
		})
	}
	// Sin limit/offset -> default limit=10, offset=0 -> 10 pacientes
	req := httptest.NewRequest(http.MethodGet, "/pacientes/search", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())

	// Act
	err := SearchPacientes(c, repo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	var pacientes []models.Paciente
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &pacientes)
	require.NoError(t, err)
	assert.Len(t, pacientes, 10)
}

func TestSearchPacientesResultadoVacio(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	// Nombre que no existe
	req := httptest.NewRequest(http.MethodGet, "/pacientes/search?nombre=inexistente", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())

	// Act
	err := SearchPacientes(c, repo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	var pacientes []models.Paciente
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &pacientes)
	require.NoError(t, err)
	assert.Len(t, pacientes, 0)
}

func TestSearchPacientesCombinadoFiltros(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	// Juan: 30 años, 75.5kg
	// Maria: 25 años, 62.0kg
	req := httptest.NewRequest(http.MethodGet, "/pacientes/search?edad_min=20&edad_max=40&peso_min=60&peso_max=70", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())

	// Act
	err := SearchPacientes(c, repo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	var pacientes []models.Paciente
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &pacientes)
	require.NoError(t, err)
	// Solo Maria cumple: edad 25 (20-40) y peso 62 (60-70)
	assert.Len(t, pacientes, 1)
	assert.Equal(t, "Maria Gomez", pacientes[0].Nombre)
}