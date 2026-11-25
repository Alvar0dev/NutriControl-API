package handlers

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
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

func TestUpdatePacienteOK(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	creado, err := repo.Create(context.Background(), models.Paciente{Nombre: "Ana Lopez", Edad: 28, Peso: 68.5, PorcentajeGrasa: 24.0, Notas: "Nueva consulta"})
	require.NoError(t, err)
	require.NotEmpty(t, creado.ID)

	payload := `{"nombre":"Ana Lopez Updated","edad":30,"peso":70.0,"porcentaje_grasa":25.0,"notas":"Control mensual"}`
	req := httptest.NewRequest(http.MethodPut, "/pacientes/"+creado.ID, strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())
	setParam(c, "id", creado.ID)

	// Act
	err = UpdatePaciente(c, repo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	var actualizado models.Paciente
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &actualizado)
	require.NoError(t, err)
	assert.Equal(t, creado.ID, actualizado.ID)
	assert.Equal(t, "Ana Lopez Updated", actualizado.Nombre)
	assert.Equal(t, 30, actualizado.Edad)
	assert.Equal(t, 70.0, actualizado.Peso)
	assert.Equal(t, 25.0, actualizado.PorcentajeGrasa)
	assert.Equal(t, "Control mensual", actualizado.Notas)
}

func TestUpdatePacienteNoEncontrado(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()

	payload := `{"nombre":"Inexistente","edad":30,"peso":70.0}`
	req := httptest.NewRequest(http.MethodPut, "/pacientes/inexistente", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())
	setParam(c, "id", "inexistente")

	// Act
	err := UpdatePaciente(c, repo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, c.Response().Status)

	var resp map[string]string
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp["error"], "No se encontró paciente")
}

// TestPatchPacienteOK: actualización parcial, solo nombre cambia.
func TestPatchPacienteOK(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	creado, err := repo.Create(context.Background(), models.Paciente{Nombre: "Ana Lopez", Edad: 28, Peso: 68.5, PorcentajeGrasa: 24.0, Notas: "Nueva consulta"})
	require.NoError(t, err)
	require.NotEmpty(t, creado.ID)

	payload := `{"nombre":"Ana Lopez Modificada"}`
	req := httptest.NewRequest(http.MethodPatch, "/pacientes/"+creado.ID, strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())
	setParam(c, "id", creado.ID)

	// Act
	err = PatchPaciente(c, repo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	var actualizado models.Paciente
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &actualizado)
	require.NoError(t, err)
	assert.Equal(t, creado.ID, actualizado.ID)
	assert.Equal(t, "Ana Lopez Modificada", actualizado.Nombre)
	assert.Equal(t, 28, actualizado.Edad)
	assert.Equal(t, 68.5, actualizado.Peso)
	assert.Equal(t, 24.0, actualizado.PorcentajeGrasa)
	assert.Equal(t, "Nueva consulta", actualizado.Notas)
}

// TestPatchPacienteSoloPeso: solo el peso se actualiza, el resto se conserva.
func TestPatchPacienteSoloPeso(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	creado, err := repo.Create(context.Background(), models.Paciente{Nombre: "Ana Lopez", Edad: 28, Peso: 68.5, PorcentajeGrasa: 24.0, Notas: "Nueva consulta"})
	require.NoError(t, err)
	require.NotEmpty(t, creado.ID)

	payload := `{"peso":72.0}`
	req := httptest.NewRequest(http.MethodPatch, "/pacientes/"+creado.ID, strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())
	setParam(c, "id", creado.ID)

	// Act
	err = PatchPaciente(c, repo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	var actualizado models.Paciente
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &actualizado)
	require.NoError(t, err)
	assert.Equal(t, creado.ID, actualizado.ID)
	assert.Equal(t, "Ana Lopez", actualizado.Nombre)
	assert.Equal(t, 28, actualizado.Edad)
	assert.Equal(t, 72.0, actualizado.Peso)
	assert.Equal(t, 24.0, actualizado.PorcentajeGrasa)
	assert.Equal(t, "Nueva consulta", actualizado.Notas)
}

// TestPatchPacienteNoEncontrado: ID inexistente -> 404.
func TestPatchPacienteNoEncontrado(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()

	payload := `{"nombre":"Inexistente"}`
	req := httptest.NewRequest(http.MethodPatch, "/pacientes/inexistente", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())
	setParam(c, "id", "inexistente")

	// Act
	err := PatchPaciente(c, repo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, c.Response().Status)

	var resp map[string]string
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp["error"], "No se encontró paciente")
}

// TestPatchPacienteEdadInvalida: edad <= 0 enviado -> 400.
func TestPatchPacienteEdadInvalida(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	creado, err := repo.Create(context.Background(), models.Paciente{Nombre: "Ana Lopez", Edad: 28, Peso: 68.5})
	require.NoError(t, err)
	require.NotEmpty(t, creado.ID)

	payload := `{"edad":0}`
	req := httptest.NewRequest(http.MethodPatch, "/pacientes/"+creado.ID, strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())
	setParam(c, "id", creado.ID)

	// Act
	err = PatchPaciente(c, repo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, c.Response().Status)

	var resp map[string]string
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp["error"], "edad")
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

// TestGetPacientesPorEdadOK filtra por edad exacta y devuelve los pacientes que coinciden.
func TestGetPacientesPorEdadOK(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	req := httptest.NewRequest(http.MethodGet, "/pacientes/edad/30", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())
	setParam(c, "edad", "30")

	// Act
	err := GetPacientesPorEdad(c, repo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	var pacientes []models.Paciente
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &pacientes)
	require.NoError(t, err)
	assert.Len(t, pacientes, 1)
	assert.Equal(t, "Juan Perez", pacientes[0].Nombre)
	assert.Equal(t, 30, pacientes[0].Edad)
}

