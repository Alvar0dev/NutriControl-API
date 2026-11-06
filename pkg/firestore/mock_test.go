package firestore

import (
	"context"
	"fmt"
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

// TestMockGetAllRespetaLimitYOffset verifica que el mock aplique limit y offset en memoria.
func TestMockGetAllRespetaLimitYOffset(t *testing.T) {
	repo := NewMockPacienteRepository()
	// store inicial: Juan (0), Maria (1)
	for i := 0; i < 3; i++ {
		_, _ = repo.Create(context.Background(), models.Paciente{
			Nombre: fmt.Sprintf("Extra %d", i),
			Edad:   20 + i,
			Peso:   60.0 + float64(i),
		})
	}
	// total 5: Juan, Maria, Extra 0, Extra 1, Extra 2

	// limit=2, offset=1 -> Maria, Extra 0
	res, err := repo.GetAll(context.Background(), 2, 1)
	require.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, "Maria Gomez", res[0].Nombre)
	assert.Equal(t, "Extra 0", res[1].Nombre)

	// limit=10, offset=0 -> todos los 5
	res, err = repo.GetAll(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Len(t, res, 5)

	// offset mayor al total -> vacío
	res, err = repo.GetAll(context.Background(), 5, 100)
	require.NoError(t, err)
	assert.Len(t, res, 0)

	// limit<=0 -> sin límite (todos)
	res, err = repo.GetAll(context.Background(), 0, 0)
	require.NoError(t, err)
	assert.Len(t, res, 5)
}
