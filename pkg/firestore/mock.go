package firestore

import (
	"context"
	"fmt"
	"sort"
	"sync/atomic"
	"time"

	"nutricontrol/models"
)

// MockPacienteRepository implementa PacienteRepository sin llamar Firestore.
// Usado en tests para verificar que handlers interactúan correctamente con el repositorio.
type MockPacienteRepository struct {
	store         []models.Paciente
	nextID        int64
	controles     map[string][]models.ControlClinico
	nextControlID int64
}

func NewMockPacienteRepository() *MockPacienteRepository {
	return &MockPacienteRepository{
		store: []models.Paciente{
			{Nombre: "Juan Perez", Edad: 30, Peso: 75.5, PorcentajeGrasa: 18.5, Notas: "Primera consulta"},
			{Nombre: "Maria Gomez", Edad: 25, Peso: 62.0, PorcentajeGrasa: 22.0, Notas: "Seguimiento mensual"},
		},
		nextID:    100,
		controles: make(map[string][]models.ControlClinico),
	}
}

func (m *MockPacienteRepository) Create(ctx context.Context, p models.Paciente) (*models.Paciente, error) {
	if p.ID == "" {
		p.ID = fmt.Sprintf("paciente-%d", atomic.AddInt64(&m.nextID, 1))
	}
	m.store = append(m.store, p)
	return &p, nil
}

func (m *MockPacienteRepository) GetAll(ctx context.Context, limit, offset int) ([]models.Paciente, error) {
	activos := make([]models.Paciente, 0, len(m.store))
	for _, p := range m.store {
		if p.DeletedAt == nil {
			activos = append(activos, p)
		}
	}
	total := len(activos)
	if offset < 0 {
		offset = 0
	}
	if offset >= total {
		return []models.Paciente{}, nil
	}
	end := total
	if limit > 0 && offset+limit < end {
		end = offset + limit
	}
	result := make([]models.Paciente, end-offset)
	copy(result, activos[offset:end])
	return result, nil
}

// GetByID recupera un paciente activo por ID desde el mock.
func (m *MockPacienteRepository) GetByID(ctx context.Context, id string) (*models.Paciente, error) {
	for _, p := range m.store {
		if p.ID == id && p.DeletedAt == nil {
			copia := p
			return &copia, nil
		}
	}
	return nil, fmt.Errorf("no se encontró paciente con ID %s", id)
}

// Update modifica un paciente activo existente por ID.
func (m *MockPacienteRepository) Update(ctx context.Context, id string, p models.Paciente) (*models.Paciente, error) {
	for i, existing := range m.store {
		if existing.ID == id && existing.DeletedAt == nil {
			m.store[i].Nombre = p.Nombre
			m.store[i].Edad = p.Edad
			m.store[i].Peso = p.Peso
			m.store[i].Altura = p.Altura
			m.store[i].PorcentajeGrasa = p.PorcentajeGrasa
			m.store[i].Notas = p.Notas
			copia := m.store[i]
			return &copia, nil
		}
	}
	return nil, fmt.Errorf("no se encontró paciente con ID %s", id)
}

// Delete realiza baja lógica (soft delete) marcando DeletedAt en el paciente.
func (m *MockPacienteRepository) Delete(ctx context.Context, id string) error {
	for i, p := range m.store {
		if p.ID == id && p.DeletedAt == nil {
			now := time.Now().UTC()
			m.store[i].DeletedAt = &now
			return nil
		}
	}
	return fmt.Errorf("no se encontró paciente con ID %s", id)
}

// Restore reactiva un paciente dado de baja lógica limpiando DeletedAt.
func (m *MockPacienteRepository) Restore(ctx context.Context, id string) (*models.Paciente, error) {
	for i, p := range m.store {
		if p.ID == id && p.DeletedAt != nil {
			m.store[i].DeletedAt = nil
			copia := m.store[i]
			return &copia, nil
		}
	}
	return nil, fmt.Errorf("no se encontró paciente inactivo con ID %s", id)
}

