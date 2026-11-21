package firestore

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

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

func TestMockSoftDeleteYRestore(t *testing.T) {
	repo := NewMockPacienteRepository()
	ctx := context.Background()

	creado, err := repo.Create(ctx, models.Paciente{Nombre: "Paciente Temporal", Edad: 28, Peso: 70.0})
	require.NoError(t, err)

	// Soft delete
	err = repo.Delete(ctx, creado.ID)
	require.NoError(t, err)

	// Excluido de GetByID y GetAll
	_, err = repo.GetByID(ctx, creado.ID)
	assert.Error(t, err)

	activos, err := repo.GetAll(ctx, 0, 0)
	require.NoError(t, err)
	assert.Len(t, activos, 2)

	// No se puede eliminar dos veces ni actualizar mientras esté inactivo
	err = repo.Delete(ctx, creado.ID)
	assert.Error(t, err)
	_, err = repo.Update(ctx, creado.ID, models.Paciente{Nombre: "Otro", Edad: 29, Peso: 71.0})
	assert.Error(t, err)

	// Aparece en GetInactivos con DeletedAt seteado
	inactivos, err := repo.GetInactivos(ctx)
	require.NoError(t, err)
	require.Len(t, inactivos, 1)
	assert.Equal(t, creado.ID, inactivos[0].ID)
	assert.NotNil(t, inactivos[0].DeletedAt)

	// Restore vuelve a activarlo
	restaurado, err := repo.Restore(ctx, creado.ID)
	require.NoError(t, err)
	assert.Nil(t, restaurado.DeletedAt)

	recuperado, err := repo.GetByID(ctx, creado.ID)
	require.NoError(t, err)
	assert.Equal(t, "Paciente Temporal", recuperado.Nombre)

	inactivosPost, err := repo.GetInactivos(ctx)
	require.NoError(t, err)
	assert.Len(t, inactivosPost, 0)

	// Restore sobre paciente activo o inexistente falla
	_, err = repo.Restore(ctx, creado.ID)
	assert.Error(t, err)
	_, err = repo.Restore(ctx, "no-existe")
	assert.Error(t, err)
}

func TestMockGetAllPaginacionExcluyeSoftDeleted(t *testing.T) {
	repo := NewMockPacienteRepository()
	ctx := context.Background()

	p1, err := repo.Create(ctx, models.Paciente{Nombre: "P1", Edad: 20, Peso: 60.0})
	require.NoError(t, err)
	_, err = repo.Create(ctx, models.Paciente{Nombre: "P2", Edad: 21, Peso: 61.0})
	require.NoError(t, err)

	// Eliminar P1 (queda en medio de m.store: Juan, Maria, P1(del), P2)
	err = repo.Delete(ctx, p1.ID)
	require.NoError(t, err)

	// limit=2, offset=1 debe saltar a Juan y devolver [Maria Gomez, P2]
	res, err := repo.GetAll(ctx, 2, 1)
	require.NoError(t, err)
	require.Len(t, res, 2)
	assert.Equal(t, "Maria Gomez", res[0].Nombre)
	assert.Equal(t, "P2", res[1].Nombre)
}

func TestMockAddYGetControles(t *testing.T) {
	// Probar también struct literal vacío (nil map safety)
	repoVacio := &MockPacienteRepository{}
	_, err := repoVacio.AddControl(context.Background(), "no-existe", models.ControlClinico{Peso: 70})
	assert.Error(t, err)

	repo := NewMockPacienteRepository()
	ctx := context.Background()
	creado, err := repo.Create(ctx, models.Paciente{Nombre: "Control Test", Edad: 30, Peso: 80.0, Altura: 1.80})
	require.NoError(t, err)

	// Lista vacía inicial
	listaInicial, err := repo.GetControles(ctx, creado.ID)
	require.NoError(t, err)
	assert.Len(t, listaInicial, 0)

	// Agregar 2 controles fuera de orden cronológico para verificar ordenamiento por Fecha asc
	t2 := time.Date(2026, 11, 10, 10, 0, 0, 0, time.UTC)
	t1 := time.Date(2026, 10, 10, 10, 0, 0, 0, time.UTC)
	c2, err := repo.AddControl(ctx, creado.ID, models.ControlClinico{Fecha: t2, Peso: 78.0, PorcentajeGrasa: 19.0, IMC: 24.07})
	require.NoError(t, err)
	assert.NotEmpty(t, c2.ID)
	assert.Equal(t, creado.ID, c2.PacienteID)

	c1, err := repo.AddControl(ctx, creado.ID, models.ControlClinico{Fecha: t1, Peso: 80.0, PorcentajeGrasa: 20.5, IMC: 24.69})
	require.NoError(t, err)
	assert.NotEmpty(t, c1.ID)

	controles, err := repo.GetControles(ctx, creado.ID)
	require.NoError(t, err)
	require.Len(t, controles, 2)
	assert.True(t, controles[0].Fecha.Equal(t1))
	assert.True(t, controles[1].Fecha.Equal(t2))

	// Paciente eliminado lógicamente no admite lectura/creación de controles
	require.NoError(t, repo.Delete(ctx, creado.ID))
	_, err = repo.GetControles(ctx, creado.ID)
	assert.Error(t, err)
	_, err = repo.AddControl(ctx, creado.ID, models.ControlClinico{Fecha: t2, Peso: 77.0})
	assert.Error(t, err)
}