// TestGetPacientesPorEdadVacio devuelve lista vacía cuando no hay coincidencia de edad.
func TestGetPacientesPorEdadVacio(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	req := httptest.NewRequest(http.MethodGet, "/pacientes/edad/99", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())
	setParam(c, "edad", "99")

	// Act
	err := GetPacientesPorEdad(c, repo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	var pacientes []models.Paciente
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &pacientes)
	require.NoError(t, err)
	assert.Len(t, pacientes, 0)
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

func TestStatsPacientesOK(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	req := httptest.NewRequest(http.MethodGet, "/pacientes/stats", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())

	// Act
	err := StatsPacientes(c, repo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	var stats struct {
		Total                  int     `json:"total"`
		EdadPromedio           float64 `json:"edad_promedio"`
		PesoPromedio           float64 `json:"peso_promedio"`
		PorcentajeGrasaPromedio float64 `json:"porcentaje_grasa_promedio"`
		Edades                 []int   `json:"edades"`
		Pesos                  []float64 `json:"pesos"`
	}
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &stats)
	require.NoError(t, err)
	assert.Equal(t, 2, stats.Total)
	assert.Equal(t, 27.5, stats.EdadPromedio)
	assert.Equal(t, 68.75, stats.PesoPromedio)
	assert.Equal(t, 20.25, stats.PorcentajeGrasaPromedio)
	assert.Equal(t, []int{25, 30}, stats.Edades)
	assert.Equal(t, []float64{62.0, 75.5}, stats.Pesos)
}

func TestStatsPacientesListaVacia(t *testing.T) {
	// Arrange: repositorio sin pacientes (mock nuevo sin seed)
	e := newTestEcho()
	repo := &firestore.MockPacienteRepository{}
	req := httptest.NewRequest(http.MethodGet, "/pacientes/stats", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())

	// Act
	err := StatsPacientes(c, repo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	var stats struct {
		Total                  int     `json:"total"`
		EdadPromedio           float64 `json:"edad_promedio"`
		PesoPromedio           float64 `json:"peso_promedio"`
		PorcentajeGrasaPromedio float64 `json:"porcentaje_grasa_promedio"`
		Edades                 []int   `json:"edades"`
		Pesos                  []float64 `json:"pesos"`
	}
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &stats)
	require.NoError(t, err)
	assert.Equal(t, 0, stats.Total)
	assert.Equal(t, 0.0, stats.EdadPromedio)
	assert.Equal(t, 0.0, stats.PesoPromedio)
	assert.Equal(t, 0.0, stats.PorcentajeGrasaPromedio)
	assert.Len(t, stats.Edades, 0)
	assert.Len(t, stats.Pesos, 0)
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

// TestGetPacientesPaginacionLimitOffset verifica paginación nativa del repo vía GetPacientes.
func TestGetPacientesPaginacionLimitOffset(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	// store inicial: Juan, Maria
	for i := 0; i < 3; i++ {
		_, _ = repo.Create(context.Background(), models.Paciente{
			Nombre: fmt.Sprintf("Extra %d", i),
			Edad:   20 + i,
			Peso:   60.0 + float64(i),
		})
	}
	// total 5: Juan, Maria, Extra 0, Extra 1, Extra 2

	req := httptest.NewRequest(http.MethodGet, "/pacientes?limit=2&offset=1", nil)
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
	assert.Equal(t, "Maria Gomez", pacientes[0].Nombre)
	assert.Equal(t, "Extra 0", pacientes[1].Nombre)
}

// TestGetPacientesOffsetSuperaTotal verifica que offset mayor al total devuelve vacío.
func TestGetPacientesOffsetSuperaTotal(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()

	req := httptest.NewRequest(http.MethodGet, "/pacientes?limit=5&offset=100", nil)
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
	assert.Len(t, pacientes, 0)
}

// TestCreatePacienteEdadFueraDeRango: edad mayor a 120 -> 400.
func TestCreatePacienteEdadFueraDeRango(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	payload := `{"nombre":"Ana Lopez","edad":150,"peso":68.5,"porcentaje_grasa":24.0,"notas":"Nueva consulta"}`
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

// TestCreatePacientePesoFueraDeRango: peso mayor a 500 -> 400.
func TestCreatePacientePesoFueraDeRango(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	payload := `{"nombre":"Ana Lopez","edad":28,"peso":600,"porcentaje_grasa":24.0,"notas":"Nueva consulta"}`
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

// TestCreatePacienteNombreConEspaciosSeNormaliza: "  juan perez  " se guarda como "Juan Perez".
func TestCreatePacienteNombreConEspaciosSeNormaliza(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	payload := `{"nombre":"  juan perez  ","edad":28,"peso":68.5,"porcentaje_grasa":24.0,"notas":"Nueva consulta"}`
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
	assert.Equal(t, "Juan Perez", creado.Nombre)
}

// TestCreatePacienteNombreSoloEspacios: nombre solo espacios -> 400.
func TestCreatePacienteNombreSoloEspacios(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	payload := `{"nombre":"   ","edad":28,"peso":68.5,"porcentaje_grasa":24.0,"notas":"Nueva consulta"}`
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

// TestCreatePacientePorcentajeGrasaFueraDeRango verifica que porcentaje_grasa fuera de [0, 100] devuelve 400.
func TestCreatePacientePorcentajeGrasaFueraDeRango(t *testing.T) {
	tests := []struct {
		name    string
		payload string
	}{
		{
			name:    "porcentaje de grasa negativo",
			payload: `{"nombre":"Ana Lopez","edad":28,"peso":68.5,"porcentaje_grasa":-5.0}`,
		},
		{
			name:    "porcentaje de grasa mayor a 100",
			payload: `{"nombre":"Ana Lopez","edad":28,"peso":68.5,"porcentaje_grasa":105.0}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			repo := firestore.NewMockPacienteRepository()
			req := httptest.NewRequest(http.MethodPost, "/pacientes", strings.NewReader(tt.payload))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c := e.NewContext(req, httptest.NewRecorder())

			err := CreatePaciente(c, repo)

			require.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, c.Response().Status)

			var resp map[string]string
			err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Contains(t, resp["error"], "porcentaje_grasa")
		})
	}
}

// TestUpdatePacientePorcentajeGrasaFueraDeRango verifica 400 en PUT cuando porcentaje_grasa supera 100.
func TestUpdatePacientePorcentajeGrasaFueraDeRango(t *testing.T) {
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	payload := `{"nombre":"Ana Lopez","edad":28,"peso":68.5,"porcentaje_grasa":110.0}`
	req := httptest.NewRequest(http.MethodPut, "/pacientes/1", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())
	setParam(c, "id", "1")

	err := UpdatePaciente(c, repo)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, c.Response().Status)

	var resp map[string]string
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp["error"], "porcentaje_grasa")
}

// TestPatchPacienteValidacionesExtra verifica validaciones de porcentaje_grasa y nombre vacío en PATCH.
func TestPatchPacienteValidacionesExtra(t *testing.T) {
	tests := []struct {
		name          string
		payload       string
		expectedError string
	}{
		{
			name:          "porcentaje de grasa negativo en PATCH",
			payload:       `{"porcentaje_grasa":-1.0}`,
			expectedError: "porcentaje_grasa",
		},
		{
			name:          "porcentaje de grasa mayor a 100 en PATCH",
			payload:       `{"porcentaje_grasa":120.0}`,
			expectedError: "porcentaje_grasa",
		},
		{
			name:          "nombre compuesto solo por espacios en PATCH",
			payload:       `{"nombre":"   "}`,
			expectedError: "nombre",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			repo := firestore.NewMockPacienteRepository()
			creado, err := repo.Create(context.Background(), models.Paciente{
				Nombre: "Ana Lopez", Edad: 28, Peso: 68.5, PorcentajeGrasa: 24.0,
			})
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPatch, "/pacientes/"+creado.ID, strings.NewReader(tt.payload))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c := e.NewContext(req, httptest.NewRecorder())
			setParam(c, "id", creado.ID)

			err = PatchPaciente(c, repo)

			require.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, c.Response().Status)

			var resp map[string]string
			err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Contains(t, resp["error"], tt.expectedError)
		})
	}
}

// TestDeletePacienteOK verifica eliminación exitosa (204 No Content) y baja efectiva del repositorio.
func TestDeletePacienteOK(t *testing.T) {
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	creado, err := repo.Create(context.Background(), models.Paciente{
		Nombre: "Ana Lopez", Edad: 28, Peso: 68.5,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodDelete, "/pacientes/"+creado.ID, nil)
	c := e.NewContext(req, httptest.NewRecorder())
	setParam(c, "id", creado.ID)

	err = DeletePaciente(c, repo)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, c.Response().Status)

	_, err = repo.GetByID(context.Background(), creado.ID)
	assert.Error(t, err)
}

// TestDeletePacienteNoEncontrado verifica 404 al intentar eliminar un ID inexistente.
func TestDeletePacienteNoEncontrado(t *testing.T) {
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()

	req := httptest.NewRequest(http.MethodDelete, "/pacientes/inexistente", nil)
	c := e.NewContext(req, httptest.NewRecorder())
	setParam(c, "id", "inexistente")

	err := DeletePaciente(c, repo)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, c.Response().Status)

	var resp map[string]string
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp["error"], "No se encontró paciente")
}

// TestGetPacientesPorEdadInvalida verifica 400 cuando el parámetro edad no es numérico.
func TestGetPacientesPorEdadInvalida(t *testing.T) {
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	req := httptest.NewRequest(http.MethodGet, "/pacientes/edad/abc", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())
	setParam(c, "edad", "abc")

	err := GetPacientesPorEdad(c, repo)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, c.Response().Status)

	var resp map[string]string
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp["error"], "edad")
}

// TestHealthCheckOK verifica que GET /health devuelve 200 OK con status y modo de almacenamiento.
func TestHealthCheckOK(t *testing.T) {
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	c := e.NewContext(req, httptest.NewRecorder())

	err := HealthCheck(c, repo)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	var resp map[string]string
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp["status"])
	assert.Equal(t, "mock", resp["storage"])
}

// TestCountPacientesOK verifica que GET /pacientes/count devuelve el total de pacientes.
func TestCountPacientesOK(t *testing.T) {
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	req := httptest.NewRequest(http.MethodGet, "/pacientes/count", nil)
	c := e.NewContext(req, httptest.NewRecorder())

	err := CountPacientes(c, repo)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	var resp map[string]int
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 2, resp["total"])
}

// TestCountPacientesVacio verifica que GET /pacientes/count devuelve total 0 cuando el repositorio está vacío.
func TestCountPacientesVacio(t *testing.T) {
	e := newTestEcho()
	repo := &firestore.MockPacienteRepository{}
	req := httptest.NewRequest(http.MethodGet, "/pacientes/count", nil)
	c := e.NewContext(req, httptest.NewRecorder())

	err := CountPacientes(c, repo)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	var resp map[string]int
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp["total"])
}

// TestSearchPacientesFiltrarPorPorcentajeGrasaMinMax verifica el filtro por rango de porcentaje_grasa en GET /pacientes/search.
func TestSearchPacientesFiltrarPorPorcentajeGrasaMinMax(t *testing.T) {
	// Seed del mock: Juan Perez (porcentaje_grasa: 18.5), Maria Gomez (porcentaje_grasa: 22.0)
	tests := []struct {
		name           string
		query          string
		expectedCount  int
		expectedNombre string
	}{
		{
			name:           "rango entre 20 y 25 devuelve solo Maria Gomez",
			query:          "/pacientes/search?porcentaje_grasa_min=20&porcentaje_grasa_max=25",
			expectedCount:  1,
			expectedNombre: "Maria Gomez",
		},
		{
			name:           "maximo 19 devuelve solo Juan Perez",
			query:          "/pacientes/search?porcentaje_grasa_max=19",
			expectedCount:  1,
			expectedNombre: "Juan Perez",
		},
		{
			name:          "minimo 30 devuelve lista vacia",
			query:         "/pacientes/search?porcentaje_grasa_min=30",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			repo := firestore.NewMockPacienteRepository()
			req := httptest.NewRequest(http.MethodGet, tt.query, nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c := e.NewContext(req, httptest.NewRecorder())

			err := SearchPacientes(c, repo)

			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, c.Response().Status)

			var pacientes []models.Paciente
			err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &pacientes)
			require.NoError(t, err)
			assert.Len(t, pacientes, tt.expectedCount)
			if tt.expectedCount == 1 {
				assert.Equal(t, tt.expectedNombre, pacientes[0].Nombre)
			}
		})
	}
}

// TestGetPacientesOrdenamiento verifica el ordenamiento por distintos campos y sentidos en GET /pacientes.
func TestGetPacientesOrdenamiento(t *testing.T) {
	// Seed del mock:
	// Juan Perez:  Edad 30, Peso 75.5, PorcentajeGrasa 18.5
	// Maria Gomez: Edad 25, Peso 62.0, PorcentajeGrasa 22.0
	tests := []struct {
		name          string
		query         string
		expectedFirst string
		expectedLast  string
	}{
		{
			name:          "ordenar por edad asc (Maria 25 primero, Juan 30 segundo)",
			query:         "/pacientes?sort=edad&order=asc",
			expectedFirst: "Maria Gomez",
			expectedLast:  "Juan Perez",
		},
		{
			name:          "ordenar por edad desc (Juan 30 primero, Maria 25 segundo)",
			query:         "/pacientes?sort=edad&order=desc",
			expectedFirst: "Juan Perez",
			expectedLast:  "Maria Gomez",
		},
		{
			name:          "ordenar por nombre desc (Maria Gomez primero, Juan Perez segundo)",
			query:         "/pacientes?sort=nombre&order=desc",
			expectedFirst: "Maria Gomez",
			expectedLast:  "Juan Perez",
		},
		{
			name:          "ordenar por peso asc (Maria 62.0 primero, Juan 75.5 segundo)",
			query:         "/pacientes?sort=peso&order=asc",
			expectedFirst: "Maria Gomez",
			expectedLast:  "Juan Perez",
		},
		{
			name:          "ordenar por porcentaje_grasa desc (Maria 22.0 primero, Juan 18.5 segundo)",
			query:         "/pacientes?sort=porcentaje_grasa&order=desc",
			expectedFirst: "Maria Gomez",
			expectedLast:  "Juan Perez",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			repo := firestore.NewMockPacienteRepository()
			req := httptest.NewRequest(http.MethodGet, tt.query, nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c := e.NewContext(req, httptest.NewRecorder())

			err := GetPacientes(c, repo)

			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, c.Response().Status)

			var pacientes []models.Paciente
			err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &pacientes)
			require.NoError(t, err)
			require.Len(t, pacientes, 2)
			assert.Equal(t, tt.expectedFirst, pacientes[0].Nombre)
			assert.Equal(t, tt.expectedLast, pacientes[1].Nombre)
		})
	}
}

// TestSearchPacientesOrdenamiento verifica el ordenamiento en GET /pacientes/search con y sin filtros.
func TestSearchPacientesOrdenamiento(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		expectedFirst string
		expectedLast  string
	}{
		{
			name:          "search con ordenamiento por edad asc",
			query:         "/pacientes/search?sort=edad&order=asc",
			expectedFirst: "Maria Gomez",
			expectedLast:  "Juan Perez",
		},
		{
			name:          "search con filtro edad_min=20 y ordenamiento por peso desc",
			query:         "/pacientes/search?edad_min=20&sort=peso&order=desc",
			expectedFirst: "Juan Perez",
			expectedLast:  "Maria Gomez",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			repo := firestore.NewMockPacienteRepository()
			req := httptest.NewRequest(http.MethodGet, tt.query, nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c := e.NewContext(req, httptest.NewRecorder())

			err := SearchPacientes(c, repo)

			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, c.Response().Status)

			var pacientes []models.Paciente
			err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &pacientes)
			require.NoError(t, err)
			require.Len(t, pacientes, 2)
			assert.Equal(t, tt.expectedFirst, pacientes[0].Nombre)
			assert.Equal(t, tt.expectedLast, pacientes[1].Nombre)
		})
	}
}

// TestExportPacientesCSVOK verifica la exportación de pacientes en formato CSV (cabeceras RFC 4180, datos y columna altura).
func TestExportPacientesCSVOK(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	_, err := repo.Create(context.Background(), models.Paciente{
		Nombre:          "Carlos Altura",
		Edad:            35,
		Peso:            78.0,
		Altura:          1.75,
		PorcentajeGrasa: 19.0,
		Notas:           "Con altura",
	})
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/pacientes/export", nil)
	c := e.NewContext(req, rec)

	// Act
	err = ExportPacientesCSV(c, repo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get(echo.HeaderContentType), "text/csv")
	assert.Equal(t, `attachment; filename="pacientes.csv"`, rec.Header().Get(echo.HeaderContentDisposition))

	reader := csv.NewReader(rec.Body)
	records, err := reader.ReadAll()
	require.NoError(t, err)

	// Cabecera + 2 pacientes del seed + Carlos Altura
	require.Len(t, records, 4)
	assert.Equal(t, []string{"id", "nombre", "edad", "peso", "porcentaje_grasa", "notas", "altura"}, records[0])

	// Paciente 1 (Juan Perez, sin altura)
	assert.Equal(t, "Juan Perez", records[1][1])
	assert.Equal(t, "30", records[1][2])
	assert.Equal(t, "75.5", records[1][3])
	assert.Equal(t, "18.5", records[1][4])
	assert.Equal(t, "Primera consulta", records[1][5])
	assert.Equal(t, "", records[1][6])

	// Paciente 2 (Maria Gomez, sin altura)
	assert.Equal(t, "Maria Gomez", records[2][1])
	assert.Equal(t, "25", records[2][2])
	assert.Equal(t, "62", records[2][3])
	assert.Equal(t, "22", records[2][4])
	assert.Equal(t, "Seguimiento mensual", records[2][5])
	assert.Equal(t, "", records[2][6])

	// Paciente 3 (Carlos Altura, 1.75)
	assert.Equal(t, "Carlos Altura", records[3][1])
	assert.Equal(t, "1.75", records[3][6])
}

// TestExportPacientesCSVVacio verifica que un repositorio sin pacientes genere únicamente la cabecera CSV.
func TestExportPacientesCSVVacio(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := &firestore.MockPacienteRepository{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/pacientes/export", nil)
	c := e.NewContext(req, rec)

	// Act
	err := ExportPacientesCSV(c, repo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	reader := csv.NewReader(rec.Body)
	records, err := reader.ReadAll()
	require.NoError(t, err)

	require.Len(t, records, 1)
	assert.Equal(t, []string{"id", "nombre", "edad", "peso", "porcentaje_grasa", "notas", "altura"}, records[0])
}

// TestExportPacientesCSVRutaRegistrada verifica que el router dirija /pacientes/export al handler de exportación
// y no caiga en /pacientes/{id}.
func TestExportPacientesCSVRutaRegistrada(t *testing.T) {
	e := echo.New()
	repo := firestore.NewMockPacienteRepository()
	SetupRoutes(e, repo)

	req := httptest.NewRequest(http.MethodGet, "/pacientes/export", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get(echo.HeaderContentType), "text/csv")
}

// TestCreatePacientesBatchOK verifica que se pueda crear un lote de pacientes válidos con 201 Created y que se asignen IDs.
func TestCreatePacientesBatchOK(t *testing.T) {
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	payload := `[
		{"nombre":"  carlos tevez  ","edad":35,"peso":75.0,"porcentaje_grasa":15.0,"notas":"Atleta"},
		{"nombre":"lucia galan","edad":42,"peso":62.0,"porcentaje_grasa":22.5,"notas":"Control"}
	]`
	req := httptest.NewRequest(http.MethodPost, "/pacientes/batch", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())

	err := CreatePacientesBatch(c, repo)

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, c.Response().Status)

	var creados []models.Paciente
	err = json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &creados)
	require.NoError(t, err)
	require.Len(t, creados, 2)
	assert.Equal(t, "Carlos Tevez", creados[0].Nombre)
	assert.NotEmpty(t, creados[0].ID)
	assert.Equal(t, "Lucia Galan", creados[1].Nombre)
	assert.NotEmpty(t, creados[1].ID)

	// Verificar que el repositorio ahora tiene 4 pacientes (2 iniciales + 2 creados)
	todos, err := repo.GetAll(context.Background(), 0, 0)
	require.NoError(t, err)
	assert.Len(t, todos, 4)
}

// TestCreatePacientesBatchVacio verifica que un lote vacío devuelva 400 Bad Request.
func TestCreatePacientesBatchVacio(t *testing.T) {
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	payload := `[]`
	req := httptest.NewRequest(http.MethodPost, "/pacientes/batch", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())

	err := CreatePacientesBatch(c, repo)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, c.Response().Status)
	assert.Contains(t, c.Response().Writer.(*httptest.ResponseRecorder).Body.String(), "El lote de pacientes no puede estar vacío")
}

// TestCreatePacientesBatchAtomicidadInvalido verifica que si un elemento del lote es inválido,
// la petición falle con 400 Bad Request y NO se inserte ningún paciente (atomicidad).
func TestCreatePacientesBatchAtomicidadInvalido(t *testing.T) {
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	// Primer paciente válido, segundo con edad inválida (-10)
	payload := `[
		{"nombre":"Paciente Valido","edad":30,"peso":70.0,"porcentaje_grasa":18.0},
		{"nombre":"Paciente Invalido","edad":-10,"peso":70.0,"porcentaje_grasa":18.0}
	]`
	req := httptest.NewRequest(http.MethodPost, "/pacientes/batch", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())

	err := CreatePacientesBatch(c, repo)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, c.Response().Status)
	assert.Contains(t, c.Response().Writer.(*httptest.ResponseRecorder).Body.String(), "edad")

	// Verificar atomicidad: el repo DEBE seguir teniendo solo los 2 pacientes iniciales
	todos, err := repo.GetAll(context.Background(), 0, 0)
	require.NoError(t, err)
	assert.Len(t, todos, 2)
}

// TestCreatePacientesBatchJSONInvalido verifica error 400 ante payload no parseable.
func TestCreatePacientesBatchJSONInvalido(t *testing.T) {
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	payload := `{"esto_no_es_un_array": true}`
	req := httptest.NewRequest(http.MethodPost, "/pacientes/batch", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())

	err := CreatePacientesBatch(c, repo)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, c.Response().Status)
}

// TestCreatePacientesBatchRutaRegistrada verifica que el router dirija POST /pacientes/batch correctamente.
func TestCreatePacientesBatchRutaRegistrada(t *testing.T) {
	e := echo.New()
	repo := firestore.NewMockPacienteRepository()
	SetupRoutes(e, repo)

	payload := `[{"nombre":"Test Router","edad":25,"peso":60.0,"porcentaje_grasa":15.0}]`
	req := httptest.NewRequest(http.MethodPost, "/pacientes/batch", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}

// TestGetPacienteIMCOK verifica el cálculo exitoso de IMC para un paciente con peso y altura válidos.
func TestGetPacienteIMCOK(t *testing.T) {
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	creado, err := repo.Create(context.Background(), models.Paciente{
		Nombre: "Lucas Silva",
		Edad:   28,
		Peso:   75.5,
		Altura: 1.75,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/pacientes/imc/"+creado.ID, nil)
	c := e.NewContext(req, httptest.NewRecorder())
	setParam(c, "id", creado.ID)

	err = GetPacienteIMC(c, repo)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	var resp map[string]any
	body := c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes()
	err = json.Unmarshal(body, &resp)
	require.NoError(t, err)

	assert.Equal(t, creado.ID, resp["id"])
	assert.Equal(t, "Lucas Silva", resp["nombre"])
	assert.Equal(t, 75.5, resp["peso"])
	assert.Equal(t, 1.75, resp["altura"])
	assert.Equal(t, 24.65, resp["imc"])
	assert.Equal(t, "Normal", resp["clasificacion"])
}

// TestGetPacienteIMCRutasRegistradas verifica que el router soporte tanto /pacientes/imc/{id} como /pacientes/{id}/imc.
func TestGetPacienteIMCRutasRegistradas(t *testing.T) {
	e := echo.New()
	repo := firestore.NewMockPacienteRepository()
	SetupRoutes(e, repo)

	creado, err := repo.Create(context.Background(), models.Paciente{
		Nombre: "Elena Rossi",
		Edad:   32,
		Peso:   60.0,
		Altura: 1.65,
	})
	require.NoError(t, err)

	// Probar ruta 1: /pacientes/imc/{id}
	req1 := httptest.NewRequest(http.MethodGet, "/pacientes/imc/"+creado.ID, nil)
	rec1 := httptest.NewRecorder()
	e.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusOK, rec1.Code)

	// Probar ruta 2: /pacientes/{id}/imc
	req2 := httptest.NewRequest(http.MethodGet, "/pacientes/"+creado.ID+"/imc", nil)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusOK, rec2.Code)
}

// TestGetPacienteIMCNoEncontrado verifica error 404 para ID inexistente.
func TestGetPacienteIMCNoEncontrado(t *testing.T) {
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()

	req := httptest.NewRequest(http.MethodGet, "/pacientes/imc/paciente-inexistente", nil)
	c := e.NewContext(req, httptest.NewRecorder())
	setParam(c, "id", "paciente-inexistente")

	err := GetPacienteIMC(c, repo)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, c.Response().Status)
}

// TestGetPacienteIMCSinAltura verifica error 400 cuando el paciente no tiene altura registrada (altura <= 0).
func TestGetPacienteIMCSinAltura(t *testing.T) {
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	creado, err := repo.Create(context.Background(), models.Paciente{
		Nombre: "Marcos Sin Altura",
		Edad:   40,
		Peso:   80.0,
		Altura: 0,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/pacientes/imc/"+creado.ID, nil)
	c := e.NewContext(req, httptest.NewRecorder())
	setParam(c, "id", creado.ID)

	err = GetPacienteIMC(c, repo)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, c.Response().Status)
	assert.Contains(t, c.Response().Writer.(*httptest.ResponseRecorder).Body.String(), "altura")
}

// TestGetPacienteIMCClasificaciones verifica table-driven las clasificaciones OMS según el IMC resultante.
func TestGetPacienteIMCClasificaciones(t *testing.T) {
	casos := []struct {
		name                 string
		peso                 float64
		altura               float64
		expectedIMC          float64
		expectedClasificacion string
	}{
		{
			name:                 "Bajo peso (IMC < 18.5)",
			peso:                 50.0,
			altura:               1.75,
			expectedIMC:          16.33,
			expectedClasificacion: "Bajo peso",
		},
		{
			name:                 "Normal (18.5 <= IMC < 25.0)",
			peso:                 70.0,
			altura:               1.75,
			expectedIMC:          22.86,
			expectedClasificacion: "Normal",
		},
		{
			name:                 "Sobrepeso (25.0 <= IMC < 30.0)",
			peso:                 85.0,
			altura:               1.75,
			expectedIMC:          27.76,
			expectedClasificacion: "Sobrepeso",
		},
		{
			name:                 "Obesidad (IMC >= 30.0)",
			peso:                 100.0,
			altura:               1.75,
			expectedIMC:          32.65,
			expectedClasificacion: "Obesidad",
		},
	}

	for _, tc := range casos {
		t.Run(tc.name, func(t *testing.T) {
			e := newTestEcho()
			repo := firestore.NewMockPacienteRepository()
			creado, err := repo.Create(context.Background(), models.Paciente{
				Nombre: "Test IMC",
				Edad:   30,
				Peso:   tc.peso,
				Altura: tc.altura,
			})
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, "/pacientes/imc/"+creado.ID, nil)
			c := e.NewContext(req, httptest.NewRecorder())
			setParam(c, "id", creado.ID)

			err = GetPacienteIMC(c, repo)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, c.Response().Status)

			var resp map[string]any
			body := c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes()
			err = json.Unmarshal(body, &resp)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedIMC, resp["imc"])
			assert.Equal(t, tc.expectedClasificacion, resp["clasificacion"])
		})
	}
}

// TestCreatePacienteAlturaValida verifica que se pueda crear un paciente informando su altura.
func TestCreatePacienteAlturaValida(t *testing.T) {
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	payload := `{"nombre":"Martin Palermo","edad":45,"peso":85.0,"altura":1.87,"porcentaje_grasa":18.0}`
	req := httptest.NewRequest(http.MethodPost, "/pacientes", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())

	err := CreatePaciente(c, repo)

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, c.Response().Status)

	var creado models.Paciente
	body := c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes()
	err = json.Unmarshal(body, &creado)
	require.NoError(t, err)
	assert.Equal(t, 1.87, creado.Altura)
}

// TestCreatePacienteAlturaFueraDeRango verifica rechazo 400 cuando la altura enviada excede los límites (0.30 a 2.50 m) o es negativa.
func TestCreatePacienteAlturaFueraDeRango(t *testing.T) {
	casos := []struct {
		name    string
		payload string
	}{
		{
			name:    "altura menor a 0.30 m",
			payload: `{"nombre":"Test Altura Baja","edad":30,"peso":70.0,"altura":0.25,"porcentaje_grasa":20.0}`,
		},
		{
			name:    "altura mayor a 2.50 m",
			payload: `{"nombre":"Test Altura Gigante","edad":30,"peso":70.0,"altura":2.80,"porcentaje_grasa":20.0}`,
		},
		{
			name:    "altura negativa",
			payload: `{"nombre":"Test Altura Negativa","edad":30,"peso":70.0,"altura":-1.75,"porcentaje_grasa":20.0}`,
		},
	}

	for _, tc := range casos {
		t.Run(tc.name, func(t *testing.T) {
			e := newTestEcho()
			repo := firestore.NewMockPacienteRepository()
			req := httptest.NewRequest(http.MethodPost, "/pacientes", strings.NewReader(tc.payload))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c := e.NewContext(req, httptest.NewRecorder())

			err := CreatePaciente(c, repo)
			require.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, c.Response().Status)
			assert.Contains(t, c.Response().Writer.(*httptest.ResponseRecorder).Body.String(), "altura")
		})
	}
}

// TestPatchPacienteAltura verifica la actualización parcial del campo altura.
func TestPatchPacienteAltura(t *testing.T) {
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	creado, err := repo.Create(context.Background(), models.Paciente{
		Nombre: "Paciente Para Patch Altura",
		Edad:   29,
		Peso:   72.0,
	})
	require.NoError(t, err)

	payload := `{"altura": 1.74}`
	req := httptest.NewRequest(http.MethodPatch, "/pacientes/"+creado.ID, strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())
	setParam(c, "id", creado.ID)

	err = PatchPaciente(c, repo)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	var resp models.Paciente
	body := c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes()
	err = json.Unmarshal(body, &resp)
	require.NoError(t, err)
	assert.Equal(t, 1.74, resp.Altura)
}

// TestPatchPacienteAlturaInvalida verifica rechazo 400 al intentar patchear una altura fuera de rango.
func TestPatchPacienteAlturaInvalida(t *testing.T) {
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	creado, err := repo.Create(context.Background(), models.Paciente{
		Nombre: "Paciente Para Patch Invalido",
		Edad:   29,
		Peso:   72.0,
	})
	require.NoError(t, err)

	payload := `{"altura": 3.10}`
	req := httptest.NewRequest(http.MethodPatch, "/pacientes/"+creado.ID, strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, httptest.NewRecorder())
	setParam(c, "id", creado.ID)

	err = PatchPaciente(c, repo)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, c.Response().Status)
	assert.Contains(t, c.Response().Writer.(*httptest.ResponseRecorder).Body.String(), "altura")
}

// TestReportePacientesOK verifica la generación del reporte analítico con todos los rangos etarios y categorías OMS de IMC.
func TestReportePacientesOK(t *testing.T) {
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()

	extras := []models.Paciente{
		{Nombre: "Menor Bajo Peso", Edad: 15, Peso: 45.0, Altura: 1.70, PorcentajeGrasa: 12.0},
		{Nombre: "Adulto Normal", Edad: 35, Peso: 70.0, Altura: 1.75, PorcentajeGrasa: 20.0},
		{Nombre: "Adulto Sobrepeso", Edad: 50, Peso: 85.0, Altura: 1.75, PorcentajeGrasa: 26.0},
		{Nombre: "Mayor Obesidad", Edad: 70, Peso: 100.0, Altura: 1.75, PorcentajeGrasa: 32.0},
	}
	for _, p := range extras {
		_, err := repo.Create(context.Background(), p)
		require.NoError(t, err)
	}

	req := httptest.NewRequest(http.MethodGet, "/pacientes/reporte", nil)
	c := e.NewContext(req, httptest.NewRecorder())

	err := ReportePacientes(c, repo)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	var resp RespuestaReporte
	body := c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes()
	err = json.Unmarshal(body, &resp)
	require.NoError(t, err)

	assert.Equal(t, 6, resp.Total)
	assert.Equal(t, 1, resp.GruposEtarios.Menores)
	assert.Equal(t, 4, resp.GruposEtarios.Adultos)
	assert.Equal(t, 1, resp.GruposEtarios.Mayores)

	assert.Equal(t, 1, resp.ClasificacionIMC.BajoPeso)
	assert.Equal(t, 1, resp.ClasificacionIMC.Normal)
	assert.Equal(t, 1, resp.ClasificacionIMC.Sobrepeso)
	assert.Equal(t, 1, resp.ClasificacionIMC.Obesidad)
	assert.Equal(t, 2, resp.ClasificacionIMC.SinDatos)

	assert.Equal(t, 37.5, resp.Promedios.Edad)
	assert.Equal(t, 72.92, resp.Promedios.Peso)
	assert.Equal(t, 21.75, resp.Promedios.PorcentajeGrasa)
	assert.Equal(t, 24.71, resp.Promedios.IMC)
}

// TestReportePacientesVacio verifica que un repositorio vacío retorne todos los contadores y promedios en 0.
func TestReportePacientesVacio(t *testing.T) {
	e := newTestEcho()
	repo := &firestore.MockPacienteRepository{}

	req := httptest.NewRequest(http.MethodGet, "/pacientes/reporte", nil)
	c := e.NewContext(req, httptest.NewRecorder())

	err := ReportePacientes(c, repo)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	var resp RespuestaReporte
	body := c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes()
	err = json.Unmarshal(body, &resp)
	require.NoError(t, err)

	assert.Equal(t, 0, resp.Total)
	assert.Equal(t, GruposEtarios{}, resp.GruposEtarios)
	assert.Equal(t, ClasificacionIMCPoblacion{}, resp.ClasificacionIMC)
	assert.Equal(t, PromediosReporte{}, resp.Promedios)
}

// TestReportePacientesRutaRegistrada verifica que el router dirija GET /pacientes/reporte antes que /pacientes/:id.
func TestReportePacientesRutaRegistrada(t *testing.T) {
	e := echo.New()
	repo := firestore.NewMockPacienteRepository()
	SetupRoutes(e, repo)

	req := httptest.NewRequest(http.MethodGet, "/pacientes/reporte", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp RespuestaReporte
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 2, resp.Total)
}

// TestSoftDeleteExcluyeDeListadosYEndpoints verifica que un paciente eliminado lógicamente quede fuera de listados, conteo, reporte, search y GET por ID.
func TestSoftDeleteExcluyeDeListadosYEndpoints(t *testing.T) {
	e := echo.New()
	repo := firestore.NewMockPacienteRepository()
	SetupRoutes(e, repo)

	creado, err := repo.Create(context.Background(), models.Paciente{
		Nombre: "Paciente Baja",
		Edad:   33,
		Peso:   70.0,
		Altura: 1.75,
	})
	require.NoError(t, err)

	// Eliminar vía DELETE
	reqDel := httptest.NewRequest(http.MethodDelete, "/pacientes/"+creado.ID, nil)
	recDel := httptest.NewRecorder()
	e.ServeHTTP(recDel, reqDel)
	assert.Equal(t, http.StatusNoContent, recDel.Code)

	// Segunda eliminación debe devolver 404
	reqDel2 := httptest.NewRequest(http.MethodDelete, "/pacientes/"+creado.ID, nil)
	recDel2 := httptest.NewRecorder()
	e.ServeHTTP(recDel2, reqDel2)
	assert.Equal(t, http.StatusNotFound, recDel2.Code)

	// GET por ID devuelve 404
	reqGet := httptest.NewRequest(http.MethodGet, "/pacientes/"+creado.ID, nil)
	recGet := httptest.NewRecorder()
	e.ServeHTTP(recGet, reqGet)
	assert.Equal(t, http.StatusNotFound, recGet.Code)

	// Count vuelve a 2 (solo seed activo)
	reqCount := httptest.NewRequest(http.MethodGet, "/pacientes/count", nil)
	recCount := httptest.NewRecorder()
	e.ServeHTTP(recCount, reqCount)
	assert.Equal(t, http.StatusOK, recCount.Code)
	var countResp map[string]int
	require.NoError(t, json.Unmarshal(recCount.Body.Bytes(), &countResp))
	assert.Equal(t, 2, countResp["total"])
}

// TestGetPacientesInactivosOK verifica el listado de pacientes dados de baja lógica y su ruta en router.
func TestGetPacientesInactivosOK(t *testing.T) {
	e := echo.New()
	repo := firestore.NewMockPacienteRepository()
	SetupRoutes(e, repo)

	creado, err := repo.Create(context.Background(), models.Paciente{
		Nombre: "Inactivo Uno",
		Edad:   41,
		Peso:   78.0,
	})
	require.NoError(t, err)
	require.NoError(t, repo.Delete(context.Background(), creado.ID))

	req := httptest.NewRequest(http.MethodGet, "/pacientes/inactivos", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var inactivos []models.Paciente
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &inactivos))
	require.Len(t, inactivos, 1)
	assert.Equal(t, creado.ID, inactivos[0].ID)
	assert.NotNil(t, inactivos[0].DeletedAt)
}

// TestRestorePacienteOK verifica la restauración de un paciente eliminado lógicamente.
func TestRestorePacienteOK(t *testing.T) {
	e := echo.New()
	repo := firestore.NewMockPacienteRepository()
	SetupRoutes(e, repo)

	creado, err := repo.Create(context.Background(), models.Paciente{
		Nombre: "Restaurable",
		Edad:   36,
		Peso:   69.0,
	})
	require.NoError(t, err)
	require.NoError(t, repo.Delete(context.Background(), creado.ID))

	req := httptest.NewRequest(http.MethodPost, "/pacientes/"+creado.ID+"/restore", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var restaurado models.Paciente
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &restaurado))
	assert.Equal(t, creado.ID, restaurado.ID)
	assert.Nil(t, restaurado.DeletedAt)

	// Vuelve a ser accesible por GET
	reqGet := httptest.NewRequest(http.MethodGet, "/pacientes/"+creado.ID, nil)
	recGet := httptest.NewRecorder()
	e.ServeHTTP(recGet, reqGet)
	assert.Equal(t, http.StatusOK, recGet.Code)
}

// TestRestorePacienteErrores verifica 404 al intentar restaurar un paciente inexistente o ya activo.
func TestRestorePacienteErrores(t *testing.T) {
	e := echo.New()
	repo := firestore.NewMockPacienteRepository()
	SetupRoutes(e, repo)

	creado, err := repo.Create(context.Background(), models.Paciente{
		Nombre: "Ya Activo",
		Edad:   27,
		Peso:   64.0,
	})
	require.NoError(t, err)

	// Ya activo -> 404
	req1 := httptest.NewRequest(http.MethodPost, "/pacientes/"+creado.ID+"/restore", nil)
	rec1 := httptest.NewRecorder()
	e.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusNotFound, rec1.Code)

	// Inexistente -> 404
	req2 := httptest.NewRequest(http.MethodPost, "/pacientes/no-existe/restore", nil)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusNotFound, rec2.Code)
}

// TestCreateControlPacienteOK verifica creación de control clínico con cálculo automático de IMC cuando hay altura y sin IMC cuando falta altura.
func TestCreateControlPacienteOK(t *testing.T) {
	e := echo.New()
	repo := firestore.NewMockPacienteRepository()
	SetupRoutes(e, repo)

	conAltura, err := repo.Create(context.Background(), models.Paciente{
		Nombre: "Paciente Con Altura",
		Edad:   30,
		Peso:   80.0,
		Altura: 1.75,
	})
	require.NoError(t, err)

	payload1 := `{"peso":75.5,"porcentaje_grasa":18.0,"notas":"Control mensual"}`
	req1 := httptest.NewRequest(http.MethodPost, "/pacientes/"+conAltura.ID+"/controles", strings.NewReader(payload1))
	req1.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec1 := httptest.NewRecorder()
	e.ServeHTTP(rec1, req1)

	assert.Equal(t, http.StatusCreated, rec1.Code)
	var ctrl1 models.ControlClinico
	require.NoError(t, json.Unmarshal(rec1.Body.Bytes(), &ctrl1))
	assert.NotEmpty(t, ctrl1.ID)
	assert.Equal(t, conAltura.ID, ctrl1.PacienteID)
	assert.False(t, ctrl1.Fecha.IsZero())
	assert.Equal(t, 75.5, ctrl1.Peso)
	assert.Equal(t, 18.0, ctrl1.PorcentajeGrasa)
	assert.Equal(t, 24.65, ctrl1.IMC)
	assert.Equal(t, "Control mensual", ctrl1.Notas)

	// Paciente sin altura -> IMC queda en 0
	sinAltura, err := repo.Create(context.Background(), models.Paciente{
		Nombre: "Paciente Sin Altura",
		Edad:   32,
		Peso:   70.0,
	})
	require.NoError(t, err)

	payload2 := `{"peso":68.0,"porcentaje_grasa":21.5}`
	req2 := httptest.NewRequest(http.MethodPost, "/pacientes/"+sinAltura.ID+"/controles", strings.NewReader(payload2))
	req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusCreated, rec2.Code)
	var ctrl2 models.ControlClinico
	require.NoError(t, json.Unmarshal(rec2.Body.Bytes(), &ctrl2))
	assert.Equal(t, 0.0, ctrl2.IMC)
}

// TestCreateControlPacienteValidaciones verifica table-driven los rechazos 400 ante peso o grasa inválidos o JSON malformado.
func TestCreateControlPacienteValidaciones(t *testing.T) {
	casos := []struct {
		name          string
		payload       string
		expectedError string
	}{
		{name: "peso cero", payload: `{"peso":0,"porcentaje_grasa":20}`, expectedError: "peso"},
		{name: "peso mayor a 500", payload: `{"peso":501,"porcentaje_grasa":20}`, expectedError: "peso"},
		{name: "porcentaje_grasa negativo", payload: `{"peso":70,"porcentaje_grasa":-1}`, expectedError: "porcentaje_grasa"},
		{name: "porcentaje_grasa mayor a 100", payload: `{"peso":70,"porcentaje_grasa":105}`, expectedError: "porcentaje_grasa"},
		{name: "json malformado", payload: `{"peso":"invalido"}`, expectedError: "Invalid request"},
	}

	for _, tc := range casos {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			repo := firestore.NewMockPacienteRepository()
			SetupRoutes(e, repo)

			creado, err := repo.Create(context.Background(), models.Paciente{Nombre: "Val Test", Edad: 30, Peso: 70})
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/pacientes/"+creado.ID+"/controles", strings.NewReader(tc.payload))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Contains(t, rec.Body.String(), tc.expectedError)
		})
	}
}

// TestGetControlesPacienteOK verifica obtención cronológica del historial y lista vacía inicial.
func TestGetControlesPacienteOK(t *testing.T) {
	e := echo.New()
	repo := firestore.NewMockPacienteRepository()
	SetupRoutes(e, repo)

	creado, err := repo.Create(context.Background(), models.Paciente{
		Nombre: "Historial Test",
		Edad:   29,
		Peso:   80.0,
		Altura: 1.80,
	})
	require.NoError(t, err)

	// Lista vacía inicial
	reqEmpty := httptest.NewRequest(http.MethodGet, "/pacientes/"+creado.ID+"/controles", nil)
	recEmpty := httptest.NewRecorder()
	e.ServeHTTP(recEmpty, reqEmpty)
	assert.Equal(t, http.StatusOK, recEmpty.Code)
	var vacio []models.ControlClinico
	require.NoError(t, json.Unmarshal(recEmpty.Body.Bytes(), &vacio))
	assert.Len(t, vacio, 0)

	// Registrar dos controles con fechas explícitas en orden inverso
	f2 := time.Date(2026, 11, 15, 9, 0, 0, 0, time.UTC)
	f1 := time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC)

	reqP2 := httptest.NewRequest(http.MethodPost, "/pacientes/"+creado.ID+"/controles", strings.NewReader(`{"fecha":"2026-11-15T09:00:00Z","peso":77.0,"porcentaje_grasa":18.5}`))
	reqP2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recP2 := httptest.NewRecorder()
	e.ServeHTTP(recP2, reqP2)
	require.Equal(t, http.StatusCreated, recP2.Code)

	reqP1 := httptest.NewRequest(http.MethodPost, "/pacientes/"+creado.ID+"/controles", strings.NewReader(`{"fecha":"2026-10-15T09:00:00Z","peso":80.0,"porcentaje_grasa":20.0}`))
	reqP1.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recP1 := httptest.NewRecorder()
	e.ServeHTTP(recP1, reqP1)
	require.Equal(t, http.StatusCreated, recP1.Code)

	reqGet := httptest.NewRequest(http.MethodGet, "/pacientes/"+creado.ID+"/controles", nil)
	recGet := httptest.NewRecorder()
	e.ServeHTTP(recGet, reqGet)
	assert.Equal(t, http.StatusOK, recGet.Code)

	var historial []models.ControlClinico
	require.NoError(t, json.Unmarshal(recGet.Body.Bytes(), &historial))
	require.Len(t, historial, 2)
	assert.True(t, historial[0].Fecha.Equal(f1))
	assert.Equal(t, 80.0, historial[0].Peso)
	assert.True(t, historial[1].Fecha.Equal(f2))
	assert.Equal(t, 77.0, historial[1].Peso)
}

// TestControlesPacienteNoEncontradoYSoftDeleted verifica 404 al consultar o crear controles para paciente inexistente o eliminado lógicamente.
func TestControlesPacienteNoEncontradoYSoftDeleted(t *testing.T) {
	e := echo.New()
	repo := firestore.NewMockPacienteRepository()
	SetupRoutes(e, repo)

	creado, err := repo.Create(context.Background(), models.Paciente{Nombre: "Soft Del", Edad: 30, Peso: 70})
	require.NoError(t, err)
	require.NoError(t, repo.Delete(context.Background(), creado.ID))

	for _, id := range []string{"no-existe", creado.ID} {
		reqGet := httptest.NewRequest(http.MethodGet, "/pacientes/"+id+"/controles", nil)
		recGet := httptest.NewRecorder()
		e.ServeHTTP(recGet, reqGet)
		assert.Equal(t, http.StatusNotFound, recGet.Code)

		reqPost := httptest.NewRequest(http.MethodPost, "/pacientes/"+id+"/controles", strings.NewReader(`{"peso":70,"porcentaje_grasa":18}`))
		reqPost.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		recPost := httptest.NewRecorder()
		e.ServeHTTP(recPost, reqPost)
		assert.Equal(t, http.StatusNotFound, recPost.Code)
	}
}

// TestGetResumenControlesPacienteOK verifica el cálculo evolutivo entre el primer y último control clínico.
func TestGetResumenControlesPacienteOK(t *testing.T) {
	e := echo.New()
	repo := firestore.NewMockPacienteRepository()
	SetupRoutes(e, repo)

	creado, err := repo.Create(context.Background(), models.Paciente{
		Nombre: "Evolucion Test",
		Edad:   31,
		Peso:   80.0,
		Altura: 1.80,
	})
	require.NoError(t, err)

	f1 := time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC)
	f2 := time.Date(2026, 10, 30, 9, 0, 0, 0, time.UTC)
	f3 := time.Date(2026, 11, 15, 9, 0, 0, 0, time.UTC)

	// Insertar fuera de orden cronológico para verificar que usa orden temporal real
	payloads := []string{
		`{"fecha":"2026-11-15T09:00:00Z","peso":76.0,"porcentaje_grasa":17.5}`,
		`{"fecha":"2026-10-15T09:00:00Z","peso":80.0,"porcentaje_grasa":20.0}`,
		`{"fecha":"2026-10-30T09:00:00Z","peso":78.2,"porcentaje_grasa":18.8}`,
	}
	for _, p := range payloads {
		req := httptest.NewRequest(http.MethodPost, "/pacientes/"+creado.ID+"/controles", strings.NewReader(p))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)
	}
	_ = f2

	req := httptest.NewRequest(http.MethodGet, "/pacientes/"+creado.ID+"/controles/resumen", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resumen ResumenControles
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resumen))
	assert.Equal(t, creado.ID, resumen.PacienteID)
	assert.Equal(t, 3, resumen.TotalControles)
	require.NotNil(t, resumen.FechaInicio)
	require.NotNil(t, resumen.FechaFin)
	assert.True(t, resumen.FechaInicio.Equal(f1))
	assert.True(t, resumen.FechaFin.Equal(f3))
	assert.Equal(t, 80.0, resumen.PesoInicial)
	assert.Equal(t, 76.0, resumen.PesoActual)
	assert.Equal(t, -4.0, resumen.DeltaPeso)
	assert.Equal(t, 20.0, resumen.PorcentajeGrasaInicial)
	assert.Equal(t, 17.5, resumen.PorcentajeGrasaActual)
	assert.Equal(t, -2.5, resumen.DeltaPorcentajeGrasa)
	assert.Equal(t, 24.69, resumen.IMCInicial)
	assert.Equal(t, 23.46, resumen.IMCActual)
	assert.Equal(t, -1.23, resumen.DeltaIMC)
}

// TestGetResumenControlesPacienteSinControlesYUnControl verifica casos borde con 0 y 1 control (sin altura).
func TestGetResumenControlesPacienteSinControlesYUnControl(t *testing.T) {
	e := echo.New()
	repo := firestore.NewMockPacienteRepository()
	SetupRoutes(e, repo)

	creado, err := repo.Create(context.Background(), models.Paciente{
		Nombre: "Borde Controles",
		Edad:   28,
		Peso:   65.0,
	})
	require.NoError(t, err)

	t.Run("sin controles", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/pacientes/"+creado.ID+"/controles/resumen", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resumen ResumenControles
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resumen))
		assert.Equal(t, creado.ID, resumen.PacienteID)
		assert.Equal(t, 0, resumen.TotalControles)
		assert.Nil(t, resumen.FechaInicio)
		assert.Nil(t, resumen.FechaFin)
		assert.Equal(t, 0.0, resumen.DeltaPeso)
		assert.Equal(t, 0.0, resumen.DeltaIMC)
	})

	t.Run("un control sin altura", func(t *testing.T) {
		reqPost := httptest.NewRequest(http.MethodPost, "/pacientes/"+creado.ID+"/controles", strings.NewReader(`{"fecha":"2026-11-01T10:00:00Z","peso":64.5,"porcentaje_grasa":21.0}`))
		reqPost.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		recPost := httptest.NewRecorder()
		e.ServeHTTP(recPost, reqPost)
		require.Equal(t, http.StatusCreated, recPost.Code)

		req := httptest.NewRequest(http.MethodGet, "/pacientes/"+creado.ID+"/controles/resumen", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resumen ResumenControles
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resumen))
		assert.Equal(t, 1, resumen.TotalControles)
		require.NotNil(t, resumen.FechaInicio)
		require.NotNil(t, resumen.FechaFin)
		assert.True(t, resumen.FechaInicio.Equal(*resumen.FechaFin))
		assert.Equal(t, 64.5, resumen.PesoInicial)
		assert.Equal(t, 64.5, resumen.PesoActual)
		assert.Equal(t, 0.0, resumen.DeltaPeso)
		assert.Equal(t, 0.0, resumen.DeltaPorcentajeGrasa)
		assert.Equal(t, 0.0, resumen.IMCInicial)
		assert.Equal(t, 0.0, resumen.IMCActual)
		assert.Equal(t, 0.0, resumen.DeltaIMC)
	})
}

// TestGetResumenControlesPacienteNoEncontradoYSoftDeleted verifica 404 para paciente inexistente o eliminado lógicamente.
func TestGetResumenControlesPacienteNoEncontradoYSoftDeleted(t *testing.T) {
	e := echo.New()
	repo := firestore.NewMockPacienteRepository()
	SetupRoutes(e, repo)

	creado, err := repo.Create(context.Background(), models.Paciente{Nombre: "Soft Del Resumen", Edad: 30, Peso: 70})
	require.NoError(t, err)
	require.NoError(t, repo.Delete(context.Background(), creado.ID))

	for _, id := range []string{"no-existe", creado.ID} {
		req := httptest.NewRequest(http.MethodGet, "/pacientes/"+id+"/controles/resumen", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}
}

// TestGetControlesPacienteFiltroFechas verifica el filtrado inclusivo por rango de fechas (YYYY-MM-DD y RFC3339).
func TestGetControlesPacienteFiltroFechas(t *testing.T) {
	e := echo.New()
	repo := firestore.NewMockPacienteRepository()
	SetupRoutes(e, repo)

	creado, err := repo.Create(context.Background(), models.Paciente{Nombre: "Filtro Fechas", Edad: 30, Peso: 75})
	require.NoError(t, err)

	fechas := []string{
		"2026-09-10T09:00:00Z",
		"2026-10-15T14:30:00Z",
		"2026-11-01T10:00:00Z",
		"2026-11-20T18:00:00Z",
	}
	for i, f := range fechas {
		parsed, parseErr := time.Parse(time.RFC3339, f)
		require.NoError(t, parseErr)
		_, addErr := repo.AddControl(context.Background(), creado.ID, models.ControlClinico{
			Fecha:           parsed,
			Peso:            75.0 - float64(i),
			PorcentajeGrasa: 20.0,
		})
		require.NoError(t, addErr)
	}

	casos := []struct {
		name          string
		query         string
		expectedCount int
		firstPeso     float64
	}{
		{
			name:          "rango YYYY-MM-DD inclusivo hasta fin de dia",
			query:         "?desde=2026-10-15&hasta=2026-11-01",
			expectedCount: 2,
			firstPeso:     74.0,
		},
		{
			name:          "desde RFC3339 exacto",
			query:         "?desde=2026-10-15T15:00:00Z",
			expectedCount: 2,
			firstPeso:     73.0,
		},
		{
			name:          "solo hasta YYYY-MM-DD",
			query:         "?hasta=2026-10-15",
			expectedCount: 2,
			firstPeso:     75.0,
		},
		{
			name:          "rango sin coincidencias",
			query:         "?desde=2026-12-01",
			expectedCount: 0,
		},
	}

	for _, tc := range casos {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/pacientes/"+creado.ID+"/controles"+tc.query, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			var res []models.ControlClinico
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
			assert.Len(t, res, tc.expectedCount)
			if tc.expectedCount > 0 {
				assert.Equal(t, tc.firstPeso, res[0].Peso)
			}
		})
	}
}

// TestGetControlesPacienteFiltroFechasInvalidas verifica rechazo 400 ante formatos inválidos o rango invertido.
func TestGetControlesPacienteFiltroFechasInvalidas(t *testing.T) {
	casos := []struct {
		name          string
		query         string
		expectedError string
	}{
		{name: "desde invalido", query: "?desde=no-es-fecha", expectedError: "desde"},
		{name: "hasta invalido", query: "?hasta=15/11/2026", expectedError: "hasta"},
		{name: "desde posterior a hasta", query: "?desde=2026-11-20&hasta=2026-10-01", expectedError: "posterior"},
	}

	for _, tc := range casos {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			repo := firestore.NewMockPacienteRepository()
			SetupRoutes(e, repo)

			creado, err := repo.Create(context.Background(), models.Paciente{Nombre: "Fechas Inv", Edad: 30, Peso: 70})
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, "/pacientes/"+creado.ID+"/controles"+tc.query, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Contains(t, rec.Body.String(), tc.expectedError)
		})
	}
}

// TestGetControlesPacientePaginacion verifica limit/offset (default 10) y combinación con filtro de fechas.
func TestGetControlesPacientePaginacion(t *testing.T) {
	e := echo.New()
	repo := firestore.NewMockPacienteRepository()
	SetupRoutes(e, repo)

	creado, err := repo.Create(context.Background(), models.Paciente{Nombre: "Pag Controles", Edad: 30, Peso: 80})
	require.NoError(t, err)

	base := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	for i := 0; i < 12; i++ {
		_, addErr := repo.AddControl(context.Background(), creado.ID, models.ControlClinico{
			Fecha:           base.AddDate(0, 0, i),
			Peso:            80.0 - float64(i),
			PorcentajeGrasa: 20.0,
		})
		require.NoError(t, addErr)
	}

	t.Run("default limit 10", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/pacientes/"+creado.ID+"/controles", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var res []models.ControlClinico
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
		assert.Len(t, res, 10)
		assert.Equal(t, 80.0, res[0].Peso)
	})

	t.Run("limit y offset", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/pacientes/"+creado.ID+"/controles?limit=3&offset=2", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var res []models.ControlClinico
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
		require.Len(t, res, 3)
		assert.Equal(t, 78.0, res[0].Peso)
		assert.Equal(t, 76.0, res[2].Peso)
	})

	t.Run("offset supera total", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/pacientes/"+creado.ID+"/controles?limit=5&offset=100", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var res []models.ControlClinico
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
		assert.Len(t, res, 0)
	})

	t.Run("combinado filtro fechas y paginacion", func(t *testing.T) {
		// Desde 2026-11-05 (indice 4 -> pesos 76..69, 8 controles); offset=1, limit=2 -> [75, 74]
		req := httptest.NewRequest(http.MethodGet, "/pacientes/"+creado.ID+"/controles?desde=2026-11-05&limit=2&offset=1", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var res []models.ControlClinico
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
		require.Len(t, res, 2)
		assert.Equal(t, 75.0, res[0].Peso)
		assert.Equal(t, 74.0, res[1].Peso)
	})
}

// TestExportControlesPacienteCSVOK verifica la exportación CSV RFC 4180 del historial cronológico de controles de un paciente.
func TestExportControlesPacienteCSVOK(t *testing.T) {
	e := echo.New()
	repo := firestore.NewMockPacienteRepository()
	SetupRoutes(e, repo)

	creado, err := repo.Create(context.Background(), models.Paciente{
		Nombre: "CSV Controles",
		Edad:   30,
		Peso:   80.0,
		Altura: 1.80,
	})
	require.NoError(t, err)

	f2 := time.Date(2026, 11, 15, 9, 0, 0, 0, time.UTC)
	f1 := time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC)
	_, err = repo.AddControl(context.Background(), creado.ID, models.ControlClinico{
		Fecha: f2, Peso: 77.0, PorcentajeGrasa: 18.5, IMC: 23.77, Notas: "Segundo control, buen avance",
	})
	require.NoError(t, err)
	_, err = repo.AddControl(context.Background(), creado.ID, models.ControlClinico{
		Fecha: f1, Peso: 80.0, PorcentajeGrasa: 20.0, IMC: 24.69, Notas: "Inicio",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/pacientes/"+creado.ID+"/controles/export", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get(echo.HeaderContentType), "text/csv")
	assert.Equal(t, `attachment; filename="controles.csv"`, rec.Header().Get(echo.HeaderContentDisposition))

	records, err := csv.NewReader(rec.Body).ReadAll()
	require.NoError(t, err)
	require.Len(t, records, 3)
	assert.Equal(t, []string{"id", "paciente_id", "fecha", "peso", "porcentaje_grasa", "imc", "notas"}, records[0])

	// Orden cronológico: f1 primero, f2 segundo
	assert.Equal(t, creado.ID, records[1][1])
	assert.Equal(t, "2026-10-15T09:00:00Z", records[1][2])
	assert.Equal(t, "80", records[1][3])
	assert.Equal(t, "20", records[1][4])
	assert.Equal(t, "24.69", records[1][5])
	assert.Equal(t, "Inicio", records[1][6])

	assert.Equal(t, "2026-11-15T09:00:00Z", records[2][2])
	assert.Equal(t, "77", records[2][3])
	assert.Equal(t, "18.5", records[2][4])
	assert.Equal(t, "23.77", records[2][5])
	assert.Equal(t, "Segundo control, buen avance", records[2][6])
}

// TestExportControlesPacienteCSVFiltroFechasYVacios verifica exportación sin controles, filtro por fechas y error 400 ante fechas inválidas.
func TestExportControlesPacienteCSVFiltroFechasYVacios(t *testing.T) {
	e := echo.New()
	repo := firestore.NewMockPacienteRepository()
	SetupRoutes(e, repo)

	creado, err := repo.Create(context.Background(), models.Paciente{Nombre: "CSV Filtro", Edad: 29, Peso: 70})
	require.NoError(t, err)

	t.Run("sin controles solo cabecera", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/pacientes/"+creado.ID+"/controles/export", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		records, readErr := csv.NewReader(rec.Body).ReadAll()
		require.NoError(t, readErr)
		require.Len(t, records, 1)
		assert.Equal(t, []string{"id", "paciente_id", "fecha", "peso", "porcentaje_grasa", "imc", "notas"}, records[0])
	})

	f1 := time.Date(2026, 10, 10, 10, 0, 0, 0, time.UTC)
	f2 := time.Date(2026, 11, 10, 10, 0, 0, 0, time.UTC)
	_, err = repo.AddControl(context.Background(), creado.ID, models.ControlClinico{Fecha: f1, Peso: 70, PorcentajeGrasa: 20})
	require.NoError(t, err)
	_, err = repo.AddControl(context.Background(), creado.ID, models.ControlClinico{Fecha: f2, Peso: 68, PorcentajeGrasa: 19})
	require.NoError(t, err)

	t.Run("con filtro de fechas", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/pacientes/"+creado.ID+"/controles/export?desde=2026-11-01", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		records, readErr := csv.NewReader(rec.Body).ReadAll()
		require.NoError(t, readErr)
		require.Len(t, records, 2)
		assert.Equal(t, "2026-11-10T10:00:00Z", records[1][2])
		assert.Equal(t, "", records[1][5])
	})

	t.Run("fecha invalida devuelve 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/pacientes/"+creado.ID+"/controles/export?desde=invalida", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

// TestExportControlesPacienteCSVNoEncontradoYSoftDeleted verifica 404 para paciente inexistente o eliminado lógicamente.
func TestExportControlesPacienteCSVNoEncontradoYSoftDeleted(t *testing.T) {
	e := echo.New()
	repo := firestore.NewMockPacienteRepository()
	SetupRoutes(e, repo)

	creado, err := repo.Create(context.Background(), models.Paciente{Nombre: "Soft Del CSV", Edad: 30, Peso: 70})
	require.NoError(t, err)
	require.NoError(t, repo.Delete(context.Background(), creado.ID))

	for _, id := range []string{"no-existe", creado.ID} {
		req := httptest.NewRequest(http.MethodGet, "/pacientes/"+id+"/controles/export", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}
}

// TestUpdateControlPacienteOK verifica edición de un control clínico puntual, recálculo automático de IMC
// y preservación de fecha cuando se omite en el payload.
func TestUpdateControlPacienteOK(t *testing.T) {
	t.Run("actualizacion completa con nueva fecha y recalculo de IMC", func(t *testing.T) {
		e := echo.New()
		repo := firestore.NewMockPacienteRepository()
		SetupRoutes(e, repo)

		paciente, err := repo.Create(context.Background(), models.Paciente{
			Nombre: "Editar Control",
			Edad:   30,
			Peso:   80.0,
			Altura: 1.80,
		})
		require.NoError(t, err)

		fOrig := time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC)
		ctrl, err := repo.AddControl(context.Background(), paciente.ID, models.ControlClinico{
			Fecha:           fOrig,
			Peso:            80.0,
			PorcentajeGrasa: 20.0,
			IMC:             24.69,
			Notas:           "Inicial",
		})
		require.NoError(t, err)

		payload := `{"fecha":"2026-10-16T11:30:00Z","peso":76.0,"porcentaje_grasa":18.5,"notas":"Corregido"}`
		req := httptest.NewRequest(http.MethodPut, "/pacientes/"+paciente.ID+"/controles/"+ctrl.ID, strings.NewReader(payload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var actualizado models.ControlClinico
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &actualizado))
		assert.Equal(t, ctrl.ID, actualizado.ID)
		assert.Equal(t, paciente.ID, actualizado.PacienteID)
		assert.True(t, actualizado.Fecha.Equal(time.Date(2026, 10, 16, 11, 30, 0, 0, time.UTC)))
		assert.Equal(t, 76.0, actualizado.Peso)
		assert.Equal(t, 18.5, actualizado.PorcentajeGrasa)
		assert.Equal(t, 23.46, actualizado.IMC)
		assert.Equal(t, "Corregido", actualizado.Notas)

		// Verificar persistencia reflejada en GET /controles
		reqGet := httptest.NewRequest(http.MethodGet, "/pacientes/"+paciente.ID+"/controles", nil)
		recGet := httptest.NewRecorder()
		e.ServeHTTP(recGet, reqGet)
		require.Equal(t, http.StatusOK, recGet.Code)

		var lista []models.ControlClinico
		require.NoError(t, json.Unmarshal(recGet.Body.Bytes(), &lista))
		require.Len(t, lista, 1)
		assert.Equal(t, 76.0, lista[0].Peso)
		assert.Equal(t, 23.46, lista[0].IMC)
		assert.Equal(t, "Corregido", lista[0].Notas)
	})

	t.Run("omite fecha y preserva fecha original", func(t *testing.T) {
		e := echo.New()
		repo := firestore.NewMockPacienteRepository()
		SetupRoutes(e, repo)

		paciente, err := repo.Create(context.Background(), models.Paciente{
			Nombre: "Sin Altura Control",
			Edad:   28,
			Peso:   70.0,
		})
		require.NoError(t, err)

		fOrig := time.Date(2026, 10, 20, 8, 0, 0, 0, time.UTC)
		ctrl, err := repo.AddControl(context.Background(), paciente.ID, models.ControlClinico{
			Fecha:           fOrig,
			Peso:            70.0,
			PorcentajeGrasa: 21.0,
			Notas:           "Original",
		})
		require.NoError(t, err)

		payload := `{"peso":69.0,"porcentaje_grasa":20.5,"notas":"Ajuste peso"}`
		req := httptest.NewRequest(http.MethodPut, "/pacientes/"+paciente.ID+"/controles/"+ctrl.ID, strings.NewReader(payload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var actualizado models.ControlClinico
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &actualizado))
		assert.True(t, actualizado.Fecha.Equal(fOrig))
		assert.Equal(t, 69.0, actualizado.Peso)
		assert.Equal(t, 20.5, actualizado.PorcentajeGrasa)
		assert.Equal(t, 0.0, actualizado.IMC)
		assert.Equal(t, "Ajuste peso", actualizado.Notas)
	})
}

// TestUpdateControlPacienteErrores verifica table-driven los rechazos 400 (validación/JSON) y 404 (inexistentes, soft delete o aislamiento).
func TestUpdateControlPacienteErrores(t *testing.T) {
	casos := []struct {
		name           string
		setup          func(t *testing.T, repo *firestore.MockPacienteRepository) (pacienteID, controlID string)
		payload        string
		expectedStatus int
		expectedError  string
	}{
		{
			name: "peso cero",
			setup: func(t *testing.T, repo *firestore.MockPacienteRepository) (string, string) {
				p, err := repo.Create(context.Background(), models.Paciente{Nombre: "Val", Edad: 30, Peso: 70})
				require.NoError(t, err)
				c, err := repo.AddControl(context.Background(), p.ID, models.ControlClinico{Fecha: time.Now().UTC(), Peso: 70, PorcentajeGrasa: 20})
				require.NoError(t, err)
				return p.ID, c.ID
			},
			payload:        `{"peso":0,"porcentaje_grasa":20}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "peso",
		},
		{
			name: "peso mayor a 500",
			setup: func(t *testing.T, repo *firestore.MockPacienteRepository) (string, string) {
				p, err := repo.Create(context.Background(), models.Paciente{Nombre: "Val", Edad: 30, Peso: 70})
				require.NoError(t, err)
				c, err := repo.AddControl(context.Background(), p.ID, models.ControlClinico{Fecha: time.Now().UTC(), Peso: 70, PorcentajeGrasa: 20})
				require.NoError(t, err)
				return p.ID, c.ID
			},
			payload:        `{"peso":501,"porcentaje_grasa":20}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "peso",
		},
		{
			name: "porcentaje grasa negativo",
			setup: func(t *testing.T, repo *firestore.MockPacienteRepository) (string, string) {
				p, err := repo.Create(context.Background(), models.Paciente{Nombre: "Val", Edad: 30, Peso: 70})
				require.NoError(t, err)
				c, err := repo.AddControl(context.Background(), p.ID, models.ControlClinico{Fecha: time.Now().UTC(), Peso: 70, PorcentajeGrasa: 20})
				require.NoError(t, err)
				return p.ID, c.ID
			},
			payload:        `{"peso":70,"porcentaje_grasa":-1}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "porcentaje_grasa",
		},
		{
			name: "porcentaje grasa mayor a 100",
			setup: func(t *testing.T, repo *firestore.MockPacienteRepository) (string, string) {
				p, err := repo.Create(context.Background(), models.Paciente{Nombre: "Val", Edad: 30, Peso: 70})
				require.NoError(t, err)
				c, err := repo.AddControl(context.Background(), p.ID, models.ControlClinico{Fecha: time.Now().UTC(), Peso: 70, PorcentajeGrasa: 20})
				require.NoError(t, err)
				return p.ID, c.ID
			},
			payload:        `{"peso":70,"porcentaje_grasa":101}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "porcentaje_grasa",
		},
		{
			name: "json malformado",
			setup: func(t *testing.T, repo *firestore.MockPacienteRepository) (string, string) {
				p, err := repo.Create(context.Background(), models.Paciente{Nombre: "Val", Edad: 30, Peso: 70})
				require.NoError(t, err)
				c, err := repo.AddControl(context.Background(), p.ID, models.ControlClinico{Fecha: time.Now().UTC(), Peso: 70, PorcentajeGrasa: 20})
				require.NoError(t, err)
				return p.ID, c.ID
			},
			payload:        `{"peso":"invalido"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request",
		},
		{
			name: "paciente inexistente",
			setup: func(t *testing.T, repo *firestore.MockPacienteRepository) (string, string) {
				return "no-existe", "control-1"
			},
			payload:        `{"peso":70,"porcentaje_grasa":20}`,
			expectedStatus: http.StatusNotFound,
			expectedError:  "paciente",
		},
		{
			name: "paciente soft deleted",
			setup: func(t *testing.T, repo *firestore.MockPacienteRepository) (string, string) {
				p, err := repo.Create(context.Background(), models.Paciente{Nombre: "Soft Del", Edad: 30, Peso: 70})
				require.NoError(t, err)
				c, err := repo.AddControl(context.Background(), p.ID, models.ControlClinico{Fecha: time.Now().UTC(), Peso: 70, PorcentajeGrasa: 20})
				require.NoError(t, err)
				require.NoError(t, repo.Delete(context.Background(), p.ID))
				return p.ID, c.ID
			},
			payload:        `{"peso":71,"porcentaje_grasa":20}`,
			expectedStatus: http.StatusNotFound,
			expectedError:  "paciente",
		},
		{
			name: "control inexistente",
			setup: func(t *testing.T, repo *firestore.MockPacienteRepository) (string, string) {
				p, err := repo.Create(context.Background(), models.Paciente{Nombre: "Sin Control", Edad: 30, Peso: 70})
				require.NoError(t, err)
				return p.ID, "control-no-existe"
			},
			payload:        `{"peso":71,"porcentaje_grasa":20}`,
			expectedStatus: http.StatusNotFound,
			expectedError:  "control",
		},
		{
			name: "control de otro paciente no accesible",
			setup: func(t *testing.T, repo *firestore.MockPacienteRepository) (string, string) {
				p1, err := repo.Create(context.Background(), models.Paciente{Nombre: "Paciente Uno", Edad: 30, Peso: 70})
				require.NoError(t, err)
				p2, err := repo.Create(context.Background(), models.Paciente{Nombre: "Paciente Dos", Edad: 31, Peso: 75})
				require.NoError(t, err)
				c1, err := repo.AddControl(context.Background(), p1.ID, models.ControlClinico{Fecha: time.Now().UTC(), Peso: 70, PorcentajeGrasa: 20})
				require.NoError(t, err)
				return p2.ID, c1.ID
			},
			payload:        `{"peso":71,"porcentaje_grasa":20}`,
			expectedStatus: http.StatusNotFound,
			expectedError:  "control",
		},
	}

	for _, tc := range casos {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			repo := firestore.NewMockPacienteRepository()
			SetupRoutes(e, repo)

			pacienteID, controlID := tc.setup(t, repo)
			req := httptest.NewRequest(http.MethodPut, "/pacientes/"+pacienteID+"/controles/"+controlID, strings.NewReader(tc.payload))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tc.expectedError)
		})
	}
}

// TestDeleteControlPacienteOK verifica eliminación puntual de un control clínico y su efecto en historial y resumen.
func TestDeleteControlPacienteOK(t *testing.T) {
	e := echo.New()
	repo := firestore.NewMockPacienteRepository()
	SetupRoutes(e, repo)

	paciente, err := repo.Create(context.Background(), models.Paciente{
		Nombre: "Borrar Control",
		Edad:   33,
		Peso:   82.0,
		Altura: 1.80,
	})
	require.NoError(t, err)

	f1 := time.Date(2026, 10, 10, 10, 0, 0, 0, time.UTC)
	f2 := time.Date(2026, 11, 10, 10, 0, 0, 0, time.UTC)
	c1, err := repo.AddControl(context.Background(), paciente.ID, models.ControlClinico{Fecha: f1, Peso: 82.0, PorcentajeGrasa: 22.0, IMC: 25.31})
	require.NoError(t, err)
	c2, err := repo.AddControl(context.Background(), paciente.ID, models.ControlClinico{Fecha: f2, Peso: 79.0, PorcentajeGrasa: 20.0, IMC: 24.38})
	require.NoError(t, err)

	reqDel := httptest.NewRequest(http.MethodDelete, "/pacientes/"+paciente.ID+"/controles/"+c1.ID, nil)
	recDel := httptest.NewRecorder()
	e.ServeHTTP(recDel, reqDel)
	assert.Equal(t, http.StatusNoContent, recDel.Code)

	// Solo queda c2 en el historial
	reqGet := httptest.NewRequest(http.MethodGet, "/pacientes/"+paciente.ID+"/controles", nil)
	recGet := httptest.NewRecorder()
	e.ServeHTTP(recGet, reqGet)
	require.Equal(t, http.StatusOK, recGet.Code)
	var lista []models.ControlClinico
	require.NoError(t, json.Unmarshal(recGet.Body.Bytes(), &lista))
	require.Len(t, lista, 1)
	assert.Equal(t, c2.ID, lista[0].ID)

	// El resumen refleja total_controles == 1
	reqRes := httptest.NewRequest(http.MethodGet, "/pacientes/"+paciente.ID+"/controles/resumen", nil)
	recRes := httptest.NewRecorder()
	e.ServeHTTP(recRes, reqRes)
	require.Equal(t, http.StatusOK, recRes.Code)
	var resumen ResumenControles
	require.NoError(t, json.Unmarshal(recRes.Body.Bytes(), &resumen))
	assert.Equal(t, 1, resumen.TotalControles)
	assert.Equal(t, 79.0, resumen.PesoInicial)
	assert.Equal(t, 79.0, resumen.PesoActual)
}

// TestDeleteControlPacienteErrores verifica 404 ante paciente inexistente/eliminado, control inexistente, doble borrado o control ajeno.
func TestDeleteControlPacienteErrores(t *testing.T) {
	casos := []struct {
		name          string
		setup         func(t *testing.T, repo *firestore.MockPacienteRepository) (pacienteID, controlID string)
		expectedError string
	}{
		{
			name: "paciente inexistente",
			setup: func(t *testing.T, repo *firestore.MockPacienteRepository) (string, string) {
				return "no-existe", "control-1"
			},
			expectedError: "paciente",
		},
		{
			name: "paciente soft deleted",
			setup: func(t *testing.T, repo *firestore.MockPacienteRepository) (string, string) {
				p, err := repo.Create(context.Background(), models.Paciente{Nombre: "Soft Del", Edad: 30, Peso: 70})
				require.NoError(t, err)
				c, err := repo.AddControl(context.Background(), p.ID, models.ControlClinico{Fecha: time.Now().UTC(), Peso: 70, PorcentajeGrasa: 20})
				require.NoError(t, err)
				require.NoError(t, repo.Delete(context.Background(), p.ID))
				return p.ID, c.ID
			},
			expectedError: "paciente",
		},
		{
			name: "control inexistente",
			setup: func(t *testing.T, repo *firestore.MockPacienteRepository) (string, string) {
				p, err := repo.Create(context.Background(), models.Paciente{Nombre: "Activo", Edad: 30, Peso: 70})
				require.NoError(t, err)
				return p.ID, "control-no-existe"
			},
			expectedError: "control",
		},
		{
			name: "control de otro paciente",
			setup: func(t *testing.T, repo *firestore.MockPacienteRepository) (string, string) {
				p1, err := repo.Create(context.Background(), models.Paciente{Nombre: "Uno", Edad: 30, Peso: 70})
				require.NoError(t, err)
				p2, err := repo.Create(context.Background(), models.Paciente{Nombre: "Dos", Edad: 31, Peso: 75})
				require.NoError(t, err)
				c1, err := repo.AddControl(context.Background(), p1.ID, models.ControlClinico{Fecha: time.Now().UTC(), Peso: 70, PorcentajeGrasa: 20})
				require.NoError(t, err)
				return p2.ID, c1.ID
			},
			expectedError: "control",
		},
	}

	for _, tc := range casos {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			repo := firestore.NewMockPacienteRepository()
			SetupRoutes(e, repo)

			pacienteID, controlID := tc.setup(t, repo)
			req := httptest.NewRequest(http.MethodDelete, "/pacientes/"+pacienteID+"/controles/"+controlID, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusNotFound, rec.Code)
			assert.Contains(t, rec.Body.String(), tc.expectedError)
		})
	}

	t.Run("doble eliminacion devuelve 404 en la segunda", func(t *testing.T) {
		e := echo.New()
		repo := firestore.NewMockPacienteRepository()
		SetupRoutes(e, repo)

		p, err := repo.Create(context.Background(), models.Paciente{Nombre: "Doble Del", Edad: 30, Peso: 70})
		require.NoError(t, err)
		c, err := repo.AddControl(context.Background(), p.ID, models.ControlClinico{Fecha: time.Now().UTC(), Peso: 70, PorcentajeGrasa: 20})
		require.NoError(t, err)

		req1 := httptest.NewRequest(http.MethodDelete, "/pacientes/"+p.ID+"/controles/"+c.ID, nil)
		rec1 := httptest.NewRecorder()
		e.ServeHTTP(rec1, req1)
		require.Equal(t, http.StatusNoContent, rec1.Code)

		req2 := httptest.NewRequest(http.MethodDelete, "/pacientes/"+p.ID+"/controles/"+c.ID, nil)
		rec2 := httptest.NewRecorder()
		e.ServeHTTP(rec2, req2)
		assert.Equal(t, http.StatusNotFound, rec2.Code)
	})
}

// TestGetControlPacienteByIDOK verifica consulta puntual de un control clínico y ausencia de colisión con /resumen y /export.
func TestGetControlPacienteByIDOK(t *testing.T) {
	e := echo.New()
	repo := firestore.NewMockPacienteRepository()
	SetupRoutes(e, repo)

	paciente, err := repo.Create(context.Background(), models.Paciente{
		Nombre: "Consulta Control",
		Edad:   30,
		Peso:   80.0,
		Altura: 1.80,
	})
	require.NoError(t, err)

	f1 := time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC)
	ctrl, err := repo.AddControl(context.Background(), paciente.ID, models.ControlClinico{
		Fecha:           f1,
		Peso:            78.5,
		PorcentajeGrasa: 19.0,
		IMC:             24.23,
		Notas:           "Control puntual",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/pacientes/"+paciente.ID+"/controles/"+ctrl.ID, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var obtenido models.ControlClinico
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &obtenido))
	assert.Equal(t, ctrl.ID, obtenido.ID)
	assert.Equal(t, paciente.ID, obtenido.PacienteID)
	assert.True(t, obtenido.Fecha.Equal(f1))
	assert.Equal(t, 78.5, obtenido.Peso)
	assert.Equal(t, 19.0, obtenido.PorcentajeGrasa)
	assert.Equal(t, 24.23, obtenido.IMC)
	assert.Equal(t, "Control puntual", obtenido.Notas)

	// Confirmar que rutas estáticas /resumen y /export siguen resolviendo correctamente
	reqResumen := httptest.NewRequest(http.MethodGet, "/pacientes/"+paciente.ID+"/controles/resumen", nil)
	recResumen := httptest.NewRecorder()
	e.ServeHTTP(recResumen, reqResumen)
	assert.Equal(t, http.StatusOK, recResumen.Code)
	assert.Contains(t, recResumen.Body.String(), `"total_controles":1`)

	reqExport := httptest.NewRequest(http.MethodGet, "/pacientes/"+paciente.ID+"/controles/export", nil)
	recExport := httptest.NewRecorder()
	e.ServeHTTP(recExport, reqExport)
	assert.Equal(t, http.StatusOK, recExport.Code)
	assert.Contains(t, recExport.Header().Get(echo.HeaderContentType), "text/csv")
}

