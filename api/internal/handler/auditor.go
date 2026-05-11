package handler

import (
	"context"
	"log/slog"
	"net/http"

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

func (a *Auditor) logLogin(r *http.Request, user *domain.User) {
	if a.repo == nil || user == nil {
		return
	}

	uid := user.ID
	name := user.Email
	entry := domain.AuditEntry{
		OrgID:      user.OrgID,
		UserID:     &uid,
		Action:     domain.AuditActionLogin,
		EntityType: domain.AuditEntityUser,
		EntityID:   &uid,
		EntityName: &name,
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
