package models

import (
	"encoding/json"
	"testing"

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