// TestGetControlPacienteByIDErrores verifica 404 ante paciente inexistente, soft-deleted, control inexistente o de otro paciente.
func TestGetControlPacienteByIDErrores(t *testing.T) {
	casos := []struct {
		name          string
		setup         func(t *testing.T, repo *firestore.MockPacienteRepository) (pacienteID, controlID string)
		expectedError string
	}{
		{
			name: "paciente inexistente",
			setup: func(t *testing.T, repo *firestore.MockPacienteRepository) (string, string) {
				return "no-existe", "control-1"
			},
			expectedError: "paciente",
		},
		{
			name: "paciente soft deleted",
			setup: func(t *testing.T, repo *firestore.MockPacienteRepository) (string, string) {
				p, err := repo.Create(context.Background(), models.Paciente{Nombre: "Soft Del", Edad: 30, Peso: 70})
				require.NoError(t, err)
				c, err := repo.AddControl(context.Background(), p.ID, models.ControlClinico{Fecha: time.Now().UTC(), Peso: 70, PorcentajeGrasa: 20})
				require.NoError(t, err)
				require.NoError(t, repo.Delete(context.Background(), p.ID))
				return p.ID, c.ID
			},
			expectedError: "paciente",
		},
		{
			name: "control inexistente",
			setup: func(t *testing.T, repo *firestore.MockPacienteRepository) (string, string) {
				p, err := repo.Create(context.Background(), models.Paciente{Nombre: "Activo", Edad: 30, Peso: 70})
				require.NoError(t, err)
				return p.ID, "control-no-existe"
			},
			expectedError: "control",
		},
		{
			name: "control de otro paciente",
			setup: func(t *testing.T, repo *firestore.MockPacienteRepository) (string, string) {
				p1, err := repo.Create(context.Background(), models.Paciente{Nombre: "Uno", Edad: 30, Peso: 70})
				require.NoError(t, err)
				p2, err := repo.Create(context.Background(), models.Paciente{Nombre: "Dos", Edad: 31, Peso: 75})
				require.NoError(t, err)
				c1, err := repo.AddControl(context.Background(), p1.ID, models.ControlClinico{Fecha: time.Now().UTC(), Peso: 70, PorcentajeGrasa: 20})
				require.NoError(t, err)
				return p2.ID, c1.ID
			},
			expectedError: "control",
		},
	}

	for _, tc := range casos {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			repo := firestore.NewMockPacienteRepository()
			SetupRoutes(e, repo)

			pacienteID, controlID := tc.setup(t, repo)
			req := httptest.NewRequest(http.MethodGet, "/pacientes/"+pacienteID+"/controles/"+controlID, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusNotFound, rec.Code)
			assert.Contains(t, rec.Body.String(), tc.expectedError)
		})
	}
}

