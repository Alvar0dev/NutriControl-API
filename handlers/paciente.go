package handlers

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"nutricontrol/models"
	"nutricontrol/pkg/firestore"
)

// paginacionDesdeQuery lee limit y offset de los query params con sus valores por defecto.
func paginacionDesdeQuery(c echo.Context) (limit, offset int) {
	limit = 10
	offset = 0
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	return limit, offset
}

// ordenarPacientes ordena un slice de pacientes en memoria según el campo y sentido solicitados.
// Campos soportados: "nombre", "edad", "peso", "porcentaje_grasa".
// Sentidos soportados: "asc" (por defecto) y "desc".
func ordenarPacientes(pacientes []models.Paciente, sortCampo, orden string) {
	if sortCampo == "" {
		return
	}
	desc := strings.ToLower(orden) == "desc"
	switch strings.ToLower(sortCampo) {
	case "nombre":
		sort.SliceStable(pacientes, func(i, j int) bool {
			if desc {
				return strings.ToLower(pacientes[i].Nombre) > strings.ToLower(pacientes[j].Nombre)
			}
			return strings.ToLower(pacientes[i].Nombre) < strings.ToLower(pacientes[j].Nombre)
		})
	case "edad":
		sort.SliceStable(pacientes, func(i, j int) bool {
			if desc {
				return pacientes[i].Edad > pacientes[j].Edad
			}
			return pacientes[i].Edad < pacientes[j].Edad
		})
	case "peso":
		sort.SliceStable(pacientes, func(i, j int) bool {
			if desc {
				return pacientes[i].Peso > pacientes[j].Peso
			}
			return pacientes[i].Peso < pacientes[j].Peso
		})
	case "porcentaje_grasa":
		sort.SliceStable(pacientes, func(i, j int) bool {
			if desc {
				return pacientes[i].PorcentajeGrasa > pacientes[j].PorcentajeGrasa
			}
			return pacientes[i].PorcentajeGrasa < pacientes[j].PorcentajeGrasa
		})
	case "altura":
		sort.SliceStable(pacientes, func(i, j int) bool {
			if desc {
				return pacientes[i].Altura > pacientes[j].Altura
			}
			return pacientes[i].Altura < pacientes[j].Altura
		})
	}
}

// GetPacientes obtiene todos los pacientes mediante el repositorio.
// Soporta filtro opcional por nombre (?nombre=, coincidencia parcial, sin distinción de mayúsculas),
// ordenamiento (?sort=nombre|edad|peso|porcentaje_grasa&order=asc|desc)
// y paginación nativa del repositorio (?limit, ?offset).
func GetPacientes(c echo.Context, repo firestore.PacienteRepository) error {
	limit, offset := paginacionDesdeQuery(c)
	pacientes, err := repo.GetAll(c.Request().Context(), limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error al obtener pacientes"})
	}
	if nombre := c.QueryParam("nombre"); nombre != "" {
		buscado := strings.ToLower(nombre)
		filtrados := pacientes[:0]
		for _, p := range pacientes {
			if strings.Contains(strings.ToLower(p.Nombre), buscado) {
				filtrados = append(filtrados, p)
			}
		}
		pacientes = filtrados
	}
	ordenarPacientes(pacientes, c.QueryParam("sort"), c.QueryParam("order"))
	return c.JSON(http.StatusOK, pacientes)
}

// SearchPacientes busca pacientes por múltiples criterios con paginación.
// Query params:
//   - nombre: coincidencia parcial, sin distinción de mayúsculas/minúsculas
//   - edad_min, edad_max: rango inclusivo de edad
//   - peso_min, peso_max: rango inclusivo de peso
//   - porcentaje_grasa_min, porcentaje_grasa_max: rango inclusivo de porcentaje de grasa
//   - limit: máximo resultados (default 10)
//   - offset: salto de resultados (default 0)
func SearchPacientes(c echo.Context, repo firestore.PacienteRepository) error {
	limit, offset := paginacionDesdeQuery(c)
	pacientes, err := repo.GetAll(c.Request().Context(), limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error al obtener pacientes"})
	}

	// Filtro por nombre (coincidencia parcial, case-insensitive)
	if nombre := c.QueryParam("nombre"); nombre != "" {
		buscado := strings.ToLower(nombre)
		filtrados := pacientes[:0]
		for _, p := range pacientes {
			if strings.Contains(strings.ToLower(p.Nombre), buscado) {
				filtrados = append(filtrados, p)
			}
		}
		pacientes = filtrados
	}

	// Filtro por rango de edad
	if edadMinStr := c.QueryParam("edad_min"); edadMinStr != "" {
		if edadMin, err := strconv.Atoi(edadMinStr); err == nil {
			filtrados := pacientes[:0]
			for _, p := range pacientes {
				if p.Edad >= edadMin {
					filtrados = append(filtrados, p)
				}
			}
			pacientes = filtrados
		}
	}
	if edadMaxStr := c.QueryParam("edad_max"); edadMaxStr != "" {
		if edadMax, err := strconv.Atoi(edadMaxStr); err == nil {
			filtrados := pacientes[:0]
			for _, p := range pacientes {
				if p.Edad <= edadMax {
					filtrados = append(filtrados, p)
				}
			}
			pacientes = filtrados
		}
	}

	// Filtro por rango de peso
	if pesoMinStr := c.QueryParam("peso_min"); pesoMinStr != "" {
		if pesoMin, err := strconv.ParseFloat(pesoMinStr, 64); err == nil {
			filtrados := pacientes[:0]
			for _, p := range pacientes {
				if p.Peso >= pesoMin {
					filtrados = append(filtrados, p)
				}
			}
			pacientes = filtrados
		}
	}
	if pesoMaxStr := c.QueryParam("peso_max"); pesoMaxStr != "" {
		if pesoMax, err := strconv.ParseFloat(pesoMaxStr, 64); err == nil {
			filtrados := pacientes[:0]
			for _, p := range pacientes {
				if p.Peso <= pesoMax {
					filtrados = append(filtrados, p)
				}
			}
			pacientes = filtrados
		}
	}

	// Filtro por rango de porcentaje de grasa
	if grasaMinStr := c.QueryParam("porcentaje_grasa_min"); grasaMinStr != "" {
		if grasaMin, err := strconv.ParseFloat(grasaMinStr, 64); err == nil {
			filtrados := pacientes[:0]
			for _, p := range pacientes {
				if p.PorcentajeGrasa >= grasaMin {
					filtrados = append(filtrados, p)
				}
			}
			pacientes = filtrados
		}
	}
	if grasaMaxStr := c.QueryParam("porcentaje_grasa_max"); grasaMaxStr != "" {
		if grasaMax, err := strconv.ParseFloat(grasaMaxStr, 64); err == nil {
			filtrados := pacientes[:0]
			for _, p := range pacientes {
				if p.PorcentajeGrasa <= grasaMax {
					filtrados = append(filtrados, p)
				}
			}
			pacientes = filtrados
		}
	}

	ordenarPacientes(pacientes, c.QueryParam("sort"), c.QueryParam("order"))

	return c.JSON(http.StatusOK, pacientes)
}

