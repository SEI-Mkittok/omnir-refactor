package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/handler"
)

type fakeEntityAttachmentRepository struct {
	attachment *domain.EntityAttachment
	err        error
}

func (f *fakeEntityAttachmentRepository) Create(context.Context, *domain.EntityAttachment) (*domain.EntityAttachment, error) {
	return nil, nil
}

func (f *fakeEntityAttachmentRepository) List(context.Context, domain.EntityType, uuid.UUID) ([]*domain.EntityAttachment, error) {
	return nil, nil
}

func (f *fakeEntityAttachmentRepository) GetByID(context.Context, uuid.UUID) (*domain.EntityAttachment, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.attachment, nil
}

func (f *fakeEntityAttachmentRepository) Delete(context.Context, uuid.UUID, domain.EntityType, uuid.UUID) error {
	return nil
}

type attachmentRecordAccessCall struct {
	module domain.ACLModule
	id     uuid.UUID
	access domain.SharingAccessLevel
}

type fakeAttachmentRecordAccessRepository struct {
	allowed bool
	calls   []attachmentRecordAccessCall
}

func (f *fakeAttachmentRecordAccessRepository) CanAccessRecord(_ context.Context, module domain.ACLModule, id uuid.UUID, access domain.SharingAccessLevel) (bool, error) {
	f.calls = append(f.calls, attachmentRecordAccessCall{module: module, id: id, access: access})
	return f.allowed, nil
}

func TestAttachmentDownloadRequiresParentReadVisibility(t *testing.T) {
	attachmentID := uuid.New()
	contactID := uuid.New()
	repo := &fakeEntityAttachmentRepository{attachment: &domain.EntityAttachment{
		ID:          attachmentID,
		EntityType:  domain.EntityTypeContact,
		EntityID:    contactID,
		Filename:    "private.txt",
		ContentType: "text/plain",
		StoragePath: filepath.Join(t.TempDir(), "private.txt"),
		CreatedAt:   time.Now(),
	}}
	accessRepo := &fakeAttachmentRecordAccessRepository{allowed: false}
	h := handler.NewAttachmentDownloadHandler(repo, accessRepo)

	req := httptest.NewRequest(http.MethodGet, "/"+attachmentID.String(), nil)
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleContacts: {domain.ACLActionRead: true},
		},
	}))
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
	require.Len(t, accessRepo.calls, 1)
	require.Equal(t, domain.ACLModuleContacts, accessRepo.calls[0].module)
	require.Equal(t, contactID, accessRepo.calls[0].id)
	require.Equal(t, domain.SharingAccessRead, accessRepo.calls[0].access)
}

func TestAttachmentDownloadRequiresParentModuleReadPermission(t *testing.T) {
	attachmentID := uuid.New()
	repo := &fakeEntityAttachmentRepository{attachment: &domain.EntityAttachment{
		ID:          attachmentID,
		EntityType:  domain.EntityTypeDeal,
		EntityID:    uuid.New(),
		Filename:    "quote.txt",
		ContentType: "text/plain",
		StoragePath: filepath.Join(t.TempDir(), "quote.txt"),
		CreatedAt:   time.Now(),
	}}
	accessRepo := &fakeAttachmentRecordAccessRepository{allowed: true}
	h := handler.NewAttachmentDownloadHandler(repo, accessRepo)

	req := httptest.NewRequest(http.MethodGet, "/"+attachmentID.String(), nil)
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{},
	}))
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
	require.Empty(t, accessRepo.calls)
}

func TestAttachmentDownloadServesFileWhenParentReadIsAllowed(t *testing.T) {
	attachmentID := uuid.New()
	accountID := uuid.New()
	path := filepath.Join(t.TempDir(), "allowed.txt")
	require.NoError(t, os.WriteFile(path, []byte("hello attachment"), 0o600))
	repo := &fakeEntityAttachmentRepository{attachment: &domain.EntityAttachment{
		ID:          attachmentID,
		EntityType:  domain.EntityTypeAccount,
		EntityID:    accountID,
		Filename:    "allowed.txt",
		ContentType: "text/plain",
		StoragePath: path,
		CreatedAt:   time.Now(),
	}}
	accessRepo := &fakeAttachmentRecordAccessRepository{allowed: true}
	h := handler.NewAttachmentDownloadHandler(repo, accessRepo)

	req := httptest.NewRequest(http.MethodGet, "/"+attachmentID.String(), nil)
	req = req.WithContext(domain.WithAccessContext(req.Context(), &domain.AccessContext{
		Permissions: map[domain.ACLModule]map[domain.ACLAction]bool{
			domain.ACLModuleAccounts: {domain.ACLActionRead: true},
		},
	}))
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "hello attachment", w.Body.String())
	require.Len(t, accessRepo.calls, 1)
	require.Equal(t, domain.ACLModuleAccounts, accessRepo.calls[0].module)
	require.Equal(t, accountID, accessRepo.calls[0].id)
}