// TestPatchControlPacienteOK verifica actualización parcial preservando campos omitidos y recalculando IMC.
func TestPatchControlPacienteOK(t *testing.T) {
	t.Run("solo peso conserva fecha, grasa y notas y recalcula IMC", func(t *testing.T) {
		e := echo.New()
		repo := firestore.NewMockPacienteRepository()
		SetupRoutes(e, repo)

		paciente, err := repo.Create(context.Background(), models.Paciente{
			Nombre: "Patch Control",
			Edad:   30,
			Peso:   80.0,
			Altura: 1.80,
		})
		require.NoError(t, err)

		fOrig := time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC)
		ctrl, err := repo.AddControl(context.Background(), paciente.ID, models.ControlClinico{
			Fecha:           fOrig,
			Peso:            80.0,
			PorcentajeGrasa: 20.0,
			IMC:             24.69,
			Notas:           "Nota original",
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPatch, "/pacientes/"+paciente.ID+"/controles/"+ctrl.ID, strings.NewReader(`{"peso":76.0}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var actualizado models.ControlClinico
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &actualizado))
		assert.Equal(t, ctrl.ID, actualizado.ID)
		assert.Equal(t, paciente.ID, actualizado.PacienteID)
		assert.True(t, actualizado.Fecha.Equal(fOrig))
		assert.Equal(t, 76.0, actualizado.Peso)
		assert.Equal(t, 20.0, actualizado.PorcentajeGrasa)
		assert.Equal(t, 23.46, actualizado.IMC)
		assert.Equal(t, "Nota original", actualizado.Notas)
	})

	t.Run("actualiza fecha RFC3339 y notas sin alterar peso ni grasa", func(t *testing.T) {
		e := echo.New()
		repo := firestore.NewMockPacienteRepository()
		SetupRoutes(e, repo)

		paciente, err := repo.Create(context.Background(), models.Paciente{
			Nombre: "Patch Fecha",
			Edad:   28,
			Peso:   70.0,
		})
		require.NoError(t, err)

		fOrig := time.Date(2026, 10, 10, 9, 0, 0, 0, time.UTC)
		ctrl, err := repo.AddControl(context.Background(), paciente.ID, models.ControlClinico{
			Fecha:           fOrig,
			Peso:            70.0,
			PorcentajeGrasa: 21.5,
			Notas:           "Vieja",
		})
		require.NoError(t, err)

		payload := `{"fecha":"2026-10-12T14:00:00Z","notas":"Actualizada"}`
		req := httptest.NewRequest(http.MethodPatch, "/pacientes/"+paciente.ID+"/controles/"+ctrl.ID, strings.NewReader(payload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var actualizado models.ControlClinico
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &actualizado))
		assert.True(t, actualizado.Fecha.Equal(time.Date(2026, 10, 12, 14, 0, 0, 0, time.UTC)))
		assert.Equal(t, 70.0, actualizado.Peso)
		assert.Equal(t, 21.5, actualizado.PorcentajeGrasa)
		assert.Equal(t, "Actualizada", actualizado.Notas)
		assert.Equal(t, 0.0, actualizado.IMC)
	})
}

// TestPatchControlPacienteErrores verifica table-driven 400 (validaciones/tipos/JSON) y 404 (inexistentes, soft delete, aislamiento).
func TestPatchControlPacienteErrores(t *testing.T) {
	setupActivo := func(t *testing.T, repo *firestore.MockPacienteRepository) (string, string) {
		p, err := repo.Create(context.Background(), models.Paciente{Nombre: "Val Patch", Edad: 30, Peso: 70})
		require.NoError(t, err)
		c, err := repo.AddControl(context.Background(), p.ID, models.ControlClinico{Fecha: time.Now().UTC(), Peso: 70, PorcentajeGrasa: 20})
		require.NoError(t, err)
		return p.ID, c.ID
	}

	casos := []struct {
		name           string
		setup          func(t *testing.T, repo *firestore.MockPacienteRepository) (pacienteID, controlID string)
		payload        string
		expectedStatus int
		expectedError  string
	}{
		{name: "peso cero", setup: setupActivo, payload: `{"peso":0}`, expectedStatus: http.StatusBadRequest, expectedError: "peso"},
		{name: "peso mayor a 500", setup: setupActivo, payload: `{"peso":501}`, expectedStatus: http.StatusBadRequest, expectedError: "peso"},
		{name: "peso tipo invalido", setup: setupActivo, payload: `{"peso":"alto"}`, expectedStatus: http.StatusBadRequest, expectedError: "peso"},
		{name: "porcentaje grasa negativo", setup: setupActivo, payload: `{"porcentaje_grasa":-1}`, expectedStatus: http.StatusBadRequest, expectedError: "porcentaje_grasa"},
		{name: "porcentaje grasa mayor a 100", setup: setupActivo, payload: `{"porcentaje_grasa":105}`, expectedStatus: http.StatusBadRequest, expectedError: "porcentaje_grasa"},
		{name: "porcentaje grasa tipo invalido", setup: setupActivo, payload: `{"porcentaje_grasa":"veinte"}`, expectedStatus: http.StatusBadRequest, expectedError: "porcentaje_grasa"},
		{name: "fecha invalida", setup: setupActivo, payload: `{"fecha":"no-fecha"}`, expectedStatus: http.StatusBadRequest, expectedError: "fecha"},
		{name: "fecha vacia", setup: setupActivo, payload: `{"fecha":""}`, expectedStatus: http.StatusBadRequest, expectedError: "fecha"},
		{name: "notas tipo invalido", setup: setupActivo, payload: `{"notas":123}`, expectedStatus: http.StatusBadRequest, expectedError: "notas"},
		{name: "json malformado", setup: setupActivo, payload: `{invalid`, expectedStatus: http.StatusBadRequest, expectedError: "Invalid request"},
		{
			name: "paciente inexistente",
			setup: func(t *testing.T, repo *firestore.MockPacienteRepository) (string, string) {
				return "no-existe", "control-1"
			},
			payload:        `{"peso":71}`,
			expectedStatus: http.StatusNotFound,
			expectedError:  "paciente",
		},
		{
			name: "paciente soft deleted",
			setup: func(t *testing.T, repo *firestore.MockPacienteRepository) (string, string) {
				p, err := repo.Create(context.Background(), models.Paciente{Nombre: "Soft Del", Edad: 30, Peso: 70})
				require.NoError(t, err)
				c, err := repo.AddControl(context.Background(), p.ID, models.ControlClinico{Fecha: time.Now().UTC(), Peso: 70, PorcentajeGrasa: 20})
				require.NoError(t, err)
				require.NoError(t, repo.Delete(context.Background(), p.ID))
				return p.ID, c.ID
			},
			payload:        `{"peso":71}`,
			expectedStatus: http.StatusNotFound,
			expectedError:  "paciente",
		},
		{
			name: "control inexistente",
			setup: func(t *testing.T, repo *firestore.MockPacienteRepository) (string, string) {
				p, err := repo.Create(context.Background(), models.Paciente{Nombre: "Activo", Edad: 30, Peso: 70})
				require.NoError(t, err)
				return p.ID, "control-no-existe"
			},
			payload:        `{"peso":71}`,
			expectedStatus: http.StatusNotFound,
			expectedError:  "control",
		},
		{
			name: "control de otro paciente",
			setup: func(t *testing.T, repo *firestore.MockPacienteRepository) (string, string) {
				p1, err := repo.Create(context.Background(), models.Paciente{Nombre: "Uno", Edad: 30, Peso: 70})
				require.NoError(t, err)
				p2, err := repo.Create(context.Background(), models.Paciente{Nombre: "Dos", Edad: 31, Peso: 75})
				require.NoError(t, err)
				c1, err := repo.AddControl(context.Background(), p1.ID, models.ControlClinico{Fecha: time.Now().UTC(), Peso: 70, PorcentajeGrasa: 20})
				require.NoError(t, err)
				return p2.ID, c1.ID
			},
			payload:        `{"peso":71}`,
			expectedStatus: http.StatusNotFound,
			expectedError:  "control",
		},
	}

	for _, tc := range casos {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			repo := firestore.NewMockPacienteRepository()
			SetupRoutes(e, repo)

			pacienteID, controlID := tc.setup(t, repo)
			req := httptest.NewRequest(http.MethodPatch, "/pacientes/"+pacienteID+"/controles/"+controlID, strings.NewReader(tc.payload))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tc.expectedError)
		})
	}
}

// TestGetUltimoControlPacienteOK verifica la obtención del control más reciente sin depender del orden de inserción
// y sin colisión de rutas con /controles/:controlId o /controles/resumen.
func TestGetUltimoControlPacienteOK(t *testing.T) {
	e := echo.New()
	repo := firestore.NewMockPacienteRepository()
	SetupRoutes(e, repo)

	paciente, err := repo.Create(context.Background(), models.Paciente{
		Nombre: "Ultimo Control",
		Edad:   32,
		Peso:   80.0,
		Altura: 1.80,
	})
	require.NoError(t, err)

	f1 := time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC)
	f2 := time.Date(2026, 10, 30, 9, 0, 0, 0, time.UTC)
	f3 := time.Date(2026, 11, 15, 9, 0, 0, 0, time.UTC)

	c3, err := repo.AddControl(context.Background(), paciente.ID, models.ControlClinico{
		Fecha: f3, Peso: 76.0, PorcentajeGrasa: 17.5, IMC: 23.46, Notas: "Mas reciente",
	})
	require.NoError(t, err)
	c1, err := repo.AddControl(context.Background(), paciente.ID, models.ControlClinico{
		Fecha: f1, Peso: 80.0, PorcentajeGrasa: 20.0, IMC: 24.69, Notas: "Primero",
	})
	require.NoError(t, err)
	c2, err := repo.AddControl(context.Background(), paciente.ID, models.ControlClinico{
		Fecha: f2, Peso: 78.0, PorcentajeGrasa: 18.8, IMC: 24.07, Notas: "Intermedio",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/pacientes/"+paciente.ID+"/controles/ultimo", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var ultimo models.ControlClinico
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &ultimo))
	assert.Equal(t, c3.ID, ultimo.ID)
	assert.Equal(t, paciente.ID, ultimo.PacienteID)
	assert.True(t, ultimo.Fecha.Equal(f3))
	assert.Equal(t, 76.0, ultimo.Peso)
	assert.Equal(t, 17.5, ultimo.PorcentajeGrasa)
	assert.Equal(t, 23.46, ultimo.IMC)
	assert.Equal(t, "Mas reciente", ultimo.Notas)

	// Si se elimina c3, /ultimo debe devolver c2
	require.NoError(t, repo.DeleteControl(context.Background(), paciente.ID, c3.ID))
	reqPostDel := httptest.NewRequest(http.MethodGet, "/pacientes/"+paciente.ID+"/controles/ultimo", nil)
	recPostDel := httptest.NewRecorder()
	e.ServeHTTP(recPostDel, reqPostDel)
	require.Equal(t, http.StatusOK, recPostDel.Code)
	var nuevoUltimo models.ControlClinico
	require.NoError(t, json.Unmarshal(recPostDel.Body.Bytes(), &nuevoUltimo))
	assert.Equal(t, c2.ID, nuevoUltimo.ID)

	// Confirmar que GET /controles/:controlId sigue resolviendo para c1
	reqByID := httptest.NewRequest(http.MethodGet, "/pacientes/"+paciente.ID+"/controles/"+c1.ID, nil)
	recByID := httptest.NewRecorder()
	e.ServeHTTP(recByID, reqByID)
	assert.Equal(t, http.StatusOK, recByID.Code)
}

// TestGetUltimoControlPacienteErrores verifica table-driven los rechazos 404 cuando no hay controles,
// cuando el paciente no existe o cuando fue dado de baja lógica.
func TestGetUltimoControlPacienteErrores(t *testing.T) {
	casos := []struct {
		name          string
		setup         func(t *testing.T, repo *firestore.MockPacienteRepository) string
		expectedError string
	}{
		{
			name: "paciente sin controles",
			setup: func(t *testing.T, repo *firestore.MockPacienteRepository) string {
				p, err := repo.Create(context.Background(), models.Paciente{Nombre: "Sin Controles", Edad: 29, Peso: 70})
				require.NoError(t, err)
				return p.ID
			},
			expectedError: "controles clínicos",
		},
		{
			name: "paciente inexistente",
			setup: func(t *testing.T, repo *firestore.MockPacienteRepository) string {
				return "no-existe"
			},
			expectedError: "paciente",
		},
		{
			name: "paciente soft deleted con controles",
			setup: func(t *testing.T, repo *firestore.MockPacienteRepository) string {
				p, err := repo.Create(context.Background(), models.Paciente{Nombre: "Soft Del", Edad: 30, Peso: 70})
				require.NoError(t, err)
				_, err = repo.AddControl(context.Background(), p.ID, models.ControlClinico{Fecha: time.Now().UTC(), Peso: 70, PorcentajeGrasa: 20})
				require.NoError(t, err)
				require.NoError(t, repo.Delete(context.Background(), p.ID))
				return p.ID
			},
			expectedError: "paciente",
		},
	}

	for _, tc := range casos {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			repo := firestore.NewMockPacienteRepository()
			SetupRoutes(e, repo)

			pacienteID := tc.setup(t, repo)
			req := httptest.NewRequest(http.MethodGet, "/pacientes/"+pacienteID+"/controles/ultimo", nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusNotFound, rec.Code)
			assert.Contains(t, rec.Body.String(), tc.expectedError)
		})
	}
}

// TestPatchPacienteTiposInvalidos verifica table-driven que tipos JSON erróneos o null devuelvan 400 sin panic ni 200 silencioso.
func TestPatchPacienteTiposInvalidos(t *testing.T) {
	casos := []struct {
		name          string
		payload       string
		expectedError string
	}{
		{name: "body null", payload: `null`, expectedError: "Invalid request"},
		{name: "nombre numerico", payload: `{"nombre":123}`, expectedError: "nombre"},
		{name: "edad string", payload: `{"edad":"treinta"}`, expectedError: "edad"},
		{name: "edad decimal", payload: `{"edad":30.5}`, expectedError: "edad"},
		{name: "peso string", payload: `{"peso":"alto"}`, expectedError: "peso"},
		{name: "porcentaje grasa string", payload: `{"porcentaje_grasa":"veinte"}`, expectedError: "porcentaje_grasa"},
		{name: "altura string", payload: `{"altura":"1.75"}`, expectedError: "altura"},
		{name: "notas numerico", payload: `{"notas":123}`, expectedError: "notas"},
		{name: "notas null", payload: `{"notas":null}`, expectedError: "notas"},
	}

	for _, tc := range casos {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			repo := firestore.NewMockPacienteRepository()
			SetupRoutes(e, repo)

			creado, err := repo.Create(context.Background(), models.Paciente{
				Nombre:          "Tipo Test",
				Edad:            30,
				Peso:            70.0,
				Altura:          1.75,
				PorcentajeGrasa: 20.0,
				Notas:           "Nota previa",
			})
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPatch, "/pacientes/"+creado.ID, strings.NewReader(tc.payload))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			assert.NotPanics(t, func() {
				e.ServeHTTP(rec, req)
			})
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Contains(t, rec.Body.String(), tc.expectedError)
		})
	}
}