// StatsPacientes devuelve un resumen de solo lectura de los pacientes:
// total, promedios de edad, peso y porcentaje de grasa, y listas ordenadas de edades y pesos.
func StatsPacientes(c echo.Context, repo firestore.PacienteRepository) error {
	limit, offset := paginacionDesdeQuery(c)
	pacientes, err := repo.GetAll(c.Request().Context(), limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error al obtener pacientes"})
	}

	total := len(pacientes)
	var edadPromedio, pesoPromedio, grasaPromedio float64
	edades := make([]int, 0, total)
	pesos := make([]float64, 0, total)
	for _, p := range pacientes {
		edadPromedio += float64(p.Edad)
		pesoPromedio += p.Peso
		grasaPromedio += p.PorcentajeGrasa
		edades = append(edades, p.Edad)
		pesos = append(pesos, p.Peso)
	}
	if total > 0 {
		edadPromedio /= float64(total)
		pesoPromedio /= float64(total)
		grasaPromedio /= float64(total)
	}
	sort.Ints(edades)
	sort.Float64s(pesos)

	return c.JSON(http.StatusOK, map[string]any{
		"total":                    total,
		"edad_promedio":            edadPromedio,
		"peso_promedio":            pesoPromedio,
		"porcentaje_grasa_promedio": grasaPromedio,
		"edades":                   edades,
		"pesos":                    pesos,
	})
}

const (
	edadMinima            = 1
	edadMaxima            = 120
	pesoMaximo            = 500
	porcentajeGrasaMaximo = 100
	alturaMinima          = 0.30
	alturaMaxima          = 2.50
)

// normalizarNombre limpia espacios sobrantes y deja cada palabra con primera letra mayúscula.
func normalizarNombre(nombre string) string {
	nombre = strings.TrimSpace(nombre)
	if nombre == "" {
		return ""
	}
	palabras := strings.Fields(nombre)
	for i, p := range palabras {
		palabras[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
	}
	return strings.Join(palabras, " ")
}

// validarPaciente verifica las reglas de negocio: nombre requerido, edad entre 1 y 120,
// peso mayor que 0 y menor o igual a 500, y porcentaje_grasa entre 0 y 100.
func validarPaciente(p models.Paciente) error {
	if p.Nombre == "" {
		return fmt.Errorf("El campo 'nombre' es requerido")
	}
	if p.Edad < edadMinima || p.Edad > edadMaxima {
		return fmt.Errorf("El campo 'edad' debe estar entre %d y %d", edadMinima, edadMaxima)
	}
	if p.Peso <= 0 || p.Peso > pesoMaximo {
		return fmt.Errorf("El campo 'peso' debe ser mayor que 0 y menor o igual a %d", pesoMaximo)
	}
	if p.PorcentajeGrasa < 0 || p.PorcentajeGrasa > porcentajeGrasaMaximo {
		return fmt.Errorf("El campo 'porcentaje_grasa' debe estar entre 0 y %d", porcentajeGrasaMaximo)
	}
	if p.Altura != 0 && (p.Altura < alturaMinima || p.Altura > alturaMaxima) {
		return fmt.Errorf("El campo 'altura' debe estar entre %.2f y %.2f metros", alturaMinima, alturaMaxima)
	}
	return nil
}

// CreatePaciente crea un nuevo paciente mediante el repositorio.
// El nombre se normaliza (trim y capitalización) antes de validar y guardar.
func CreatePaciente(c echo.Context, repo firestore.PacienteRepository) error {
	paciente := new(models.Paciente)
	if err := c.Bind(paciente); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	paciente.Nombre = normalizarNombre(paciente.Nombre)
	paciente.DeletedAt = nil
	if err := validarPaciente(*paciente); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	pacienteCreado, err := repo.Create(c.Request().Context(), *paciente)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error al crear paciente"})
	}
	return c.JSON(http.StatusCreated, pacienteCreado)
}

