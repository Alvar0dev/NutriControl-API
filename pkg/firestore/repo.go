package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

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
	client, err := firestore.NewClient(ctx, projectID, option.WithEndpoint("firestore.googleapis.com:443"))
	if err != nil {
		return nil, err
	}

	return &PacienteFirestoreRepo{
		client: client,
		col:    client.Collection("pacientes"),
	}, nil
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

// GetAll recupera todos los pacientes de Firestore.
func (r *PacienteFirestoreRepo) GetAll(ctx context.Context) ([]models.Paciente, error) {
	var pacientes []models.Paciente
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
		p.ID = doc.Ref.ID
		if err := doc.DataTo(&p); err != nil {
			return nil, err
		}
		pacientes = append(pacientes, p)
	}
	return pacientes, nil
}
