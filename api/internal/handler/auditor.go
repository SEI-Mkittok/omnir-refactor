package handler

import (
	"net/http"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
)

// Auditor is a thin wrapper around AuditLogRepository for use in handlers.
type Auditor interface {
	logExport(r *http.Request, entity domain.AuditEntity)
}

type noopAuditor struct{}

func (n *noopAuditor) logExport(_ *http.Request, _ domain.AuditEntity) {}

type auditLogger struct {
	repo repository.AuditLogRepository
}

func newAuditor(repo repository.AuditLogRepository) Auditor {
	if repo == nil {
		return &noopAuditor{}
	}
	return &auditLogger{repo: repo}
}

func (a *auditLogger) logExport(r *http.Request, entity domain.AuditEntity) {
	entry := &domain.AuditLogEntry{
		Entity:     entity,
		Action:     domain.AuditActionExport,
		RemoteAddr: r.RemoteAddr,
	}
	// Best-effort — don't block the export on audit failure
	_ = a.repo.Create(r.Context(), entry)
}