// CreatePacientesBatch crea un lote de nuevos pacientes mediante el repositorio en una operación atómica a nivel de validación.
// Si algún paciente del lote es inválido o el lote está vacío, se rechaza la petición completa (400 Bad Request)
// sin persistir ningún paciente. Si todos son válidos, se crean y devuelven con 201 Created.
func CreatePacientesBatch(c echo.Context, repo firestore.PacienteRepository) error {
	var pacientes []models.Paciente
	if err := c.Bind(&pacientes); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	if len(pacientes) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "El lote de pacientes no puede estar vacío"})
	}

	// Fase de normalización y validación atómica: todos deben ser válidos antes de persistir
	for i := range pacientes {
		pacientes[i].Nombre = normalizarNombre(pacientes[i].Nombre)
		pacientes[i].DeletedAt = nil
		if err := validarPaciente(pacientes[i]); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("Paciente en posición %d inválido: %s", i, err.Error()),
			})
		}
	}

	// Fase de persistencia: todos pasaron validación
	creados := make([]*models.Paciente, 0, len(pacientes))
	for _, p := range pacientes {
		creado, err := repo.Create(c.Request().Context(), p)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error al crear paciente del lote"})
		}
		creados = append(creados, creado)
	}

	return c.JSON(http.StatusCreated, creados)
}

// UpdatePaciente actualiza un paciente existente por ID.
// El nombre se normaliza (trim y capitalización) antes de validar y guardar.
func UpdatePaciente(c echo.Context, repo firestore.PacienteRepository) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "El campo 'id' es requerido"})
	}
	p := new(models.Paciente)
	if err := c.Bind(p); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	p.Nombre = normalizarNombre(p.Nombre)
	p.DeletedAt = nil
	if err := validarPaciente(*p); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	p.ID = id
	pacienteActualizado, err := repo.Update(c.Request().Context(), id, *p)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "No se encontró paciente con ID " + id})
	}
	return c.JSON(http.StatusOK, pacienteActualizado)
}

// PatchPaciente aplica una actualización parcial sobre un paciente existente.
// Solo se modifican los campos enviados en el body; los demás se conservan.
// Body acepta cualquier subconjunto de: nombre, edad, peso, porcentaje_grasa, notas.
func PatchPaciente(c echo.Context, repo firestore.PacienteRepository) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "El campo 'id' es requerido"})
	}

	// Obtener el paciente existente
	existente, err := repo.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "No se encontró paciente con ID " + id})
	}

	// Leer y parsear el body como mapa para aplicar solo los campos enviados
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	var cambios map[string]any
	if err := json.Unmarshal(body, &cambios); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// Aplicar solo los campos que vinieron en el body
	if nombre, ok := cambios["nombre"]; ok {
		if n, ok := nombre.(string); ok {
			existente.Nombre = normalizarNombre(n)
			if existente.Nombre == "" {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "El campo 'nombre' es requerido"})
			}
		}
	}
	if edad, ok := cambios["edad"]; ok {
		if e, ok := edad.(float64); ok {
			if int(e) < edadMinima || int(e) > edadMaxima {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("El campo 'edad' debe estar entre %d y %d", edadMinima, edadMaxima)})
			}
			existente.Edad = int(e)
		}
	}
	if peso, ok := cambios["peso"]; ok {
		if p, ok := peso.(float64); ok {
			if p <= 0 || p > pesoMaximo {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("El campo 'peso' debe ser mayor que 0 y menor o igual a %d", pesoMaximo)})
			}
			existente.Peso = p
		}
	}
	if grasa, ok := cambios["porcentaje_grasa"]; ok {
		if g, ok := grasa.(float64); ok {
			if g < 0 || g > porcentajeGrasaMaximo {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("El campo 'porcentaje_grasa' debe estar entre 0 y %d", porcentajeGrasaMaximo)})
			}
			existente.PorcentajeGrasa = g
		}
	}
	if altura, ok := cambios["altura"]; ok {
		if a, ok := altura.(float64); ok {
			if a < alturaMinima || a > alturaMaxima {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("El campo 'altura' debe estar entre %.2f y %.2f metros", alturaMinima, alturaMaxima)})
			}
			existente.Altura = a
		}
	}
	if notas, ok := cambios["notas"]; ok {
		existente.Notas = notas.(string)
	}

	existente.ID = id
	pacienteActualizado, err := repo.Update(c.Request().Context(), id, *existente)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "No se encontró paciente con ID " + id})
	}
	return c.JSON(http.StatusOK, pacienteActualizado)
}

// GetPacientesPorEdad obtiene pacientes cuya edad coincida exactamente con la dada en la ruta.
func GetPacientesPorEdad(c echo.Context, repo firestore.PacienteRepository) error {
	edadStr := c.Param("edad")
	if edadStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "El campo 'edad' es requerido"})
	}
	edad, err := strconv.Atoi(edadStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "El campo 'edad' debe ser un número entero"})
	}
	pacientes, err := repo.GetAll(c.Request().Context(), 0, 0)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error al obtener pacientes"})
	}
	filtrados := make([]models.Paciente, 0)
	for _, p := range pacientes {
		if p.Edad == edad {
			filtrados = append(filtrados, p)
		}
	}
	return c.JSON(http.StatusOK, filtrados)
}

// RespuestaIMC estructura los datos calculados de IMC y su categorización clínica según la OMS.
type RespuestaIMC struct {
	ID            string  `json:"id"`
	Nombre        string  `json:"nombre"`
	Peso          float64 `json:"peso"`
	Altura        float64 `json:"altura"`
	IMC           float64 `json:"imc"`
	Clasificacion string  `json:"clasificacion"`
}

// clasificarIMC categoriza el IMC según las pautas estándar de la OMS.
func clasificarIMC(imc float64) string {
	switch {
	case imc < 18.5:
		return "Bajo peso"
	case imc < 25.0:
		return "Normal"
	case imc < 30.0:
		return "Sobrepeso"
	default:
		return "Obesidad"
	}
}

