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
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/config"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/handler"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository/postgres"
)

var Version = "dev"

func main() {
	// Load .env (ignored in production/CI where vars are injected)
	_ = godotenv.Load()

	cfg := config.Load()
	logger := setupLogger(cfg.Env)

	// Connect to PostgreSQL
	db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		logger.Error("database ping failed", "err", err)
		os.Exit(1)
	}
	logger.Info("database connected")

	// JWT service
	jwtSvc := auth.NewJWTService(cfg.JWTSecret)

	// Repositories
	contactRepo := postgres.NewContactRepo(db)
	accountRepo := postgres.NewAccountRepo(db)
	dealRepo := postgres.NewDealRepo(db)
	activityRepo := postgres.NewActivityRepo(db)
	noteRepo := postgres.NewNoteRepo(db)
	userRepo := postgres.NewUserRepo(db)

	// Handlers
	setupHandler := handler.NewSetupHandler(userRepo, jwtSvc)
	authHandler := handler.NewAuthHandler(userRepo, jwtSvc)
	contactHandler := handler.NewContactHandler(contactRepo)
	accountHandler := handler.NewAccountHandler(accountRepo)
	dealHandler := handler.NewDealHandler(dealRepo)
	activityHandler := handler.NewActivityHandler(activityRepo)
	contactNoteHandler := handler.NewNoteHandler(noteRepo, domain.NoteEntityContact, "id")
	accountNoteHandler := handler.NewNoteHandler(noteRepo, domain.NoteEntityAccount, "id")
	dealNoteHandler := handler.NewNoteHandler(noteRepo, domain.NoteEntityDeal, "id")

	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Logger(logger))
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(30 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check (unauthenticated)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","version":"` + Version + `"}`))
	})

	// Setup endpoints (unauthenticated — fresh install only)
	r.Mount("/api/setup", setupHandler.Router())

	// Auth endpoints (unauthenticated)
	r.Mount("/api/auth", authHandler.Router())

	// API v1 (all routes require authentication + org scoping)
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.Authenticate(jwtSvc))
		r.Mount("/contacts", contactHandler.Router())
		r.Route("/contacts/{id}/notes", func(r chi.Router) {
			r.Mount("/", contactNoteHandler.Router())
		})
		r.Mount("/accounts", accountHandler.Router())
		r.Route("/accounts/{id}/notes", func(r chi.Router) {
			r.Mount("/", accountNoteHandler.Router())
		})
		r.Mount("/deals", dealHandler.Router())
		r.Route("/deals/{id}/notes", func(r chi.Router) {
			r.Mount("/", dealNoteHandler.Router())
		})
		r.Mount("/activities", activityHandler.Router())
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
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
	var handler slog.Handler
	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(handler)
}
