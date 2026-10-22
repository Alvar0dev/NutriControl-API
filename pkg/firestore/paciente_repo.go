package firestore

import (
	"context"
	"nutricontrol/models"
)

// PacienteRepository define la interfaz del repositorio de pacientes.
// Se usa para separar la dependencia de concreción real (Firestore) de los handlers.
// Para testing, se sustituye por un MockPacienteRepository sin llamar Firestore.
type PacienteRepository interface {
	Create(ctx context.Context, p models.Paciente) (*models.Paciente, error)
	GetAll(ctx context.Context) ([]models.Paciente, error)
	GetByID(ctx context.Context, id string) (*models.Paciente, error)
	Update(ctx context.Context, id string, p models.Paciente) (*models.Paciente, error)
	Delete(ctx context.Context, id string) error
}
