package firestore

import (
	"context"
	"errors"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/firestore/apiv1/firestorepb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"nutricontrol/models"
)

// fakeFirestoreServer implementa los RPCs básicos de Firestore (Commit, BatchGetDocuments, RunQuery)
// en memoria para ejercitar PacienteFirestoreRepo con un *firestore.Client real sin dependencias externas.
type fakeFirestoreServer struct {
	firestorepb.UnimplementedFirestoreServer
	mu    sync.Mutex
	docs  map[string]*firestorepb.Document
	order []string
}

func newFakeFirestoreServer() *fakeFirestoreServer {
	return &fakeFirestoreServer{
		docs: make(map[string]*firestorepb.Document),
	}
}

func (s *fakeFirestoreServer) Commit(_ context.Context, req *firestorepb.CommitRequest) (*firestorepb.CommitResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := timestamppb.Now()
	results := make([]*firestorepb.WriteResult, 0, len(req.GetWrites()))

	for _, w := range req.GetWrites() {
		if upd := w.GetUpdate(); upd != nil {
			name := upd.GetName()
			existing, exists := s.docs[name]
			if pre := w.GetCurrentDocument(); pre != nil {
				switch cond := pre.GetConditionType().(type) {
				case *firestorepb.Precondition_Exists:
					if cond.Exists && !exists {
						return nil, status.Errorf(codes.NotFound, "document %s does not exist", name)
					}
					if !cond.Exists && exists {
						return nil, status.Errorf(codes.AlreadyExists, "document %s already exists", name)
					}
				}
			}

			stored := proto.Clone(upd).(*firestorepb.Document)
			if exists && existing.GetCreateTime() != nil {
				stored.CreateTime = existing.GetCreateTime()
			} else {
				stored.CreateTime = now
			}
			stored.UpdateTime = now

			if !exists {
				s.order = append(s.order, name)
			}
			s.docs[name] = stored
			results = append(results, &firestorepb.WriteResult{UpdateTime: now})
			continue
		}

		if del := w.GetDelete(); del != "" {
			delete(s.docs, del)
			for i, n := range s.order {
				if n == del {
					s.order = append(s.order[:i], s.order[i+1:]...)
					break
				}
			}
			results = append(results, &firestorepb.WriteResult{UpdateTime: now})
		}
	}

	return &firestorepb.CommitResponse{
		WriteResults: results,
		CommitTime:   now,
	}, nil
}

func (s *fakeFirestoreServer) BatchGetDocuments(req *firestorepb.BatchGetDocumentsRequest, stream firestorepb.Firestore_BatchGetDocumentsServer) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := timestamppb.Now()
	for _, name := range req.GetDocuments() {
		if doc, ok := s.docs[name]; ok {
			if err := stream.Send(&firestorepb.BatchGetDocumentsResponse{
				Result:   &firestorepb.BatchGetDocumentsResponse_Found{Found: proto.Clone(doc).(*firestorepb.Document)},
				ReadTime: now,
			}); err != nil {
				return err
			}
			continue
		}
		if err := stream.Send(&firestorepb.BatchGetDocumentsResponse{
			Result:   &firestorepb.BatchGetDocumentsResponse_Missing{Missing: name},
			ReadTime: now,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *fakeFirestoreServer) RunQuery(req *firestorepb.RunQueryRequest, stream firestorepb.Firestore_RunQueryServer) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := timestamppb.Now()
	sq := req.GetStructuredQuery()
	var colID string
	if sq != nil && len(sq.GetFrom()) > 0 {
		colID = sq.GetFrom()[0].GetCollectionId()
	}
	prefix := req.GetParent() + "/" + colID + "/"

	sent := 0
	for _, name := range s.order {
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		rest := strings.TrimPrefix(name, prefix)
		if strings.Contains(rest, "/") {
			continue
		}
		doc := s.docs[name]
		if err := stream.Send(&firestorepb.RunQueryResponse{
			Document: proto.Clone(doc).(*firestorepb.Document),
			ReadTime: now,
		}); err != nil {
			return err
		}
		sent++
	}

	if sent == 0 {
		return stream.Send(&firestorepb.RunQueryResponse{
			ReadTime:             now,
			ContinuationSelector: &firestorepb.RunQueryResponse_Done{Done: true},
		})
	}
	return nil
}

func newTestFirestoreRepo(t *testing.T) *PacienteFirestoreRepo {
	t.Helper()
	ctx := context.Background()

	if os.Getenv("FIRESTORE_EMULATOR_HOST") != "" {
		repo, err := NewPacienteFirestoreRepo(ctx, "test-project")
		require.NoError(t, err)
		t.Cleanup(func() { _ = repo.client.Close() })
		return repo
	}

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	srv := grpc.NewServer()
	firestorepb.RegisterFirestoreServer(srv, newFakeFirestoreServer())
	go func() { _ = srv.Serve(lis) }()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	client, err := firestore.NewClient(ctx, "test-project", option.WithGRPCConn(conn))
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = client.Close()
		_ = conn.Close()
		srv.Stop()
		_ = lis.Close()
	})

	return NewPacienteFirestoreRepoWithClient(client)
}