// GetPacienteIMC calcula y devuelve el Índice de Masa Corporal (IMC) y su clasificación OMS.
// Retorna 404 si el paciente no existe, o 400 si el paciente no tiene altura registrada (<= 0) o peso válido.
func GetPacienteIMC(c echo.Context, repo firestore.PacienteRepository) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "El campo 'id' es requerido"})
	}

	paciente, err := repo.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "No se encontró paciente con ID " + id})
	}

	if paciente.Altura <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "El paciente no tiene altura registrada para calcular el IMC",
		})
	}

	if paciente.Peso <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "El paciente no tiene un peso válido para calcular el IMC",
		})
	}

	imc := paciente.Peso / (paciente.Altura * paciente.Altura)
	imc = math.Round(imc*100) / 100

	return c.JSON(http.StatusOK, RespuestaIMC{
		ID:            paciente.ID,
		Nombre:        paciente.Nombre,
		Peso:          paciente.Peso,
		Altura:        paciente.Altura,
		IMC:           imc,
		Clasificacion: clasificarIMC(imc),
	})
}

// GetPacienteByID obtiene un paciente por su ID.
func GetPacienteByID(c echo.Context, repo firestore.PacienteRepository) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "El campo 'id' es requerido"})
	}
	paciente, err := repo.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "No se encontró paciente con ID " + id})
	}
	return c.JSON(http.StatusOK, paciente)
}

// DeletePaciente aplica baja lógica (soft delete) a un paciente por su ID.
func DeletePaciente(c echo.Context, repo firestore.PacienteRepository) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "El campo 'id' es requerido"})
	}
	if err := repo.Delete(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "No se encontró paciente con ID " + id})
	}
	return c.NoContent(http.StatusNoContent)
}

// RestorePaciente reactiva un paciente dado de baja lógica por su ID.
func RestorePaciente(c echo.Context, repo firestore.PacienteRepository) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "El campo 'id' es requerido"})
	}
	paciente, err := repo.Restore(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "No se encontró paciente inactivo con ID " + id})
	}
	return c.JSON(http.StatusOK, paciente)
}

// GetPacientesInactivos devuelve los pacientes dados de baja lógica.
func GetPacientesInactivos(c echo.Context, repo firestore.PacienteRepository) error {
	inactivos, err := repo.GetInactivos(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error al obtener pacientes inactivos"})
	}
	return c.JSON(http.StatusOK, inactivos)
}

// validarControlClinico verifica que peso y porcentaje_grasa del control estén en los rangos clínicos permitidos.
func validarControlClinico(control models.ControlClinico) error {
	if control.Peso <= 0 || control.Peso > pesoMaximo {
		return fmt.Errorf("El campo 'peso' debe ser mayor que 0 y menor o igual a %d", pesoMaximo)
	}
	if control.PorcentajeGrasa < 0 || control.PorcentajeGrasa > porcentajeGrasaMaximo {
		return fmt.Errorf("El campo 'porcentaje_grasa' debe estar entre 0 y %d", porcentajeGrasaMaximo)
	}
	return nil
}

// calcularIMCControl calcula el IMC redondeado a 2 decimales cuando el paciente tiene altura válida.
func calcularIMCControl(peso, altura float64) float64 {
	if altura > 0 {
		return redondear2(peso / (altura * altura))
	}
	return 0
}

// CreateControlPaciente registra un control clínico periódico para un paciente activo,
// calculando automáticamente el IMC si el paciente tiene altura registrada.
func CreateControlPaciente(c echo.Context, repo firestore.PacienteRepository) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "El campo 'id' es requerido"})
	}

	ctx := c.Request().Context()
	paciente, err := repo.GetByID(ctx, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "No se encontró paciente con ID " + id})
	}

	control := new(models.ControlClinico)
	if err := c.Bind(control); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	if err := validarControlClinico(*control); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	control.ID = ""
	control.PacienteID = id
	if control.Fecha.IsZero() {
		control.Fecha = time.Now().UTC()
	}
	control.IMC = calcularIMCControl(control.Peso, paciente.Altura)

	creado, err := repo.AddControl(ctx, id, *control)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error al registrar control clínico"})
	}
	return c.JSON(http.StatusCreated, creado)
}

// UpdateControlPaciente edita un control clínico puntual de un paciente activo,
// recalculando automáticamente el IMC según la altura actual del paciente.
func UpdateControlPaciente(c echo.Context, repo firestore.PacienteRepository) error {
	id := c.Param("id")
	controlID := c.Param("controlId")
	if id == "" || controlID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Los campos 'id' y 'controlId' son requeridos"})
	}

	ctx := c.Request().Context()
	paciente, err := repo.GetByID(ctx, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "No se encontró paciente con ID " + id})
	}

	control := new(models.ControlClinico)
	if err := c.Bind(control); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	if err := validarControlClinico(*control); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	control.ID = controlID
	control.PacienteID = id
	control.IMC = calcularIMCControl(control.Peso, paciente.Altura)

	actualizado, err := repo.UpdateControl(ctx, id, controlID, *control)
	if err != nil {
		if errors.Is(err, firestore.ErrControlNoEncontrado) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "No se encontró control clínico con ID " + controlID})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error al actualizar control clínico"})
	}
	return c.JSON(http.StatusOK, actualizado)
}

// DeleteControlPaciente elimina un control clínico puntual perteneciente a un paciente activo.
func DeleteControlPaciente(c echo.Context, repo firestore.PacienteRepository) error {
	id := c.Param("id")
	controlID := c.Param("controlId")
	if id == "" || controlID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Los campos 'id' y 'controlId' son requeridos"})
	}

	ctx := c.Request().Context()
	if _, err := repo.GetByID(ctx, id); err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "No se encontró paciente con ID " + id})
	}

	if err := repo.DeleteControl(ctx, id, controlID); err != nil {
		if errors.Is(err, firestore.ErrControlNoEncontrado) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "No se encontró control clínico con ID " + controlID})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error al eliminar control clínico"})
	}
	return c.NoContent(http.StatusNoContent)
}