func TestMockUpdateYDeleteControl(t *testing.T) {
	ctx := context.Background()

	// nil-map safety en struct literal vacío
	repoVacio := &MockPacienteRepository{}
	_, err := repoVacio.UpdateControl(ctx, "no-existe", "c1", models.ControlClinico{Peso: 70})
	assert.Error(t, err)
	assert.Error(t, repoVacio.DeleteControl(ctx, "no-existe", "c1"))

	repo := NewMockPacienteRepository()
	p, err := repo.Create(ctx, models.Paciente{Nombre: "Control CRUD", Edad: 30, Peso: 80, Altura: 1.80})
	require.NoError(t, err)

	t1 := time.Date(2026, 10, 10, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 10, 12, 10, 0, 0, 0, time.UTC)
	c1, err := repo.AddControl(ctx, p.ID, models.ControlClinico{Fecha: t1, Peso: 80.0, PorcentajeGrasa: 20.0, IMC: 24.69, Notas: "Inicial"})
	require.NoError(t, err)

	// Update con fecha cero conserva fecha original
	act1, err := repo.UpdateControl(ctx, p.ID, c1.ID, models.ControlClinico{Peso: 78.0, PorcentajeGrasa: 19.0, IMC: 24.07, Notas: "Ajustado"})
	require.NoError(t, err)
	assert.True(t, act1.Fecha.Equal(t1))
	assert.Equal(t, 78.0, act1.Peso)
	assert.Equal(t, "Ajustado", act1.Notas)

	// Update con fecha explícita reemplaza fecha
	act2, err := repo.UpdateControl(ctx, p.ID, c1.ID, models.ControlClinico{Fecha: t2, Peso: 77.0, PorcentajeGrasa: 18.5, IMC: 23.77, Notas: "Reprogramado"})
	require.NoError(t, err)
	assert.True(t, act2.Fecha.Equal(t2))
	assert.Equal(t, 77.0, act2.Peso)

	// Control inexistente devuelve ErrControlNoEncontrado
	_, err = repo.UpdateControl(ctx, p.ID, "inexistente", models.ControlClinico{Peso: 70})
	assert.True(t, errors.Is(err, ErrControlNoEncontrado))
	err = repo.DeleteControl(ctx, p.ID, "inexistente")
	assert.True(t, errors.Is(err, ErrControlNoEncontrado))

	// Delete elimina el control del historial
	require.NoError(t, repo.DeleteControl(ctx, p.ID, c1.ID))
	lista, err := repo.GetControles(ctx, p.ID)
	require.NoError(t, err)
	assert.Empty(t, lista)

	// Segunda eliminación devuelve ErrControlNoEncontrado
	err = repo.DeleteControl(ctx, p.ID, c1.ID)
	assert.True(t, errors.Is(err, ErrControlNoEncontrado))

	// Soft delete bloquea update y delete de controles
	c2, err := repo.AddControl(ctx, p.ID, models.ControlClinico{Fecha: t1, Peso: 76.0})
	require.NoError(t, err)
	require.NoError(t, repo.Delete(ctx, p.ID))
	_, err = repo.UpdateControl(ctx, p.ID, c2.ID, models.ControlClinico{Peso: 75.0})
	assert.Error(t, err)
	assert.Error(t, repo.DeleteControl(ctx, p.ID, c2.ID))
}
