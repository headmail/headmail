package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/headmail/headmail/pkg/api/admin"
	"github.com/headmail/headmail/pkg/config"
	"github.com/headmail/headmail/pkg/db"
	"github.com/headmail/headmail/pkg/repository"
	"github.com/headmail/headmail/pkg/service"

	// Import providers to register them
	_ "github.com/headmail/headmail/internal/db/sqlite"
)

// Server holds the dependencies for an HTTP server.
type Server struct {
	cfg          *config.Config
	db           repository.DB
	adminRouter  *chi.Mux
	publicRouter *chi.Mux
	adminServer  *http.Server
	publicServer *http.Server

	// Services
	listService *service.ListService
	// TODO: Add other services
}

// Option defines a function that configures a Server.
type Option func(*Server)

// WithDB is an option to set the database for the server.
func WithDB(db repository.DB) Option {
	return func(s *Server) {
		s.db = db
	}
}

// New creates a new Server instance.
func New(cfg *config.Config, opts ...Option) (*Server, error) {
	srv := &Server{
		cfg:          cfg,
		adminRouter:  chi.NewRouter(),
		publicRouter: chi.NewRouter(),
	}

	// Apply options
	for _, opt := range opts {
		opt(srv)
	}

	// If DB is not provided via options, create one using the provider.
	if srv.db == nil {
		provider, err := db.GetProvider(cfg.Database.Type)
		if err != nil {
			return nil, err
		}
		dbConn, err := provider.New(cfg.Database)
		if err != nil {
			return nil, err
		}
		srv.db = dbConn
	}

	// Initialize services
	srv.listService = service.NewListService(srv.db.ListRepository())
	// TODO: Initialize other services

	// Register routes
	srv.registerMiddlewares()
	srv.registerAdminRoutes()
	srv.registerPublicRoutes()

	return srv, nil
}

func (s *Server) registerMiddlewares() {
	s.adminRouter.Use(middleware.Logger)
	s.publicRouter.Use(middleware.Logger)
}

func (s *Server) registerAdminRoutes() {
	listHandler := admin.NewListHandler(s.listService)
	s.adminRouter.Route("/api", func(r chi.Router) {
		listHandler.RegisterRoutes(r)
	})
}

func (s *Server) registerPublicRoutes() {
	// TODO: Add public routes
}

// Serve starts the admin and public API servers.
func (s *Server) Serve() {
	s.adminServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.cfg.Server.Admin.Port),
		Handler: s.adminRouter,
	}

	s.publicServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.cfg.Server.Public.Port),
		Handler: s.publicRouter,
	}

	go func() {
		log.Printf("Starting admin server on port %d", s.cfg.Server.Admin.Port)
		if err := s.adminServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not start admin server: %v", err)
		}
	}()

	go func() {
		log.Printf("Starting public server on port %d", s.cfg.Server.Public.Port)
		if err := s.publicServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not start public server: %v", err)
		}
	}()
}

// Shutdown gracefully shuts down the servers.
func (s *Server) Shutdown(ctx context.Context) error {
	var errs []error

	if err := s.adminServer.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("admin server shutdown failed: %w", err))
	}

	if err := s.publicServer.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("public server shutdown failed: %w", err))
	}

	if len(errs) > 0 {
		// In a real app, you might want to join these errors.
		return errs[0]
	}

	return nil
}
