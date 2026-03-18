package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

// Auditor is a helper embedded in handlers to write audit log entries.
// Audit writes are fire-and-forget (async) so they never block request handling.
type Auditor struct {
	repo repository.AuditLogRepository
}

func newAuditor(repo repository.AuditLogRepository) Auditor {
	return Auditor{repo: repo}
}

// log writes an audit entry asynchronously. Errors are logged but not returned.
func (a *Auditor) log(r *http.Request, action domain.AuditAction, entityType domain.AuditEntityType, entityID *uuid.UUID, entityName *string, changes domain.AuditChanges) {
	if a.repo == nil {
		return
	}

	orgID, _ := domain.OrgIDFromContext(r.Context())

	entry := domain.AuditEntry{
		OrgID:      orgID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		EntityName: entityName,
		Changes:    changes,
	}

	if claims, ok := middleware.ClaimsFromContext(r); ok {
		uid := claims.UserID
		entry.UserID = &uid
	}

	if ip := realClientIP(r); ip != "" {
		entry.IPAddress = &ip
	}
	if ua := r.UserAgent(); ua != "" {
		entry.UserAgent = &ua
	}

	go func() {
		if err := a.repo.Append(context.Background(), entry); err != nil {
			slog.Error("audit log write failed", "error", err)
		}
	}()
}

func realClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		for i, c := range ip {
			if c == ',' {
				return ip[:i]
			}
		}
		return ip
	}
	addr := r.RemoteAddr
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i]
		}
	}
	return addr
}


// logLogin writes a login audit entry (success or failure) asynchronously.
func (a *Auditor) logLogin(r *http.Request, userID *uuid.UUID, success bool) {
	if a.repo == nil {
		return
	}

	orgID, _ := domain.OrgIDFromContext(r.Context())
	result := "failure"
	if success {
		result = "success"
	}

	entry := domain.AuditEntry{
		OrgID:      orgID,
		Action:     domain.AuditActionLogin,
		EntityType: domain.AuditEntityUser,
		EntityID:   userID,
		UserID:     userID,
		EntityName: &result,
	}

	if ip := realClientIP(r); ip != "" {
		entry.IPAddress = &ip
	}
	if ua := r.UserAgent(); ua != "" {
		entry.UserAgent = &ua
	}

	go func() {
		if err := a.repo.Append(context.Background(), entry); err != nil {
			slog.Error("audit log write failed", "action", "login", "error", err)
		}
	}()
}

// logExport writes an export audit entry asynchronously.
func (a *Auditor) logExport(r *http.Request, entityType domain.AuditEntityType) {
	if a.repo == nil {
		return
	}

	orgID, _ := domain.OrgIDFromContext(r.Context())

	entry := domain.AuditEntry{
		OrgID:      orgID,
		Action:     domain.AuditActionExport,
		EntityType: entityType,
	}

	if claims, ok := middleware.ClaimsFromContext(r); ok {
		uid := claims.UserID
		entry.UserID = &uid
	}

	if ip := realClientIP(r); ip != "" {
		entry.IPAddress = &ip
	}
	if ua := r.UserAgent(); ua != "" {
		entry.UserAgent = &ua
	}

	go func() {
		if err := a.repo.Append(context.Background(), entry); err != nil {
			slog.Error("audit log write failed", "action", "export", "error", err)
		}
	}()
}

func idPtr(id uuid.UUID) *uuid.UUID { return &id }
