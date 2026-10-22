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

// GetByID recupera un paciente por ID desde el mock.
func (m *MockPacienteRepository) GetByID(ctx context.Context, id string) (*models.Paciente, error) {
	for _, p := range m.store {
		if p.ID == id {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("no se encontró paciente con ID %s", id)
}

// Update modifica un paciente existente por ID.
func (m *MockPacienteRepository) Update(ctx context.Context, id string, p models.Paciente) (*models.Paciente, error) {
	for i, existing := range m.store {
		if existing.ID == id {
			m.store[i].Nombre = p.Nombre
			m.store[i].Edad = p.Edad
			m.store[i].Peso = p.Peso
			m.store[i].PorcentajeGrasa = p.PorcentajeGrasa
			m.store[i].Notas = p.Notas
			return &m.store[i], nil
		}
	}
	return nil, fmt.Errorf("no se encontró paciente con ID %s", id)
}

// Delete elimina un paciente del mock por ID.
func (m *MockPacienteRepository) Delete(ctx context.Context, id string) error {
	for i, p := range m.store {
		if p.ID == id {
			m.store = append(m.store[:i], m.store[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("no se encontró paciente con ID %s", id)
}