// GetControlPacienteByID obtiene un control clínico puntual perteneciente a un paciente activo.
func GetControlPacienteByID(c echo.Context, repo firestore.PacienteRepository) error {
	id := c.Param("id")
	controlID := c.Param("controlId")
	if id == "" || controlID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Los campos 'id' y 'controlId' son requeridos"})
	}

	ctx := c.Request().Context()
	if _, err := repo.GetByID(ctx, id); err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "No se encontró paciente con ID " + id})
	}

	control, err := repo.GetControlByID(ctx, id, controlID)
	if err != nil {
		if errors.Is(err, firestore.ErrControlNoEncontrado) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "No se encontró control clínico con ID " + controlID})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error al obtener control clínico"})
	}
	return c.JSON(http.StatusOK, control)
}

// PatchControlPaciente aplica una actualización parcial sobre un control clínico puntual de un paciente activo.
// Solo modifica los campos enviados en el body (fecha, peso, porcentaje_grasa, notas) y recalcula el IMC.
func PatchControlPaciente(c echo.Context, repo firestore.PacienteRepository) error {
	id := c.Param("id")
	controlID := c.Param("controlId")
	if id == "" || controlID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Los campos 'id' y 'controlId' son requeridos"})
	}

	ctx := c.Request().Context()
	paciente, err := repo.GetByID(ctx, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "No se encontró paciente con ID " + id})
	}

	existente, err := repo.GetControlByID(ctx, id, controlID)
	if err != nil {
		if errors.Is(err, firestore.ErrControlNoEncontrado) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "No se encontró control clínico con ID " + controlID})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error al obtener control clínico"})
	}

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	var cambios map[string]any
	if err := json.Unmarshal(body, &cambios); err != nil || cambios == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	if rawFecha, ok := cambios["fecha"]; ok {
		fechaStr, okStr := rawFecha.(string)
		if !okStr || strings.TrimSpace(fechaStr) == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "El campo 'fecha' debe tener formato YYYY-MM-DD o RFC3339"})
		}
		parsedFecha, parseErr := parseFechaQuery(fechaStr, false)
		if parseErr != nil || parsedFecha == nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "El campo 'fecha' debe tener formato YYYY-MM-DD o RFC3339"})
		}
		existente.Fecha = *parsedFecha
	}
	if rawPeso, ok := cambios["peso"]; ok {
		peso, okNum := rawPeso.(float64)
		if !okNum {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "El campo 'peso' debe ser un número"})
		}
		if peso <= 0 || peso > pesoMaximo {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("El campo 'peso' debe ser mayor que 0 y menor o igual a %d", pesoMaximo),
			})
		}
		existente.Peso = peso
	}
	if rawGrasa, ok := cambios["porcentaje_grasa"]; ok {
		grasa, okNum := rawGrasa.(float64)
		if !okNum {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "El campo 'porcentaje_grasa' debe ser un número"})
		}
		if grasa < 0 || grasa > porcentajeGrasaMaximo {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("El campo 'porcentaje_grasa' debe estar entre 0 y %d", porcentajeGrasaMaximo),
			})
		}
		existente.PorcentajeGrasa = grasa
	}
	if rawNotas, ok := cambios["notas"]; ok {
		notas, okStr := rawNotas.(string)
		if !okStr {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "El campo 'notas' debe ser una cadena de texto"})
		}
		existente.Notas = notas
	}

	existente.ID = controlID
	existente.PacienteID = id
	existente.IMC = calcularIMCControl(existente.Peso, paciente.Altura)

	actualizado, err := repo.UpdateControl(ctx, id, controlID, *existente)
	if err != nil {
		if errors.Is(err, firestore.ErrControlNoEncontrado) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "No se encontró control clínico con ID " + controlID})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error al actualizar control clínico"})
	}
	return c.JSON(http.StatusOK, actualizado)
}

// parseFechaQuery parsea una fecha de query param en formato RFC3339 o YYYY-MM-DD.
// Si finDeDia es true y el formato es YYYY-MM-DD, ajusta el límite al final del día UTC (23:59:59.999999999).
func parseFechaQuery(valor string, finDeDia bool) (*time.Time, error) {
	if valor == "" {
		return nil, nil
	}
	if t, err := time.Parse(time.RFC3339, valor); err == nil {
		return &t, nil
	}
	if t, err := time.Parse("2006-01-02", valor); err == nil {
		if finDeDia {
			t = t.Add(24*time.Hour - time.Nanosecond)
		}
		return &t, nil
	}
	return nil, fmt.Errorf("formato de fecha inválido")
}

// obtenerControlesFiltrados valida la existencia del paciente activo y los query params ?desde=&hasta=,
// devolviendo los controles cronológicos filtrados o un código/mensaje de error HTTP.
func obtenerControlesFiltrados(c echo.Context, repo firestore.PacienteRepository) ([]models.ControlClinico, int, string) {
	id := c.Param("id")
	if id == "" {
		return nil, http.StatusBadRequest, "El campo 'id' es requerido"
	}

	ctx := c.Request().Context()
	if _, err := repo.GetByID(ctx, id); err != nil {
		return nil, http.StatusNotFound, "No se encontró paciente con ID " + id
	}

	desde, err := parseFechaQuery(c.QueryParam("desde"), false)
	if err != nil {
		return nil, http.StatusBadRequest, "El parámetro 'desde' debe tener formato YYYY-MM-DD o RFC3339"
	}
	hasta, err := parseFechaQuery(c.QueryParam("hasta"), true)
	if err != nil {
		return nil, http.StatusBadRequest, "El parámetro 'hasta' debe tener formato YYYY-MM-DD o RFC3339"
	}
	if desde != nil && hasta != nil && desde.After(*hasta) {
		return nil, http.StatusBadRequest, "El parámetro 'desde' no puede ser posterior a 'hasta'"
	}

	controles, err := repo.GetControles(ctx, id)
	if err != nil {
		return nil, http.StatusInternalServerError, "Error al obtener controles clínicos"
	}

	if desde != nil || hasta != nil {
		filtrados := make([]models.ControlClinico, 0, len(controles))
		for _, ctrl := range controles {
			if desde != nil && ctrl.Fecha.Before(*desde) {
				continue
			}
			if hasta != nil && ctrl.Fecha.After(*hasta) {
				continue
			}
			filtrados = append(filtrados, ctrl)
		}
		controles = filtrados
	}

	return controles, 0, ""
}

