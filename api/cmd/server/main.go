package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/config"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/email"
	"github.com/omnir/crm-api/internal/handler"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository/postgres"
	"github.com/omnir/crm-api/internal/storage"
	"github.com/omnir/crm-api/internal/worker"
)

var Version = "dev"

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	logger := setupLogger(cfg.Env)

	db, err := postgres.NewOrgScopedPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		logger.Error("database ping failed", "err", err)
		os.Exit(1)
	}
	logger.Info("database connected", "org_mode", cfg.OrgMode)

	if cfg.OrgMode == config.OrgModeSaaS || cfg.OrgMode == config.OrgModeMultitenant || cfg.OrgMode == config.OrgModeEnterprise {
		if err := postgres.EnableRLS(context.Background(), db); err != nil {
			logger.Error("failed to enable RLS", "err", err)
			os.Exit(1)
		}
		logger.Info("row-level security enforced", "org_mode", cfg.OrgMode)
	}

	jwtSvc := auth.NewJWTService(cfg.JWTSecret)

	orgRepo := postgres.NewOrgRepo(db)
	contactRepo := postgres.NewContactRepo(db)
	accountRepo := postgres.NewAccountRepo(db)
	dealRepo := postgres.NewDealRepo(db)
	activityRepo := postgres.NewActivityRepo(db)
	noteRepo := postgres.NewNoteRepo(db)
	userRepo := postgres.NewUserRepo(db)
	notificationRepo := postgres.NewNotificationRepo(db)
	notifPrefRepo := postgres.NewNotificationPrefRepo(db)
	reportsRepo := postgres.NewReportsRepo(db)
	ticketRepo := postgres.NewTicketRepo(db)
	customFieldRepo := postgres.NewCustomFieldDefinitionRepo(db)
	ticketCommentRepo := postgres.NewTicketCommentRepo(db)
	ticketAttachmentRepo := postgres.NewTicketAttachmentRepo(db)
	slaPolicyRepo := postgres.NewSLAPolicyRepo(db)
	slaInstanceRepo := postgres.NewSLAInstanceRepo(db)
	leadRepo := postgres.NewLeadRepo(db)
	apiKeyRepo := postgres.NewAPIKeyRepo(db)
	emailRepo := postgres.NewEmailRepo(db)
	outboundWebhookRepo := postgres.NewOutboundWebhookRepo(db)
	portalLinkRepo := postgres.NewPortalLinkRepo(db)
	entityAttachmentRepo := postgres.NewEntityAttachmentRepo(db)
	auditLogRepo := postgres.NewAuditLogRepo(db)
	sequenceRepo := postgres.NewSequenceRepo(db)
	productRepo := postgres.NewProductRepo(db)
	quoteRepo := postgres.NewQuoteRepo(db)

	smtpSender := email.NewSender(cfg.SMTP)
	appURL := getEnv("APP_URL", "http://localhost:5173")
	mailer := email.NewMailer(smtpSender, appURL)

	workerCtx, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker()

	reminderWorker := worker.NewReminderWorker(notificationRepo, time.Minute, logger)
	reminderWorker.Start(workerCtx)

	emailNotifier := worker.NewEmailNotifier(mailer, logger)
	emailNotifier.Start(workerCtx, 3)

	webhookDispatcher := worker.NewWebhookDispatcher(outboundWebhookRepo, 30*time.Second, logger)
	webhookDispatcher.Start(workerCtx)

	slaBreachWorker := worker.NewSLABreachWorker(slaInstanceRepo, notificationRepo, 5*time.Minute, logger)
	slaBreachWorker.Start(workerCtx)

	if cfg.SMTP.Enabled {
		logger.Info("email notifications enabled", "smtp_host", cfg.SMTP.Host)
	} else {
		logger.Info("email notifications disabled")
	}

	setupHandler := handler.NewSetupHandler(userRepo, orgRepo, jwtSvc)
	orgHandler := handler.NewOrgHandler(orgRepo, userRepo, jwtSvc, cfg.OrgMode)
	authHandler := handler.NewAuthHandler(userRepo, jwtSvc).WithAuditLog(auditLogRepo)
	userHandler := handler.NewUserHandler(userRepo)
	contactHandler := handler.NewContactHandler(contactRepo).WithCustomFields(customFieldRepo).WithDeals(dealRepo)
	accountHandler := handler.NewAccountHandler(accountRepo).WithCustomFields(customFieldRepo)
	slaInstanceHandler := handler.NewSLAInstanceHandler(slaInstanceRepo)
	dealHandler := handler.NewDealHandler(dealRepo).WithCustomFields(customFieldRepo).WithNotifications(notificationRepo).WithSLA(slaPolicyRepo, slaInstanceRepo)
	activityHandler := handler.NewActivityHandler(activityRepo)
	notificationHandler := handler.NewNotificationHandler(notificationRepo)
	notifPrefHandler := handler.NewNotificationPrefHandler(notifPrefRepo)
	// Initialize file storage backend (S3-compatible or local fallback)
	var storageBackend storage.Backend
	if cfg.Storage.Backend == config.StorageBackendS3 {
		storageBackend, err = storage.NewS3Backend(
			cfg.Storage.S3Endpoint,
			cfg.Storage.S3Bucket,
			cfg.Storage.S3AccessKey,
			cfg.Storage.S3SecretKey,
			cfg.Storage.S3UseSSL,
		)
		if err != nil {
			logger.Error("failed to initialize S3 storage backend", "err", err)
			os.Exit(1)
		}
		logger.Info("storage: S3 backend", "endpoint", cfg.Storage.S3Endpoint, "bucket", cfg.Storage.S3Bucket)
	} else {
		basePath := cfg.Storage.LocalBasePath
		if basePath == "" {
			basePath = "./uploads"
		}
		storageBackend, err = storage.NewLocalBackend(basePath)
		if err != nil {
			logger.Error("failed to initialize local storage backend", "err", err)
			os.Exit(1)
		}
		logger.Info("storage: local backend", "path", basePath)
	}

	ticketHandler := handler.NewTicketHandler(ticketRepo, ticketCommentRepo, ticketAttachmentRepo, storageBackend)
	portalHandler := handler.NewPortalHandler(ticketRepo, ticketCommentRepo)
	slaPolicyHandler := handler.NewSLAPolicyHandler(slaPolicyRepo)
	inboundWebhookHandler := handler.NewWebhookHandler(ticketRepo, ticketCommentRepo, contactRepo, userRepo, cfg.WebhookSecret, cfg.OrgMode, logger)
	contactNoteHandler := handler.NewNoteHandler(noteRepo, domain.NoteEntityContact, "id")
	accountNoteHandler := handler.NewNoteHandler(noteRepo, domain.NoteEntityAccount, "id")
	dealNoteHandler := handler.NewNoteHandler(noteRepo, domain.NoteEntityDeal, "id")
	leadNoteHandler := handler.NewNoteHandler(noteRepo, domain.NoteEntityLead, "id")
	searchRepo := postgres.NewSearchRepo(db)
	searchHandler := handler.NewSearchHandler(searchRepo)
	reportsHandler := handler.NewReportsHandler(reportsRepo)
	exportHandler := handler.NewExportHandler(contactRepo, accountRepo, dealRepo, reportsRepo).WithAuditLog(auditLogRepo)
	leadHandler := handler.NewLeadHandler(leadRepo, contactRepo)
	customFieldHandler := handler.NewCustomFieldHandler(customFieldRepo)
	apiKeyHandler := handler.NewAPIKeyHandler(apiKeyRepo)
	emailHandler := handler.NewEmailHandler(emailRepo, mailer, cfg.SMTP.From)
	importHandler := handler.NewImportHandler(contactRepo, accountRepo, leadRepo)
	outboundWebhookHandler := handler.NewOutboundWebhookHandler(outboundWebhookRepo)
	savedViewRepo := postgres.NewSavedViewRepo(db)
	savedViewHandler := handler.NewSavedViewHandler(savedViewRepo)
	inboundEmailHandler := handler.NewInboundEmailHandler(emailRepo, contactRepo, cfg.WebhookSecret, cfg.OrgMode)
	dealPortalLinksHandler := handler.NewDealPortalLinksHandler(portalLinkRepo, dealRepo, noteRepo, orgRepo)

	uploadsDir := getEnv("UPLOADS_DIR", "uploads")
	contactAttachmentHandler := handler.NewEntityAttachmentHandler(entityAttachmentRepo, uploadsDir, domain.EntityTypeContact, "id")
	accountAttachmentHandler := handler.NewEntityAttachmentHandler(entityAttachmentRepo, uploadsDir, domain.EntityTypeAccount, "id")
	dealAttachmentHandler := handler.NewEntityAttachmentHandler(entityAttachmentRepo, uploadsDir, domain.EntityTypeDeal, "id")
	attachmentDownloadHandler := handler.NewAttachmentDownloadHandler(entityAttachmentRepo)
	sequenceHandler := handler.NewSequenceHandler(sequenceRepo)
	sequenceTrackingHandler := handler.NewSequenceTrackingHandler(sequenceRepo, contactRepo, cfg.SequenceTokenSecret)
	auditLogHandler := handler.NewAuditLogHandler(auditLogRepo)
	productHandler := handler.NewProductHandler(productRepo)
	quoteHandler := handler.NewQuoteHandler(quoteRepo).WithMailer(mailer, cfg.SMTP.From)
	sequenceWorker := worker.NewSequenceWorker(sequenceRepo, mailer, cfg.SequenceTokenSecret, time.Minute, logger)
	sequenceWorker.Start(workerCtx)

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Logger(logger))
	r.Use(chimiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","version":"` + Version + `"}`))
	})

	r.Mount("/api/setup", setupHandler.Router())
	r.Mount("/api/orgs", orgHandler.Router())
	r.Mount("/api/auth", authHandler.Router())
	r.Mount("/webhooks/email", inboundWebhookHandler.Router())
	r.Mount("/api/emails/inbound", inboundEmailHandler.Router())
	// Public deal portal — token IS the credential, no JWT required.
	r.Mount("/api/portal", dealPortalLinksHandler.PublicRouter())
	// Public sequence tracking — HMAC-signed tokens, no JWT required.
	r.Mount("/track", sequenceTrackingHandler.TrackRouter())
	r.Mount("/unsubscribe", sequenceTrackingHandler.UnsubscribeRouter())
	// Bounce webhook — outside /api/v1 auth group, accepts webhook provider calls.
	r.Mount("/api/emails/bounce", sequenceTrackingHandler.BounceRouter())

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(chimiddleware.Timeout(30 * time.Second))
		r.Use(middleware.Authenticate(jwtSvc, apiKeyRepo, userRepo))
		r.Use(middleware.OrgScope(cfg.OrgMode))
		r.Mount("/contacts", contactHandler.Router())
		r.Mount("/leads", leadHandler.Router())
		r.Route("/leads/{id}/notes", func(r chi.Router) {
			r.Mount("/", leadNoteHandler.Router())
		})
		r.Route("/contacts/{id}/notes", func(r chi.Router) {
			r.Mount("/", contactNoteHandler.Router())
		})
		r.Route("/contacts/{id}/emails", func(r chi.Router) {
			r.Mount("/", emailHandler.ContactEmailRouter())
		})
		r.Mount("/accounts", accountHandler.Router())
		r.Route("/accounts/{id}/notes", func(r chi.Router) { r.Mount("/", accountNoteHandler.Router()) })
		r.Mount("/deals", dealHandler.Router())
		r.Route("/deals/{id}/notes", func(r chi.Router) { r.Mount("/", dealNoteHandler.Router()) })
		r.Route("/deals/{dealId}/portal-links", func(r chi.Router) { r.Mount("/", dealPortalLinksHandler.AuthRouter()) })
		r.Mount("/portal-links", dealPortalLinksHandler.RevokeRouter())
		r.Mount("/activities", activityHandler.Router())
		r.Mount("/notifications", notificationHandler.Router())
		r.Mount("/tickets", ticketHandler.Router())
		r.Mount("/portal", portalHandler.Router())
		r.Mount("/sla-policies", slaPolicyHandler.Router())
		r.Mount("/sla-instances", slaInstanceHandler.Router())
		r.Mount("/sla-dashboard", slaInstanceHandler.DashboardRouter())
		r.Mount("/users", userHandler.Router())
		r.Route("/users/me/notification-prefs", func(r chi.Router) { r.Mount("/", notifPrefHandler.Router()) })
		r.Mount("/search", searchHandler.Router())
		r.Mount("/reports", reportsHandler.Router())
		r.Mount("/export", exportHandler.Router())
		r.Mount("/custom-fields", customFieldHandler.Router())
		r.Mount("/api-keys", apiKeyHandler.Router())
		r.Mount("/emails", emailHandler.Router())
		r.Mount("/webhooks", outboundWebhookHandler.Router())
		r.Mount("/sequences", sequenceHandler.Router())
		r.Route("/contacts/{id}/attachments", func(r chi.Router) { r.Mount("/", contactAttachmentHandler.Router()) })
		r.Route("/accounts/{id}/attachments", func(r chi.Router) { r.Mount("/", accountAttachmentHandler.Router()) })
		r.Route("/deals/{id}/attachments", func(r chi.Router) { r.Mount("/", dealAttachmentHandler.Router()) })
		r.Mount("/attachments", attachmentDownloadHandler.Router())
		r.Mount("/admin/audit-log", auditLogHandler.Router())
		r.Mount("/views", savedViewHandler.Router())
		r.Mount("/products", productHandler.Router())
		r.Mount("/quotes", quoteHandler.Router())
		r.Route("/deals/{dealId}/quotes", func(r chi.Router) { r.Mount("/", quoteHandler.DealQuotesRouter()) })
	})

	r.Group(func(r chi.Router) {
		r.Use(chimiddleware.Timeout(5 * time.Minute))
		r.Use(middleware.Authenticate(jwtSvc, apiKeyRepo, userRepo))
		r.Use(middleware.OrgScope(cfg.OrgMode))
		r.Mount("/api/v1/import", importHandler.Router())
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("server starting", "addr", srv.Addr, "version", Version)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-done
	logger.Info("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("shutdown error", "err", err)
	}
	logger.Info("server stopped")
}

func setupLogger(env string) *slog.Logger {
	level := slog.LevelInfo
	if env == "development" {
		level = slog.LevelDebug
	}
	opts := &slog.HandlerOptions{Level: level}
	var h slog.Handler
	if env == "production" {
		h = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		h = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(h)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
