package firestore

import (
	"context"
	"errors"

	"nutricontrol/models"
)

// ErrControlNoEncontrado indica que el control clínico solicitado no existe para el paciente dado.
var ErrControlNoEncontrado = errors.New("control clínico no encontrado")

// PacienteRepository define la interfaz del repositorio de pacientes.
// Se usa para separar la dependencia de concreción real (Firestore) de los handlers.
// Para testing, se sustituye por un MockPacienteRepository sin llamar Firestore.
type PacienteRepository interface {
	Create(ctx context.Context, p models.Paciente) (*models.Paciente, error)
	GetAll(ctx context.Context, limit, offset int) ([]models.Paciente, error)
	GetByID(ctx context.Context, id string) (*models.Paciente, error)
	Update(ctx context.Context, id string, p models.Paciente) (*models.Paciente, error)
	Delete(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) (*models.Paciente, error)
	GetInactivos(ctx context.Context) ([]models.Paciente, error)
	AddControl(ctx context.Context, pacienteID string, control models.ControlClinico) (*models.ControlClinico, error)
	GetControles(ctx context.Context, pacienteID string) ([]models.ControlClinico, error)
	UpdateControl(ctx context.Context, pacienteID, controlID string, control models.ControlClinico) (*models.ControlClinico, error)
	DeleteControl(ctx context.Context, pacienteID, controlID string) error
}