// GetControlesPaciente devuelve el historial de controles clínicos de un paciente activo ordenados cronológicamente.
// Soporta filtrado por rango de fechas (?desde=&hasta= en formato YYYY-MM-DD o RFC3339) y paginación (?limit=&offset=).
func GetControlesPaciente(c echo.Context, repo firestore.PacienteRepository) error {
	controles, status, errMsg := obtenerControlesFiltrados(c, repo)
	if status != 0 {
		return c.JSON(status, map[string]string{"error": errMsg})
	}

	limit, offset := paginacionDesdeQuery(c)
	total := len(controles)
	if offset >= total {
		controles = []models.ControlClinico{}
	} else {
		end := total
		if limit > 0 && offset+limit < end {
			end = offset + limit
		}
		controles = controles[offset:end]
	}

	return c.JSON(http.StatusOK, controles)
}

// ResumenControles resume la evolución clínica de un paciente entre su primer y último control registrado.
type ResumenControles struct {
	PacienteID             string     `json:"paciente_id"`
	TotalControles         int        `json:"total_controles"`
	FechaInicio            *time.Time `json:"fecha_inicio,omitempty"`
	FechaFin               *time.Time `json:"fecha_fin,omitempty"`
	PesoInicial            float64    `json:"peso_inicial"`
	PesoActual             float64    `json:"peso_actual"`
	DeltaPeso              float64    `json:"delta_peso"`
	PorcentajeGrasaInicial float64    `json:"porcentaje_grasa_inicial"`
	PorcentajeGrasaActual  float64    `json:"porcentaje_grasa_actual"`
	DeltaPorcentajeGrasa   float64    `json:"delta_porcentaje_grasa"`
	IMCInicial             float64    `json:"imc_inicial"`
	IMCActual              float64    `json:"imc_actual"`
	DeltaIMC               float64    `json:"delta_imc"`
}

func redondear2(val float64) float64 {
	return math.Round(val*100) / 100
}

// GetResumenControlesPaciente calcula el resumen evolutivo de peso, porcentaje de grasa e IMC
// entre el primer y último control clínico de un paciente activo.
func GetResumenControlesPaciente(c echo.Context, repo firestore.PacienteRepository) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "El campo 'id' es requerido"})
	}

	ctx := c.Request().Context()
	if _, err := repo.GetByID(ctx, id); err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "No se encontró paciente con ID " + id})
	}

	controles, err := repo.GetControles(ctx, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error al obtener controles clínicos"})
	}

	resumen := ResumenControles{
		PacienteID:     id,
		TotalControles: len(controles),
	}
	if len(controles) == 0 {
		return c.JSON(http.StatusOK, resumen)
	}

	primero := controles[0]
	ultimo := controles[len(controles)-1]
	fechaInicio := primero.Fecha
	fechaFin := ultimo.Fecha

	resumen.FechaInicio = &fechaInicio
	resumen.FechaFin = &fechaFin
	resumen.PesoInicial = primero.Peso
	resumen.PesoActual = ultimo.Peso
	resumen.DeltaPeso = redondear2(ultimo.Peso - primero.Peso)
	resumen.PorcentajeGrasaInicial = primero.PorcentajeGrasa
	resumen.PorcentajeGrasaActual = ultimo.PorcentajeGrasa
	resumen.DeltaPorcentajeGrasa = redondear2(ultimo.PorcentajeGrasa - primero.PorcentajeGrasa)
	resumen.IMCInicial = primero.IMC
	resumen.IMCActual = ultimo.IMC
	if primero.IMC > 0 && ultimo.IMC > 0 {
		resumen.DeltaIMC = redondear2(ultimo.IMC - primero.IMC)
	}

	return c.JSON(http.StatusOK, resumen)
}

// HealthCheck devuelve el estado operativo del servicio y el modo de persistencia activo.
func HealthCheck(c echo.Context, repo firestore.PacienteRepository) error {
	storage := "firestore"
	if _, isMock := repo.(*firestore.MockPacienteRepository); isMock {
		storage = "mock"
	}
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "ok",
		"storage": storage,
	})
}

// CountPacientes devuelve el total de pacientes registrados en el repositorio.
func CountPacientes(c echo.Context, repo firestore.PacienteRepository) error {
	pacientes, err := repo.GetAll(c.Request().Context(), 0, 0)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error al obtener pacientes"})
	}
	return c.JSON(http.StatusOK, map[string]int{
		"total": len(pacientes),
	})
}

// GruposEtarios desglosa la cantidad de pacientes por franja de edad.
type GruposEtarios struct {
	Menores int `json:"menores"`
	Adultos int `json:"adultos"`
	Mayores int `json:"mayores"`
}

// ClasificacionIMCPoblacion agrupa la cantidad de pacientes según su categoría OMS de IMC.
type ClasificacionIMCPoblacion struct {
	BajoPeso  int `json:"bajo_peso"`
	Normal    int `json:"normal"`
	Sobrepeso int `json:"sobrepeso"`
	Obesidad  int `json:"obesidad"`
	SinDatos  int `json:"sin_datos"`
}

