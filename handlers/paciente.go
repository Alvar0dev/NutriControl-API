package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"

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

// GetPacientes obtiene todos los pacientes mediante el repositorio.
// Soporta filtro opcional por nombre (?nombre=, coincidencia parcial, sin distinción de mayúsculas)
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
	return c.JSON(http.StatusOK, pacientes)
}

// SearchPacientes busca pacientes por múltiples criterios con paginación.
// Query params:
//   - nombre: coincidencia parcial, sin distinción de mayúsculas/minúsculas
//   - edad_min, edad_max: rango inclusivo de edad
//   - peso_min, peso_max: rango inclusivo de peso
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
	if err := validarPaciente(*paciente); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	pacienteCreado, err := repo.Create(c.Request().Context(), *paciente)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error al crear paciente"})
	}
	return c.JSON(http.StatusCreated, pacienteCreado)
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

// DeletePaciente elimina un paciente por su ID.
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

// SetupRoutes configura las rutas de pacientes con inyección de dependencia.
func SetupRoutes(e *echo.Echo, repo firestore.PacienteRepository) {
	e.GET("/pacientes", func(c echo.Context) error {
		return GetPacientes(c, repo)
	})
	e.POST("/pacientes", func(c echo.Context) error {
		return CreatePaciente(c, repo)
	})
	e.GET("/pacientes/{id}", func(c echo.Context) error {
		return GetPacienteByID(c, repo)
	})
	e.PUT("/pacientes/{id}", func(c echo.Context) error {
		return UpdatePaciente(c, repo)
	})
	e.PATCH("/pacientes/{id}", func(c echo.Context) error {
		return PatchPaciente(c, repo)
	})
	e.DELETE("/pacientes/{id}", func(c echo.Context) error {
		return DeletePaciente(c, repo)
	})
	e.GET("/pacientes/search", func(c echo.Context) error {
		return SearchPacientes(c, repo)
	})
	e.GET("/pacientes/stats", func(c echo.Context) error {
		return StatsPacientes(c, repo)
	})
	e.GET("/pacientes/edad/{edad}", func(c echo.Context) error {
		return GetPacientesPorEdad(c, repo)
	})
}
