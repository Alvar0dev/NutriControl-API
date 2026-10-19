package firestore

import (
	"context"
	"fmt"
	"nutricontrol/models"
)

// MockPacienteRepository implementa PacienteRepository sin llamar Firestore.
// Usado en tests para verificar que handlers interactúan correctamente con el repositorio.
type MockPacienteRepository struct {
	store []models.Paciente
}

func NewMockPacienteRepository() *MockPacienteRepository {
	return &MockPacienteRepository{
		store: []models.Paciente{
			{Nombre: "Juan Perez", Edad: 30, Peso: 75.5, PorcentajeGrasa: 18.5, Notas: "Primera consulta"},
			{Nombre: "Maria Gomez", Edad: 25, Peso: 62.0, PorcentajeGrasa: 22.0, Notas: "Seguimiento mensual"},
		},
	}
}

func (m *MockPacienteRepository) Create(ctx context.Context, p models.Paciente) (*models.Paciente, error) {
	m.store = append(m.store, p)
	return &p, nil
}

func (m *MockPacienteRepository) GetAll(ctx context.Context) ([]models.Paciente, error) {
	return m.store, nil
}