// PromediosReporte contiene los promedios poblacionales redondeados a 2 decimales.
type PromediosReporte struct {
	Edad            float64 `json:"edad"`
	Peso            float64 `json:"peso"`
	PorcentajeGrasa float64 `json:"porcentaje_grasa"`
	IMC             float64 `json:"imc"`
}

// RespuestaReporte representa el reporte analítico demográfico y nutricional de la población de pacientes.
type RespuestaReporte struct {
	Total            int                       `json:"total"`
	GruposEtarios    GruposEtarios             `json:"grupos_etarios"`
	ClasificacionIMC ClasificacionIMCPoblacion `json:"clasificacion_imc"`
	Promedios        PromediosReporte          `json:"promedios"`
}

// ReportePacientes genera un reporte analítico consolidado con la distribución demográfica por edad,
// clasificación nutricional según el IMC (OMS) y los promedios generales de la población.
func ReportePacientes(c echo.Context, repo firestore.PacienteRepository) error {
	pacientes, err := repo.GetAll(c.Request().Context(), 0, 0)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error al obtener pacientes"})
	}

	total := len(pacientes)
	reporte := RespuestaReporte{
		Total: total,
	}

	var sumaEdad, sumaPeso, sumaGrasa, sumaIMC float64
	var conIMC int

	for _, p := range pacientes {
		switch {
		case p.Edad < 18:
			reporte.GruposEtarios.Menores++
		case p.Edad <= 65:
			reporte.GruposEtarios.Adultos++
		default:
			reporte.GruposEtarios.Mayores++
		}

		sumaEdad += float64(p.Edad)
		sumaPeso += p.Peso
		sumaGrasa += p.PorcentajeGrasa

		if p.Altura > 0 && p.Peso > 0 {
			imc := math.Round((p.Peso/(p.Altura*p.Altura))*100) / 100
			sumaIMC += imc
			conIMC++
			switch clasificarIMC(imc) {
			case "Bajo peso":
				reporte.ClasificacionIMC.BajoPeso++
			case "Normal":
				reporte.ClasificacionIMC.Normal++
			case "Sobrepeso":
				reporte.ClasificacionIMC.Sobrepeso++
			case "Obesidad":
				reporte.ClasificacionIMC.Obesidad++
			}
		} else {
			reporte.ClasificacionIMC.SinDatos++
		}
	}

	if total > 0 {
		reporte.Promedios.Edad = math.Round((sumaEdad/float64(total))*100) / 100
		reporte.Promedios.Peso = math.Round((sumaPeso/float64(total))*100) / 100
		reporte.Promedios.PorcentajeGrasa = math.Round((sumaGrasa/float64(total))*100) / 100
	}
	if conIMC > 0 {
		reporte.Promedios.IMC = math.Round((sumaIMC/float64(conIMC))*100) / 100
	}

	return c.JSON(http.StatusOK, reporte)
}

// escribirCSV serializa cabecera y filas en formato RFC 4180 y las devuelve como descarga adjunta.
func escribirCSV(c echo.Context, filename string, header []string, rows [][]string) error {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	if err := writer.Write(header); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error al generar CSV"})
	}

	for _, record := range rows {
		if err := writer.Write(record); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error al escribir registro en CSV"})
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error al finalizar CSV"})
	}

	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf(`attachment; filename="%s"`, filename))
	return c.Blob(http.StatusOK, "text/csv; charset=utf-8", buf.Bytes())
}

// ExportPacientesCSV genera un archivo CSV con el listado completo de pacientes registrados, incluyendo altura.
// Retorna Content-Type: text/csv; charset=utf-8 y Content-Disposition para forzar descarga.
func ExportPacientesCSV(c echo.Context, repo firestore.PacienteRepository) error {
	pacientes, err := repo.GetAll(c.Request().Context(), 0, 0)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error al obtener pacientes"})
	}

	header := []string{"id", "nombre", "edad", "peso", "porcentaje_grasa", "notas", "altura"}
	rows := make([][]string, 0, len(pacientes))
	for _, p := range pacientes {
		alturaStr := ""
		if p.Altura > 0 {
			alturaStr = strconv.FormatFloat(p.Altura, 'f', -1, 64)
		}
		rows = append(rows, []string{
			p.ID,
			p.Nombre,
			strconv.Itoa(p.Edad),
			strconv.FormatFloat(p.Peso, 'f', -1, 64),
			strconv.FormatFloat(p.PorcentajeGrasa, 'f', -1, 64),
			p.Notas,
			alturaStr,
		})
	}

	return escribirCSV(c, "pacientes.csv", header, rows)
}

// ExportControlesPacienteCSV exporta en formato CSV RFC 4180 el historial cronológico de controles clínicos
// de un paciente activo, soportando filtrado opcional por rango de fechas (?desde=&hasta=).
func ExportControlesPacienteCSV(c echo.Context, repo firestore.PacienteRepository) error {
	controles, status, errMsg := obtenerControlesFiltrados(c, repo)
	if status != 0 {
		return c.JSON(status, map[string]string{"error": errMsg})
	}

	header := []string{"id", "paciente_id", "fecha", "peso", "porcentaje_grasa", "imc", "notas"}
	rows := make([][]string, 0, len(controles))
	for _, ctrl := range controles {
		imcStr := ""
		if ctrl.IMC > 0 {
			imcStr = strconv.FormatFloat(ctrl.IMC, 'f', -1, 64)
		}
		rows = append(rows, []string{
			ctrl.ID,
			ctrl.PacienteID,
			ctrl.Fecha.UTC().Format(time.RFC3339),
			strconv.FormatFloat(ctrl.Peso, 'f', -1, 64),
			strconv.FormatFloat(ctrl.PorcentajeGrasa, 'f', -1, 64),
			imcStr,
			ctrl.Notas,
		})
	}

	return escribirCSV(c, "controles.csv", header, rows)
}

