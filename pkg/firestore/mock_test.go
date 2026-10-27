package firestore

import (
	"context"
	"testing"

	"nutricontrol/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockCreateAsignaIDCuandoEstaVacio(t *testing.T) {
	repo := NewMockPacienteRepository()
	p := models.Paciente{Nombre: "Sin ID", Edad: 30, Peso: 70}

	creado, err := repo.Create(context.Background(), p)
	require.NoError(t, err)
	assert.NotEmpty(t, creado.ID)
	assert.Equal(t, "Sin ID", creado.Nombre)
	assert.Equal(t, 30, creado.Edad)
	assert.Equal(t, 70.0, creado.Peso)
}

func TestMockCreateRespetaIDExistente(t *testing.T) {
	repo := NewMockPacienteRepository()
	p := models.Paciente{ID: "mi-id", Nombre: "Con ID", Edad: 30, Peso: 70}

	creado, err := repo.Create(context.Background(), p)
	require.NoError(t, err)
	assert.Equal(t, "mi-id", creado.ID)
}