// GetInactivos devuelve todos los pacientes dados de baja lógica.
func (m *MockPacienteRepository) GetInactivos(ctx context.Context) ([]models.Paciente, error) {
	inactivos := make([]models.Paciente, 0)
	for _, p := range m.store {
		if p.DeletedAt != nil {
			inactivos = append(inactivos, p)
		}
	}
	return inactivos, nil
}

// AddControl registra un nuevo control clínico asociado a un paciente activo en el mock.
func (m *MockPacienteRepository) AddControl(ctx context.Context, pacienteID string, control models.ControlClinico) (*models.ControlClinico, error) {
	if _, err := m.GetByID(ctx, pacienteID); err != nil {
		return nil, err
	}
	if m.controles == nil {
		m.controles = make(map[string][]models.ControlClinico)
	}
	if control.ID == "" {
		control.ID = fmt.Sprintf("control-%d", atomic.AddInt64(&m.nextControlID, 1))
	}
	control.PacienteID = pacienteID
	m.controles[pacienteID] = append(m.controles[pacienteID], control)
	copia := control
	return &copia, nil
}

// GetControles devuelve los controles clínicos de un paciente activo ordenados cronológicamente (Fecha ascendente).
func (m *MockPacienteRepository) GetControles(ctx context.Context, pacienteID string) ([]models.ControlClinico, error) {
	if _, err := m.GetByID(ctx, pacienteID); err != nil {
		return nil, err
	}
	if m.controles == nil || len(m.controles[pacienteID]) == 0 {
		return []models.ControlClinico{}, nil
	}
	res := make([]models.ControlClinico, len(m.controles[pacienteID]))
	copy(res, m.controles[pacienteID])
	sort.SliceStable(res, func(i, j int) bool {
		return res[i].Fecha.Before(res[j].Fecha)
	})
	return res, nil
}

// GetControlByID recupera un control clínico puntual de un paciente activo en el mock.
func (m *MockPacienteRepository) GetControlByID(ctx context.Context, pacienteID, controlID string) (*models.ControlClinico, error) {
	if _, err := m.GetByID(ctx, pacienteID); err != nil {
		return nil, err
	}
	if m.controles != nil {
		for _, existing := range m.controles[pacienteID] {
			if existing.ID == controlID {
				copia := existing
				return &copia, nil
			}
		}
	}
	return nil, fmt.Errorf("%w con ID %s", ErrControlNoEncontrado, controlID)
}

// UpdateControl actualiza un control clínico existente de un paciente activo en el mock.
// Si control.Fecha es cero, conserva la fecha original del control.
func (m *MockPacienteRepository) UpdateControl(ctx context.Context, pacienteID, controlID string, control models.ControlClinico) (*models.ControlClinico, error) {
	if _, err := m.GetByID(ctx, pacienteID); err != nil {
		return nil, err
	}
	if m.controles != nil {
		for i, existing := range m.controles[pacienteID] {
			if existing.ID == controlID {
				if control.Fecha.IsZero() {
					control.Fecha = existing.Fecha
				}
				control.ID = controlID
				control.PacienteID = pacienteID
				m.controles[pacienteID][i] = control
				copia := control
				return &copia, nil
			}
		}
	}
	return nil, fmt.Errorf("%w con ID %s", ErrControlNoEncontrado, controlID)
}

// DeleteControl elimina un control clínico puntual de un paciente activo en el mock.
func (m *MockPacienteRepository) DeleteControl(ctx context.Context, pacienteID, controlID string) error {
	if _, err := m.GetByID(ctx, pacienteID); err != nil {
		return err
	}
	if m.controles != nil {
		lista := m.controles[pacienteID]
		for i, existing := range lista {
			if existing.ID == controlID {
				m.controles[pacienteID] = append(lista[:i], lista[i+1:]...)
				return nil
			}
		}
	}
	return fmt.Errorf("%w con ID %s", ErrControlNoEncontrado, controlID)
}