// SetupRoutes configura las rutas de pacientes con inyección de dependencia.
func SetupRoutes(e *echo.Echo, repo firestore.PacienteRepository) {
	e.GET("/health", func(c echo.Context) error {
		return HealthCheck(c, repo)
	})
	e.GET("/pacientes", func(c echo.Context) error {
		return GetPacientes(c, repo)
	})
	e.POST("/pacientes", func(c echo.Context) error {
		return CreatePaciente(c, repo)
	})
	e.POST("/pacientes/batch", func(c echo.Context) error {
		return CreatePacientesBatch(c, repo)
	})
	e.GET("/pacientes/search", func(c echo.Context) error {
		return SearchPacientes(c, repo)
	})
	e.GET("/pacientes/stats", func(c echo.Context) error {
		return StatsPacientes(c, repo)
	})
	e.GET("/pacientes/reporte", func(c echo.Context) error {
		return ReportePacientes(c, repo)
	})
	e.GET("/pacientes/count", func(c echo.Context) error {
		return CountPacientes(c, repo)
	})
	e.GET("/pacientes/export", func(c echo.Context) error {
		return ExportPacientesCSV(c, repo)
	})
	e.GET("/pacientes/inactivos", func(c echo.Context) error {
		return GetPacientesInactivos(c, repo)
	})
	e.POST("/pacientes/:id/restore", func(c echo.Context) error {
		return RestorePaciente(c, repo)
	})
	e.POST("/pacientes/{id}/restore", func(c echo.Context) error {
		return RestorePaciente(c, repo)
	})
	e.GET("/pacientes/edad/:edad", func(c echo.Context) error {
		return GetPacientesPorEdad(c, repo)
	})
	e.GET("/pacientes/edad/{edad}", func(c echo.Context) error {
		return GetPacientesPorEdad(c, repo)
	})
	e.GET("/pacientes/imc/:id", func(c echo.Context) error {
		return GetPacienteIMC(c, repo)
	})
	e.GET("/pacientes/imc/{id}", func(c echo.Context) error {
		return GetPacienteIMC(c, repo)
	})
	e.GET("/pacientes/:id/imc", func(c echo.Context) error {
		return GetPacienteIMC(c, repo)
	})
	e.GET("/pacientes/{id}/imc", func(c echo.Context) error {
		return GetPacienteIMC(c, repo)
	})
	e.GET("/pacientes/:id/controles/resumen", func(c echo.Context) error {
		return GetResumenControlesPaciente(c, repo)
	})
	e.GET("/pacientes/{id}/controles/resumen", func(c echo.Context) error {
		return GetResumenControlesPaciente(c, repo)
	})
	e.GET("/pacientes/:id/controles/export", func(c echo.Context) error {
		return ExportControlesPacienteCSV(c, repo)
	})
	e.GET("/pacientes/{id}/controles/export", func(c echo.Context) error {
		return ExportControlesPacienteCSV(c, repo)
	})
	e.GET("/pacientes/:id/controles/:controlId", func(c echo.Context) error {
		return GetControlPacienteByID(c, repo)
	})
	e.GET("/pacientes/{id}/controles/{controlId}", func(c echo.Context) error {
		return GetControlPacienteByID(c, repo)
	})
	e.GET("/pacientes/:id/controles", func(c echo.Context) error {
		return GetControlesPaciente(c, repo)
	})
	e.GET("/pacientes/{id}/controles", func(c echo.Context) error {
		return GetControlesPaciente(c, repo)
	})
	e.POST("/pacientes/:id/controles", func(c echo.Context) error {
		return CreateControlPaciente(c, repo)
	})
	e.POST("/pacientes/{id}/controles", func(c echo.Context) error {
		return CreateControlPaciente(c, repo)
	})
	e.PUT("/pacientes/:id/controles/:controlId", func(c echo.Context) error {
		return UpdateControlPaciente(c, repo)
	})
	e.PUT("/pacientes/{id}/controles/{controlId}", func(c echo.Context) error {
		return UpdateControlPaciente(c, repo)
	})
	e.PATCH("/pacientes/:id/controles/:controlId", func(c echo.Context) error {
		return PatchControlPaciente(c, repo)
	})
	e.PATCH("/pacientes/{id}/controles/{controlId}", func(c echo.Context) error {
		return PatchControlPaciente(c, repo)
	})
	e.DELETE("/pacientes/:id/controles/:controlId", func(c echo.Context) error {
		return DeleteControlPaciente(c, repo)
	})
	e.DELETE("/pacientes/{id}/controles/{controlId}", func(c echo.Context) error {
		return DeleteControlPaciente(c, repo)
	})
	e.GET("/pacientes/:id", func(c echo.Context) error {
		return GetPacienteByID(c, repo)
	})
	e.GET("/pacientes/{id}", func(c echo.Context) error {
		return GetPacienteByID(c, repo)
	})
	e.PUT("/pacientes/:id", func(c echo.Context) error {
		return UpdatePaciente(c, repo)
	})
	e.PUT("/pacientes/{id}", func(c echo.Context) error {
		return UpdatePaciente(c, repo)
	})
	e.PATCH("/pacientes/:id", func(c echo.Context) error {
		return PatchPaciente(c, repo)
	})
	e.PATCH("/pacientes/{id}", func(c echo.Context) error {
		return PatchPaciente(c, repo)
	})
	e.DELETE("/pacientes/:id", func(c echo.Context) error {
		return DeletePaciente(c, repo)
	})
	e.DELETE("/pacientes/{id}", func(c echo.Context) error {
		return DeletePaciente(c, repo)
	})
}