func TestFirestoreRepoCRUDYGetAllPaginado(t *testing.T) {
	repo := newTestFirestoreRepo(t)
	ctx := context.Background()

	// Repositorio vacío
	inicial, err := repo.GetAll(ctx, 0, 0)
	require.NoError(t, err)
	assert.Len(t, inicial, 0)

	// Create y GetByID
	p1, err := repo.Create(ctx, models.Paciente{
		Nombre:          "Ana Torres",
		Edad:            29,
		Peso:            64.5,
		Altura:          1.68,
		PorcentajeGrasa: 21.0,
		Notas:           "Control inicial",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, p1.ID)
	assert.Nil(t, p1.DeletedAt)

	p2, err := repo.Create(ctx, models.Paciente{Nombre: "Bruno Diaz", Edad: 34, Peso: 80.0})
	require.NoError(t, err)
	p3, err := repo.Create(ctx, models.Paciente{Nombre: "Clara Vera", Edad: 41, Peso: 68.0})
	require.NoError(t, err)

	obtenido, err := repo.GetByID(ctx, p1.ID)
	require.NoError(t, err)
	assert.Equal(t, "Ana Torres", obtenido.Nombre)
	assert.Equal(t, 1.68, obtenido.Altura)
	assert.Nil(t, obtenido.DeletedAt)

	// GetAll con limit y offset
	pagina, err := repo.GetAll(ctx, 2, 1)
	require.NoError(t, err)
	require.Len(t, pagina, 2)
	assert.Equal(t, p2.ID, pagina[0].ID)
	assert.Equal(t, p3.ID, pagina[1].ID)

	// Offset fuera de rango
	vacio, err := repo.GetAll(ctx, 5, 10)
	require.NoError(t, err)
	assert.Len(t, vacio, 0)

	// Update sobre paciente activo
	actualizado, err := repo.Update(ctx, p1.ID, models.Paciente{
		Nombre:          "Ana Torres Editada",
		Edad:            30,
		Peso:            63.0,
		Altura:          1.68,
		PorcentajeGrasa: 20.0,
		Notas:           "Actualizado",
	})
	require.NoError(t, err)
	assert.Equal(t, "Ana Torres Editada", actualizado.Nombre)
	assert.Equal(t, 63.0, actualizado.Peso)

	// GetByID y Update sobre ID inexistente fallan
	_, err = repo.GetByID(ctx, "inexistente")
	assert.Error(t, err)
	_, err = repo.Update(ctx, "inexistente", models.Paciente{Nombre: "X", Edad: 20, Peso: 60})
	assert.Error(t, err)
}

func TestFirestoreRepoSoftDeleteYRestore(t *testing.T) {
	repo := newTestFirestoreRepo(t)
	ctx := context.Background()

	p1, err := repo.Create(ctx, models.Paciente{Nombre: "Activo 1", Edad: 25, Peso: 70})
	require.NoError(t, err)
	p2, err := repo.Create(ctx, models.Paciente{Nombre: "Para Baja", Edad: 31, Peso: 75})
	require.NoError(t, err)
	p3, err := repo.Create(ctx, models.Paciente{Nombre: "Activo 2", Edad: 40, Peso: 82})
	require.NoError(t, err)

	// Soft delete de p2
	require.NoError(t, repo.Delete(ctx, p2.ID))

	// Excluido de GetByID, Update y segunda eliminación
	_, err = repo.GetByID(ctx, p2.ID)
	assert.Error(t, err)
	_, err = repo.Update(ctx, p2.ID, models.Paciente{Nombre: "No Debe", Edad: 31, Peso: 75})
	assert.Error(t, err)
	err = repo.Delete(ctx, p2.ID)
	assert.Error(t, err)

	// GetAll filtra DeletedAt != nil antes de paginar: limit=2 offset=0 -> [p1, p3]
	activos, err := repo.GetAll(ctx, 2, 0)
	require.NoError(t, err)
	require.Len(t, activos, 2)
	assert.Equal(t, p1.ID, activos[0].ID)
	assert.Equal(t, p3.ID, activos[1].ID)

	// GetInactivos devuelve únicamente p2 con DeletedAt != nil
	inactivos, err := repo.GetInactivos(ctx)
	require.NoError(t, err)
	require.Len(t, inactivos, 1)
	assert.Equal(t, p2.ID, inactivos[0].ID)
	assert.NotNil(t, inactivos[0].DeletedAt)

	// Restore reactiva p2
	restaurado, err := repo.Restore(ctx, p2.ID)
	require.NoError(t, err)
	assert.Nil(t, restaurado.DeletedAt)

	recuperado, err := repo.GetByID(ctx, p2.ID)
	require.NoError(t, err)
	assert.Equal(t, "Para Baja", recuperado.Nombre)

	// Restore sobre activo o inexistente falla
	_, err = repo.Restore(ctx, p2.ID)
	assert.Error(t, err)
	_, err = repo.Restore(ctx, "no-existe")
	assert.Error(t, err)
}

func TestFirestoreRepoControlesClinicos(t *testing.T) {
	repo := newTestFirestoreRepo(t)
	ctx := context.Background()

	p, err := repo.Create(ctx, models.Paciente{Nombre: "Seguimiento", Edad: 28, Peso: 80, Altura: 1.80})
	require.NoError(t, err)

	// Sin controles inicialmente
	vacio, err := repo.GetControles(ctx, p.ID)
	require.NoError(t, err)
	assert.Len(t, vacio, 0)

	// Registrar en orden inverso y verificar orden cronológico ascendente
	f2 := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	f1 := time.Date(2026, 10, 15, 10, 0, 0, 0, time.UTC)

	c2, err := repo.AddControl(ctx, p.ID, models.ControlClinico{Fecha: f2, Peso: 77.5, PorcentajeGrasa: 18.2, IMC: 23.92, Notas: "Mes 2"})
	require.NoError(t, err)
	assert.NotEmpty(t, c2.ID)
	assert.Equal(t, p.ID, c2.PacienteID)

	c1, err := repo.AddControl(ctx, p.ID, models.ControlClinico{Fecha: f1, Peso: 80.0, PorcentajeGrasa: 20.0, IMC: 24.69, Notas: "Mes 1"})
	require.NoError(t, err)
	assert.NotEmpty(t, c1.ID)

	controles, err := repo.GetControles(ctx, p.ID)
	require.NoError(t, err)
	require.Len(t, controles, 2)
	assert.True(t, controles[0].Fecha.Equal(f1))
	assert.Equal(t, "Mes 1", controles[0].Notas)
	assert.True(t, controles[1].Fecha.Equal(f2))
	assert.Equal(t, "Mes 2", controles[1].Notas)

	// Soft delete bloquea lectura y creación de controles en subcolección
	require.NoError(t, repo.Delete(ctx, p.ID))
	_, err = repo.GetControles(ctx, p.ID)
	assert.Error(t, err)
	_, err = repo.AddControl(ctx, p.ID, models.ControlClinico{Fecha: f2, Peso: 76.0})
	assert.Error(t, err)
}

func TestFirestoreRepoUpdateYDeleteControl(t *testing.T) {
	repo := newTestFirestoreRepo(t)
	ctx := context.Background()

	p, err := repo.Create(ctx, models.Paciente{Nombre: "Firestore Control CRUD", Edad: 31, Peso: 80, Altura: 1.80})
	require.NoError(t, err)

	f1 := time.Date(2026, 10, 10, 10, 0, 0, 0, time.UTC)
	f2 := time.Date(2026, 10, 14, 10, 0, 0, 0, time.UTC)
	c1, err := repo.AddControl(ctx, p.ID, models.ControlClinico{
		Fecha: f1, Peso: 80.0, PorcentajeGrasa: 20.0, IMC: 24.69, Notas: "Inicial",
	})
	require.NoError(t, err)

	// Update con fecha cero conserva fecha original
	act1, err := repo.UpdateControl(ctx, p.ID, c1.ID, models.ControlClinico{
		Peso: 78.5, PorcentajeGrasa: 19.2, IMC: 24.23, Notas: "Editado",
	})
	require.NoError(t, err)
	assert.Equal(t, c1.ID, act1.ID)
	assert.Equal(t, p.ID, act1.PacienteID)
	assert.True(t, act1.Fecha.Equal(f1))
	assert.Equal(t, 78.5, act1.Peso)
	assert.Equal(t, "Editado", act1.Notas)

	// Update con nueva fecha actualiza fecha y valores
	act2, err := repo.UpdateControl(ctx, p.ID, c1.ID, models.ControlClinico{
		Fecha: f2, Peso: 77.0, PorcentajeGrasa: 18.5, IMC: 23.77, Notas: "Nueva fecha",
	})
	require.NoError(t, err)
	assert.True(t, act2.Fecha.Equal(f2))

	controles, err := repo.GetControles(ctx, p.ID)
	require.NoError(t, err)
	require.Len(t, controles, 1)
	assert.Equal(t, 77.0, controles[0].Peso)
	assert.True(t, controles[0].Fecha.Equal(f2))

	// Control inexistente devuelve ErrControlNoEncontrado
	_, err = repo.UpdateControl(ctx, p.ID, "inexistente", models.ControlClinico{Peso: 75})
	assert.True(t, errors.Is(err, ErrControlNoEncontrado))
	err = repo.DeleteControl(ctx, p.ID, "inexistente")
	assert.True(t, errors.Is(err, ErrControlNoEncontrado))

	// Delete elimina el documento de subcolección
	require.NoError(t, repo.DeleteControl(ctx, p.ID, c1.ID))
	controlesPost, err := repo.GetControles(ctx, p.ID)
	require.NoError(t, err)
	assert.Empty(t, controlesPost)

	err = repo.DeleteControl(ctx, p.ID, c1.ID)
	assert.True(t, errors.Is(err, ErrControlNoEncontrado))

	// Soft delete del paciente bloquea edición y borrado de controles
	c2, err := repo.AddControl(ctx, p.ID, models.ControlClinico{Fecha: f1, Peso: 76.5})
	require.NoError(t, err)
	require.NoError(t, repo.Delete(ctx, p.ID))
	_, err = repo.UpdateControl(ctx, p.ID, c2.ID, models.ControlClinico{Peso: 75.0})
	assert.Error(t, err)
	assert.Error(t, repo.DeleteControl(ctx, p.ID, c2.ID))
}
