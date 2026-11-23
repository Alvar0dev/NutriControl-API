package firestore

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"nutricontrol/models"
)

// PacienteFirestoreRepo es la implementación real con Firestore.
// Usa el cliente de Firestore de Google Cloud SDK.
type PacienteFirestoreRepo struct {
	client *firestore.Client
	col    *firestore.CollectionRef
}

// NewPacienteFirestoreRepo inicializa el repositorio con Firestore.
func NewPacienteFirestoreRepo(ctx context.Context, projectID string) (*PacienteFirestoreRepo, error) {
	var opts []option.ClientOption
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		opts = append(opts, option.WithEndpoint("firestore.googleapis.com:443"))
	}
	client, err := firestore.NewClient(ctx, projectID, opts...)
	if err != nil {
		return nil, err
	}

	return NewPacienteFirestoreRepoWithClient(client), nil
}

// NewPacienteFirestoreRepoWithClient inicializa el repositorio reutilizando un cliente Firestore existente.
func NewPacienteFirestoreRepoWithClient(client *firestore.Client) *PacienteFirestoreRepo {
	return &PacienteFirestoreRepo{
		client: client,
		col:    client.Collection("pacientes"),
	}
}

// Create inserta un nuevo paciente en Firestore.
func (r *PacienteFirestoreRepo) Create(ctx context.Context, p models.Paciente) (*models.Paciente, error) {
	docRef, _, err := r.col.Add(ctx, p)
	if err != nil {
		return nil, err
	}
	p.ID = docRef.ID
	return &p, nil
}

// GetAll recupera pacientes activos de Firestore aplicando paginación (limit/offset) sobre registros sin DeletedAt.
// Cuando limit <= 0 no se aplica límite; cuando offset <= 0 no se aplica offset.
func (r *PacienteFirestoreRepo) GetAll(ctx context.Context, limit, offset int) ([]models.Paciente, error) {
	activos := make([]models.Paciente, 0)
	iter := r.col.Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var p models.Paciente
		if err := doc.DataTo(&p); err != nil {
			return nil, err
		}
		p.ID = doc.Ref.ID
		if p.DeletedAt != nil {
			continue
		}
		activos = append(activos, p)
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
	return activos[offset:end], nil
}

// Update modifica un paciente activo existente en Firestore.
func (r *PacienteFirestoreRepo) Update(ctx context.Context, id string, p models.Paciente) (*models.Paciente, error) {
	docRef := r.col.Doc(id)
	doc, err := docRef.Get(ctx)
	if err != nil {
		return nil, err
	}
	if !doc.Exists() {
		return nil, fmt.Errorf("no se encontró paciente con ID %s", id)
	}
	var existing models.Paciente
	if err := doc.DataTo(&existing); err != nil {
		return nil, err
	}
	if existing.DeletedAt != nil {
		return nil, fmt.Errorf("no se encontró paciente con ID %s", id)
	}

	p.DeletedAt = nil
	if _, err := docRef.Set(ctx, p); err != nil {
		return nil, err
	}
	p.ID = id
	return &p, nil
}

// GetByID recupera un paciente activo por su ID desde Firestore.
func (r *PacienteFirestoreRepo) GetByID(ctx context.Context, id string) (*models.Paciente, error) {
	doc, err := r.col.Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}
	if !doc.Exists() {
		return nil, fmt.Errorf("no se encontró paciente con ID %s", id)
	}

	var p models.Paciente
	if err := doc.DataTo(&p); err != nil {
		return nil, err
	}
	if p.DeletedAt != nil {
		return nil, fmt.Errorf("no se encontró paciente con ID %s", id)
	}
	p.ID = doc.Ref.ID
	return &p, nil
}

// Delete realiza baja lógica (soft delete) de un paciente en Firestore marcando DeletedAt.
func (r *PacienteFirestoreRepo) Delete(ctx context.Context, id string) error {
	docRef := r.col.Doc(id)
	doc, err := docRef.Get(ctx)
	if err != nil {
		return err
	}
	if !doc.Exists() {
		return fmt.Errorf("no se encontró paciente con ID %s", id)
	}

	var p models.Paciente
	if err := doc.DataTo(&p); err != nil {
		return err
	}
	if p.DeletedAt != nil {
		return fmt.Errorf("no se encontró paciente con ID %s", id)
	}

	now := time.Now().UTC()
	p.DeletedAt = &now
	_, err = docRef.Set(ctx, p)
	return err
}

