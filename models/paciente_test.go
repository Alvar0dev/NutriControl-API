package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPacienteJSONSerializacion(t *testing.T) {
	// Arrange
	paciente := Paciente{
		Nombre:          "Juan Perez",
		Edad:            30,
		Peso:            75.5,
		PorcentajeGrasa: 18.5,
		Notas:           "Primera consulta",
	}

	// Act
	data, err := json.Marshal(paciente)

	// Assert
	require.NoError(t, err)
	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "Juan Perez", decoded["nombre"])
	assert.Equal(t, float64(30), decoded["edad"])
	assert.Equal(t, 75.5, decoded["peso"])
	assert.Equal(t, 18.5, decoded["porcentaje_grasa"])
	assert.Equal(t, "Primera consulta", decoded["notas"])
}

func TestPacienteJSONSerializacionConAltura(t *testing.T) {
	// Arrange
	paciente := Paciente{
		Nombre:          "Ana Silva",
		Edad:            28,
		Peso:            65.0,
		Altura:          1.70,
		PorcentajeGrasa: 20.0,
		Notas:           "Control anual",
	}

	// Act
	data, err := json.Marshal(paciente)

	// Assert
	require.NoError(t, err)
	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "Ana Silva", decoded["nombre"])
	assert.Equal(t, float64(28), decoded["edad"])
	assert.Equal(t, 65.0, decoded["peso"])
	assert.Equal(t, 1.70, decoded["altura"])
	assert.Equal(t, 20.0, decoded["porcentaje_grasa"])
	assert.Equal(t, "Control anual", decoded["notas"])
	_, tieneDeletedAt := decoded["deleted_at"]
	assert.False(t, tieneDeletedAt)
}

func TestPacienteJSONSerializacionConDeletedAt(t *testing.T) {
	fechaBaja := time.Date(2026, 11, 17, 10, 0, 0, 0, time.UTC)
	paciente := Paciente{
		Nombre:    "Carlos Baja",
		Edad:      40,
		Peso:      80.0,
		DeletedAt: &fechaBaja,
	}

	data, err := json.Marshal(paciente)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "2026-11-17T10:00:00Z", decoded["deleted_at"])
}

func TestControlClinicoJSONSerializacion(t *testing.T) {
	fecha := time.Date(2026, 11, 17, 12, 30, 0, 0, time.UTC)
	control := ControlClinico{
		ID:              "control-1",
		PacienteID:      "paciente-101",
		Fecha:           fecha,
		Peso:            74.2,
		PorcentajeGrasa: 17.8,
		IMC:             24.23,
		Notas:           "Buen progreso",
	}

	data, err := json.Marshal(control)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "control-1", decoded["id"])
	assert.Equal(t, "paciente-101", decoded["paciente_id"])
	assert.Equal(t, "2026-11-17T12:30:00Z", decoded["fecha"])
	assert.Equal(t, 74.2, decoded["peso"])
	assert.Equal(t, 17.8, decoded["porcentaje_grasa"])
	assert.Equal(t, 24.23, decoded["imc"])
	assert.Equal(t, "Buen progreso", decoded["notas"])
}
