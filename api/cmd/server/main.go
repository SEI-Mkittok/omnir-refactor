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
	"github.com/go-chi/httprate"
	"github.com/joho/godotenv"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/config"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/email"
	enrichmentpkg "github.com/omnir/crm-api/internal/enrichment"
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
	userPrefRepo := postgres.NewUserPreferenceRepo(db)
	reportsRepo := postgres.NewReportsRepo(db)
	ticketRepo := postgres.NewTicketRepo(db)
	customFieldRepo := postgres.NewCustomFieldDefinitionRepo(db)
	moduleLayoutRepo := postgres.NewModuleLayoutRepo(db)
	moduleRelationshipDefinitionRepo := postgres.NewModuleRelationshipDefinitionRepo(db)
	crmEntityLinkRepo := postgres.NewCRMEntityLinkRepo(db)
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
	automationRepo := postgres.NewAutomationRepo(db)
	calendarConnectionRepo := postgres.NewCalendarConnectionRepo(db)
	emailConnectionRepo := postgres.NewEmailConnectionRepo(db)
	integrationCredRepo := postgres.NewIntegrationCredentialRepo(db)
	orgSettingsRepo := postgres.NewOrgSettingsRepo(db)
	currencyRepo := postgres.NewCurrencyRepo(db)
	picklistRepo := postgres.NewPicklistRepo(db)
	leadConversionMappingRepo := postgres.NewLeadConversionMappingRepo(db)
	emailInboxRepo := postgres.NewEmailInboxRepo(db)
	ssoConfigRepo := postgres.NewSSOConfigRepo(db)
	totpRepo := postgres.NewTOTPRepo(db)
	enrichmentCacheRepo := postgres.NewEnrichmentCacheRepo(db)
	kbArticleRepo := postgres.NewKBArticleRepo(db)
	kbCategoryRepo := postgres.NewKBCategoryRepo(db)
	teamsConnectionRepo := postgres.NewTeamsConnectionRepo(db)
	billingRepo := postgres.NewBillingRepo(db)
	dashboardRepo := postgres.NewDashboardRepo(db)
	onboardingRepo := postgres.NewOnboardingRepo(db)
	emailTemplateRepo := postgres.NewEmailTemplateRepo(db)
	opsFinanceRepo := postgres.NewOperationsFinanceRepo(db)
	accessRepo := postgres.NewAccessRepo(db)

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

	automationWorker := worker.NewAutomationWorker(
		automationRepo, activityRepo, contactRepo, dealRepo, sequenceRepo,
		mailer, time.Minute, logger,
	)
	automationWorker.Start(workerCtx)

	calendarSyncWorker := worker.NewCalendarSyncWorker(
		calendarConnectionRepo, activityRepo,
		5*time.Minute, logger,
		cfg.Calendar.GoogleClientID, cfg.Calendar.GoogleClientSecret,
		cfg.Calendar.MicrosoftClientID, cfg.Calendar.MicrosoftClientSecret, cfg.Calendar.MicrosoftTenantID,
	)
	calendarSyncWorker.Start(workerCtx)

	emailInboxSyncWorker := worker.NewEmailInboxSyncWorker(
		emailConnectionRepo, emailInboxRepo, contactRepo,
		5*time.Minute, logger,
		cfg.EmailInbox.EncryptionKey,
		cfg.EmailInbox.GoogleClientID, cfg.EmailInbox.GoogleClientSecret,
		cfg.EmailInbox.MicrosoftClientID, cfg.EmailInbox.MicrosoftClientSecret, cfg.EmailInbox.MicrosoftTenantID,
	)
	emailInboxSyncWorker.Start(workerCtx)

	if cfg.SMTP.Enabled {
		logger.Info("email notifications enabled", "smtp_host", cfg.SMTP.Host)
	} else {
		logger.Info("email notifications disabled")
	}

	teamsNotifier := worker.NewTeamsNotifier(teamsConnectionRepo, appURL, logger)
	teamsHandler := handler.NewTeamsHandler(teamsConnectionRepo)
	onboardingHandler := handler.NewOnboardingHandler(onboardingRepo, userRepo, orgRepo, mailer, appURL)

	pushSubscriptionRepo := postgres.NewPushSubscriptionRepo(db)
	pushNotifier := worker.NewPushNotifier(
		pushSubscriptionRepo,
		cfg.VAPIDPublicKey,
		cfg.VAPIDPrivateKey,
		getEnv("VAPID_SUBJECT", "mailto:support@omnir.io"),
		logger,
	)
	pushHandler := handler.NewPushHandler(pushSubscriptionRepo, cfg.VAPIDPublicKey)
	if pushNotifier.Enabled() {
		logger.Info("web push notifications enabled")
	} else {
		logger.Info("web push notifications disabled (VAPID keys not set)")
	}

	setupHandler := handler.NewSetupHandler(userRepo, orgRepo, jwtSvc, cfg.OrgMode)
	orgHandler := handler.NewOrgHandler(orgRepo, userRepo, jwtSvc, cfg.OrgMode)
	authHandler := handler.NewAuthHandler(userRepo, jwtSvc).WithAuditLog(auditLogRepo).WithTOTP(totpRepo).WithOrgs(orgRepo)
	userHandler := handler.NewUserHandler(userRepo, accessRepo)
	contactHandler := handler.NewContactHandler(contactRepo).WithCustomFields(customFieldRepo).WithModuleLayouts(moduleLayoutRepo).WithDeals(dealRepo).WithAutomationEvents(automationWorker.Events)
	accountHandler := handler.NewAccountHandler(accountRepo).WithCustomFields(customFieldRepo).WithModuleLayouts(moduleLayoutRepo)
	slaInstanceHandler := handler.NewSLAInstanceHandler(slaInstanceRepo)
	dealHandler := handler.NewDealHandler(dealRepo).WithContacts(contactRepo).WithCustomFields(customFieldRepo).WithModuleLayouts(moduleLayoutRepo).WithNotifications(notificationRepo).WithSLA(slaPolicyRepo, slaInstanceRepo).WithAutomationEvents(automationWorker.Events).WithTeamsNotifier(teamsNotifier).WithPushNotifier(pushNotifier)
	activityHandler := handler.NewActivityHandler(activityRepo).WithContacts(contactRepo)
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

	ticketHandler := handler.NewTicketHandler(ticketRepo, ticketCommentRepo, ticketAttachmentRepo, storageBackend).
		WithContacts(contactRepo).
		WithCustomFields(customFieldRepo).
		WithModuleLayouts(moduleLayoutRepo).
		WithEmailNotifier(emailNotifier, userRepo, contactRepo, logger).
		WithTeamsNotifier(teamsNotifier).
		WithPushNotifier(pushNotifier)
	portalHandler := handler.NewPortalHandler(ticketRepo, ticketCommentRepo)
	slaPolicyHandler := handler.NewSLAPolicyHandler(slaPolicyRepo)
	inboundWebhookHandler := handler.NewWebhookHandler(ticketRepo, ticketCommentRepo, contactRepo, userRepo, cfg.WebhookSecret, cfg.OrgMode, logger).
		WithDispatcher(webhookDispatcher.Dispatch)
	contactNoteHandler := handler.NewNoteHandler(noteRepo, domain.NoteEntityContact, "id")
	accountNoteHandler := handler.NewNoteHandler(noteRepo, domain.NoteEntityAccount, "id")
	dealNoteHandler := handler.NewNoteHandler(noteRepo, domain.NoteEntityDeal, "id")
	leadNoteHandler := handler.NewNoteHandler(noteRepo, domain.NoteEntityLead, "id")
	searchRepo := postgres.NewSearchRepo(db)
	searchHandler := handler.NewSearchHandler(searchRepo)
	timelineRepo := postgres.NewTimelineRepo(db)
	timelineHandler := handler.NewTimelineHandler(timelineRepo)
	reportsHandler := handler.NewReportsHandler(reportsRepo)
	exportHandler := handler.NewExportHandler(contactRepo, accountRepo, dealRepo, reportsRepo).WithAuditLog(auditLogRepo)
	leadHandler := handler.NewLeadHandler(leadRepo, contactRepo, accountRepo, dealRepo, leadConversionMappingRepo, customFieldRepo).WithModuleLayouts(moduleLayoutRepo)
	customFieldHandler := handler.NewCustomFieldHandler(customFieldRepo).WithPicklistValueReader(picklistRepo)
	moduleConfigurationHandler := handler.NewModuleConfigurationHandler(moduleLayoutRepo, moduleRelationshipDefinitionRepo, customFieldRepo, crmEntityLinkRepo, accessRepo)
	apiKeyHandler := handler.NewAPIKeyHandler(apiKeyRepo)
	emailHandler := handler.NewEmailHandler(emailRepo, activityRepo, contactRepo, dealRepo, mailer, cfg.SMTP.From)
	importHandler := handler.NewImportHandler(contactRepo, accountRepo, leadRepo)
	outboundWebhookHandler := handler.NewOutboundWebhookHandler(outboundWebhookRepo)
	savedViewRepo := postgres.NewSavedViewRepo(db)
	savedViewHandler := handler.NewSavedViewHandler(savedViewRepo)
	inboundEmailHandler := handler.NewInboundEmailHandler(emailRepo, contactRepo, activityRepo, userRepo, cfg.WebhookSecret, cfg.OrgMode)
	dealPortalLinksHandler := handler.NewDealPortalLinksHandler(portalLinkRepo, dealRepo, noteRepo, orgRepo)

	uploadsDir := getEnv("UPLOADS_DIR", "uploads")
	appBaseURL := getEnv("BASE_URL", "http://localhost:8080")
	contactAttachmentHandler := handler.NewEntityAttachmentHandler(entityAttachmentRepo, uploadsDir, appBaseURL, domain.EntityTypeContact, "id")
	accountAttachmentHandler := handler.NewEntityAttachmentHandler(entityAttachmentRepo, uploadsDir, appBaseURL, domain.EntityTypeAccount, "id")
	dealAttachmentHandler := handler.NewEntityAttachmentHandler(entityAttachmentRepo, uploadsDir, appBaseURL, domain.EntityTypeDeal, "id")
	attachmentDownloadHandler := handler.NewAttachmentDownloadHandler(entityAttachmentRepo, accessRepo)
	sequenceHandler := handler.NewSequenceHandler(sequenceRepo)
	sequenceTrackingHandler := handler.NewSequenceTrackingHandler(sequenceRepo, contactRepo, cfg.SequenceTokenSecret)
	auditLogHandler := handler.NewAuditLogHandler(auditLogRepo)
	productHandler := handler.NewProductHandler(productRepo)
	quoteHandler := handler.NewQuoteHandler(quoteRepo).WithRelations(contactRepo, dealRepo).WithMailer(mailer, cfg.SMTP.From)
	automationHandler := handler.NewAutomationHandler(automationRepo)
	calendarHandler := handler.NewCalendarHandler(calendarConnectionRepo, cfg.Calendar)
	emailInboxHandler := handler.NewEmailInboxHandler(emailConnectionRepo, emailInboxRepo, cfg.EmailInbox)
	integrationsHandler := handler.NewIntegrationsHandler(integrationCredRepo, emailConnectionRepo, cfg.IntegrationCredentialsEncKey)
	orgSettingsHandler := handler.NewOrgSettingsHandler(orgSettingsRepo, cfg.IntegrationCredentialsEncKey).WithAuditLog(auditLogRepo)
	preferenceSettingsHandler := handler.NewPreferenceSettingsHandler(userPrefRepo)
	currencySettingsHandler := handler.NewCurrencySettingsHandler(currencyRepo).WithAuditLog(auditLogRepo)
	picklistSettingsHandler := handler.NewPicklistSettingsHandler(customFieldRepo, picklistRepo)
	picklistDependencySettingsHandler := handler.NewPicklistDependencySettingsHandler(picklistRepo)
	leadConversionMappingSettingsHandler := handler.NewLeadConversionMappingSettingsHandler(leadConversionMappingRepo, customFieldRepo)
	ssoHandler := handler.NewSSOHandler(ssoConfigRepo, orgRepo, userRepo, jwtSvc, cfg.SSOEncryptionKey, cfg.SSOCallbackURL).
		WithAPICallbackURL(cfg.SSOAPICallbackURL).
		WithAuditLog(auditLogRepo)
	twoFAHandler := handler.NewTwoFAHandler(totpRepo, userRepo, jwtSvc, cfg.SSOEncryptionKey).WithAuditLog(auditLogRepo)
	enrichmentSvc := enrichmentpkg.New(enrichmentCacheRepo, cfg.ClearbitAPIKey)
	enrichmentHandler := handler.NewEnrichmentHandler(enrichmentSvc, contactRepo)
	kbHandler := handler.NewKBHandler(kbArticleRepo, kbCategoryRepo).WithOrgs(orgRepo)
	billingHandler := handler.NewBillingHandler(billingRepo, cfg.Stripe, appURL)
	emailTemplateHandler := handler.NewEmailTemplateHandler(emailTemplateRepo)
	opsFinanceHandler := handler.NewOperationsFinanceHandler(opsFinanceRepo)
	accessSettingsHandler := handler.NewAccessSettingsHandler(accessRepo, userRepo)
	sequenceWorker := worker.NewSequenceWorker(sequenceRepo, emailTemplateRepo, mailer, cfg.SequenceTokenSecret, time.Minute, logger)
	sequenceWorker.Start(workerCtx)

	dashboardHandler := handler.NewDashboardHandler(dashboardRepo, reportsRepo)
	reportSchedulerWorker := worker.NewReportSchedulerWorker(dashboardRepo, reportsRepo, mailer, time.Minute, logger)
	reportSchedulerWorker.Start(workerCtx)

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
	// Public self-service registration — must be outside /api/v1 auth group.
	r.Mount("/api/v1/auth", authHandler.RegisterRouter())
	// Public SSO login/callback — no JWT required.
	r.Mount("/auth/sso", ssoHandler.Router())
	// API-style SSO: POST /api/auth/sso/microsoft|google, GET /api/auth/sso/callback
	r.Mount("/api/auth/sso", ssoHandler.APIRouter())
	// Calendar OAuth callbacks must be public because cross-site provider redirects
	// may not include SameSite-strict session cookies.
	r.Get("/api/v1/calendar/auth/google/callback", calendarHandler.CallbackGoogle)
	r.Get("/api/v1/calendar/auth/microsoft/callback", calendarHandler.CallbackMicrosoft)
	// Rate-limit public/unauthenticated endpoints to mitigate brute-force and DoS (OMN-615).
	publicRateLimit := httprate.LimitByIP(120, time.Minute)
	r.With(publicRateLimit).Mount("/webhooks/email", inboundWebhookHandler.Router())
	r.With(publicRateLimit).Mount("/api/emails/inbound", inboundEmailHandler.Router())
	// Email inbox OAuth — requires auth (browser session cookie sent on redirect callback).
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate(jwtSvc, apiKeyRepo, userRepo))
		r.Use(middleware.OrgScope(cfg.OrgMode))
		r.Mount("/api/integrations/email", emailInboxHandler.OAuthRouter())
	})
	// Public deal portal — token IS the credential, no JWT required.
	r.With(publicRateLimit).Mount("/api/portal", dealPortalLinksHandler.PublicRouter())
	r.Mount("/api/portal/help", kbHandler.PublicRouter())
	r.Mount("/api/public", kbHandler.PublicCompatibilityRouter())
	// Public sequence tracking — HMAC-signed tokens, no JWT required.
	r.Mount("/track", sequenceTrackingHandler.TrackRouter())
	r.Mount("/unsubscribe", sequenceTrackingHandler.UnsubscribeRouter())
	// Bounce webhook — outside /api/v1 auth group, accepts webhook provider calls.
	r.Mount("/api/emails/bounce", sequenceTrackingHandler.BounceRouter())
	// Stripe webhook — outside /api/v1 auth group, verified by Stripe signature.
	r.Mount("/api/billing/webhook", billingHandler.WebhookRouter())

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(chimiddleware.Timeout(30 * time.Second))
		r.Use(middleware.Authenticate(jwtSvc, apiKeyRepo, userRepo))
		r.Use(middleware.OrgScope(cfg.OrgMode))
		r.Use(middleware.ResolveAccess(accessRepo))
		r.Use(middleware.RequireModulePermission())
		r.Use(middleware.RequireNestedParentRecordAccess(accessRepo))
		r.Use(middleware.RequireFieldWriteAccess())
		r.Mount("/contacts", contactHandler.Router())
		r.Mount("/leads", leadHandler.Router())
		r.Route("/enrich", func(r chi.Router) {
			r.Use(middleware.RequirePlan(billingRepo, domain.BillingPlanPro))
			r.Mount("/", enrichmentHandler.Router())
		})
		r.Route("/leads/{id}/notes", func(r chi.Router) {
			r.Mount("/", leadNoteHandler.Router())
		})
		r.Route("/contacts/{id}/notes", func(r chi.Router) {
			r.Mount("/", contactNoteHandler.Router())
		})
		r.Route("/contacts/{id}/emails", func(r chi.Router) {
			r.Mount("/", emailHandler.ContactEmailRouter())
		})
		r.With(middleware.RequirePlan(billingRepo, domain.BillingPlanPro)).Post("/contacts/{id}/enrich", enrichmentHandler.EnrichContact)
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
		r.Mount("/timeline", timelineHandler.Router())
		r.Mount("/reports", reportsHandler.Router())
		r.Mount("/reports/schedules", dashboardHandler.ScheduleRouter())
		r.Mount("/dashboards", dashboardHandler.Router())
		r.Mount("/export", exportHandler.Router())
		r.Mount("/custom-fields", customFieldHandler.Router())
		r.Mount("/module-layouts", moduleConfigurationHandler.RuntimeLayoutRouter())
		r.Mount("/module-relationships", moduleConfigurationHandler.RuntimeRelationshipRouter())
		r.Mount("/entity-links", moduleConfigurationHandler.EntityLinksRouter())
		r.Mount("/api-keys", apiKeyHandler.Router())
		r.Route("/emails", func(r chi.Router) {
			r.Mount("/", emailHandler.Router())
			emailInboxHandler.RegisterInboxRoutes(r)
		})
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
		r.Mount("/automations", automationHandler.Router())
		r.Route("/kb", func(r chi.Router) {
			r.Use(middleware.RequirePlan(billingRepo, domain.BillingPlanPro))
			r.Mount("/", kbHandler.Router())
		})
		r.Mount("/email-templates", emailTemplateHandler.Router())
		r.Mount("/calendar", calendarHandler.Router())
		r.Route("/integrations/teams", func(r chi.Router) {
			r.Use(middleware.RequirePlan(billingRepo, domain.BillingPlanPro))
			r.Mount("/", teamsHandler.Router())
		})
		r.Mount("/integrations/email/inbox", emailInboxHandler.InboxRouter())
		r.Mount("/integrations", integrationsHandler.Router())
		r.Mount("/settings/numbering", orgSettingsHandler.Router())
		r.Mount("/settings/preferences", preferenceSettingsHandler.Router())
		r.Mount("/settings/calendar-preferences", preferenceSettingsHandler.CalendarRouter())
		r.Mount("/settings/currencies", currencySettingsHandler.Router())
		r.Mount("/settings/picklists", picklistSettingsHandler.Router())
		r.Mount("/settings/picklist-dependencies", picklistDependencySettingsHandler.Router())
		r.Mount("/settings/lead-conversion-mapping", leadConversionMappingSettingsHandler.Router())
		r.Mount("/settings/module-layouts", moduleConfigurationHandler.LayoutSettingsRouter())
		r.Mount("/settings/module-relationships", moduleConfigurationHandler.RelationshipSettingsRouter())
		r.Mount("/settings/roles", accessSettingsHandler.RolesRouter())
		r.Mount("/settings/profiles", accessSettingsHandler.ProfilesRouter())
		r.Mount("/settings/groups", accessSettingsHandler.GroupsRouter())
		r.Mount("/settings/sharing-rules", accessSettingsHandler.SharingRulesRouter())
		r.Mount("/settings/company", orgSettingsHandler.CompanyRouter())
		r.Mount("/settings/portal", orgSettingsHandler.PortalRouter())
		r.Mount("/settings/outgoing-server", orgSettingsHandler.OutgoingServerRouter())
		r.Mount("/settings/config-editor", orgSettingsHandler.ConfigEditorRouter())
		r.Mount("/settings/menu", orgSettingsHandler.MenuConfigRouter())
		r.Mount("/billing", billingHandler.Router())
		r.Mount("/ops-finance", opsFinanceHandler.Router())
		r.Mount("/onboarding", onboardingHandler.Router())
		r.Mount("/push", pushHandler.Router())
		r.Route("/deals/{dealId}/quotes", func(r chi.Router) { r.Mount("/", quoteHandler.DealQuotesRouter()) })
		r.Mount("/auth/2fa", twoFAHandler.LoginRouter())
		r.Route("/users/me/2fa", func(r chi.Router) { r.Mount("/", twoFAHandler.Router()) })
		r.Route("/orgs/{orgId}/sso", func(r chi.Router) {
			r.Use(middleware.RequirePlan(billingRepo, domain.BillingPlanEnterprise))
			r.Mount("/", ssoHandler.OrgSSORouter())
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(chimiddleware.Timeout(5 * time.Minute))
		r.Use(middleware.Authenticate(jwtSvc, apiKeyRepo, userRepo))
		r.Use(middleware.OrgScope(cfg.OrgMode))
		r.Use(middleware.ResolveAccess(accessRepo))
		r.Use(middleware.RequireModulePermission())
		r.Use(middleware.RequireNestedParentRecordAccess(accessRepo))
		r.Use(middleware.RequireFieldWriteAccess())
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
