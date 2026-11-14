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

// TestExportPacientesCSVOK verifica la exportación de pacientes en formato CSV (cabeceras RFC 4180 y datos).
func TestExportPacientesCSVOK(t *testing.T) {
	// Arrange
	e := newTestEcho()
	repo := firestore.NewMockPacienteRepository()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/pacientes/export", nil)
	c := e.NewContext(req, rec)

	// Act
	err := ExportPacientesCSV(c, repo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get(echo.HeaderContentType), "text/csv")
	assert.Equal(t, `attachment; filename="pacientes.csv"`, rec.Header().Get(echo.HeaderContentDisposition))

	reader := csv.NewReader(rec.Body)
	records, err := reader.ReadAll()
	require.NoError(t, err)

	// Cabecera + 2 pacientes del seed
	require.Len(t, records, 3)
	assert.Equal(t, []string{"id", "nombre", "edad", "peso", "porcentaje_grasa", "notas"}, records[0])

	// Paciente 1 (Juan Perez)
	assert.Equal(t, "Juan Perez", records[1][1])
	assert.Equal(t, "30", records[1][2])
	assert.Equal(t, "75.5", records[1][3])
	assert.Equal(t, "18.5", records[1][4])
	assert.Equal(t, "Primera consulta", records[1][5])

	// Paciente 2 (Maria Gomez)
	assert.Equal(t, "Maria Gomez", records[2][1])
	assert.Equal(t, "25", records[2][2])
	assert.Equal(t, "62", records[2][3])
	assert.Equal(t, "22", records[2][4])
	assert.Equal(t, "Seguimiento mensual", records[2][5])
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
	assert.Equal(t, []string{"id", "nombre", "edad", "peso", "porcentaje_grasa", "notas"}, records[0])
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