// Restore reactiva un paciente dado de baja lógica en Firestore limpiando DeletedAt.
func (r *PacienteFirestoreRepo) Restore(ctx context.Context, id string) (*models.Paciente, error) {
	docRef := r.col.Doc(id)
	doc, err := docRef.Get(ctx)
	if err != nil {
		return nil, err
	}
	if !doc.Exists() {
		return nil, fmt.Errorf("no se encontró paciente inactivo con ID %s", id)
	}

	var p models.Paciente
	if err := doc.DataTo(&p); err != nil {
		return nil, err
	}
	if p.DeletedAt == nil {
		return nil, fmt.Errorf("no se encontró paciente inactivo con ID %s", id)
	}

	p.DeletedAt = nil
	if _, err := docRef.Set(ctx, p); err != nil {
		return nil, err
	}
	p.ID = docRef.ID
	return &p, nil
}

// GetInactivos recupera los pacientes dados de baja lógica desde Firestore.
func (r *PacienteFirestoreRepo) GetInactivos(ctx context.Context) ([]models.Paciente, error) {
	inactivos := make([]models.Paciente, 0)
	iter := r.col.Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var p models.Paciente
		if err := doc.DataTo(&p); err != nil {
			return nil, err
		}
		if p.DeletedAt == nil {
			continue
		}
		p.ID = doc.Ref.ID
		inactivos = append(inactivos, p)
	}
	return inactivos, nil
}

// AddControl registra un nuevo control clínico en la subcolección "controles" de un paciente activo.
func (r *PacienteFirestoreRepo) AddControl(ctx context.Context, pacienteID string, control models.ControlClinico) (*models.ControlClinico, error) {
	if _, err := r.GetByID(ctx, pacienteID); err != nil {
		return nil, err
	}
	control.PacienteID = pacienteID
	docRef, _, err := r.col.Doc(pacienteID).Collection("controles").Add(ctx, control)
	if err != nil {
		return nil, err
	}
	control.ID = docRef.ID
	return &control, nil
}

// GetControles recupera los controles clínicos de un paciente activo ordenados cronológicamente (Fecha ascendente).
func (r *PacienteFirestoreRepo) GetControles(ctx context.Context, pacienteID string) ([]models.ControlClinico, error) {
	if _, err := r.GetByID(ctx, pacienteID); err != nil {
		return nil, err
	}
	controles := make([]models.ControlClinico, 0)
	iter := r.col.Doc(pacienteID).Collection("controles").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var c models.ControlClinico
		if err := doc.DataTo(&c); err != nil {
			return nil, err
		}
		c.ID = doc.Ref.ID
		controles = append(controles, c)
	}
	sort.SliceStable(controles, func(i, j int) bool {
		return controles[i].Fecha.Before(controles[j].Fecha)
	})
	return controles, nil
}

func (r *PacienteFirestoreRepo) obtenerControlExistente(ctx context.Context, pacienteID, controlID string) (*firestore.DocumentRef, models.ControlClinico, error) {
	if _, err := r.GetByID(ctx, pacienteID); err != nil {
		return nil, models.ControlClinico{}, err
	}
	docRef := r.col.Doc(pacienteID).Collection("controles").Doc(controlID)
	doc, err := docRef.Get(ctx)
	if status.Code(err) == codes.NotFound || (err == nil && !doc.Exists()) {
		return nil, models.ControlClinico{}, fmt.Errorf("%w con ID %s", ErrControlNoEncontrado, controlID)
	}
	if err != nil {
		return nil, models.ControlClinico{}, err
	}
	var existing models.ControlClinico
	if err := doc.DataTo(&existing); err != nil {
		return nil, models.ControlClinico{}, err
	}
	return docRef, existing, nil
}

// GetControlByID recupera un control clínico puntual de un paciente activo en Firestore.
func (r *PacienteFirestoreRepo) GetControlByID(ctx context.Context, pacienteID, controlID string) (*models.ControlClinico, error) {
	_, existing, err := r.obtenerControlExistente(ctx, pacienteID, controlID)
	if err != nil {
		return nil, err
	}
	existing.ID = controlID
	existing.PacienteID = pacienteID
	return &existing, nil
}

// UpdateControl actualiza un control clínico existente en la subcolección "controles" de un paciente activo.
// Si control.Fecha es cero, conserva la fecha original del registro.
func (r *PacienteFirestoreRepo) UpdateControl(ctx context.Context, pacienteID, controlID string, control models.ControlClinico) (*models.ControlClinico, error) {
	docRef, existing, err := r.obtenerControlExistente(ctx, pacienteID, controlID)
	if err != nil {
		return nil, err
	}
	if control.Fecha.IsZero() {
		control.Fecha = existing.Fecha
	}
	control.ID = controlID
	control.PacienteID = pacienteID
	if _, err := docRef.Set(ctx, control); err != nil {
		return nil, err
	}
	return &control, nil
}

// DeleteControl elimina un control clínico puntual en la subcolección "controles" de un paciente activo.
func (r *PacienteFirestoreRepo) DeleteControl(ctx context.Context, pacienteID, controlID string) error {
	docRef, _, err := r.obtenerControlExistente(ctx, pacienteID, controlID)
	if err != nil {
		return err
	}
	_, err = docRef.Delete(ctx)
	return err
}
