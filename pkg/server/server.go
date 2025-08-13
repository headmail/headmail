package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/headmail/headmail/pkg/template"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/headmail/headmail/pkg/api/admin"
	"github.com/headmail/headmail/pkg/api/public"
	"github.com/headmail/headmail/pkg/config"
	"github.com/headmail/headmail/pkg/db"
	"github.com/headmail/headmail/pkg/mailer"
	"github.com/headmail/headmail/pkg/queue"
	"github.com/headmail/headmail/pkg/repository"
	"github.com/headmail/headmail/pkg/service" // for worker package in same module
	httpSwagger "github.com/swaggo/http-swagger"

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
	listService     service.ListServiceProvider
	campaignService service.CampaignServiceProvider
	deliveryService service.DeliveryServiceProvider
	templateService service.TemplateServiceProvider
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
	// create queue from DB provider and pass into delivery service
	q := srv.db.QueueRepository()
	// create mailer implementation from config
	mailerImpl := mailer.NewSMTPMailer(cfg.SMTP)
	trackingHost := cfg.Server.Public.URL
	maxAttempts := cfg.SMTP.Send.Attempts
	srv.deliveryService = service.NewDeliveryService(srv.db, q, mailerImpl, trackingHost, maxAttempts)

	templateService := template.NewService()
	srv.listService = service.NewListService(srv.db)
	srv.campaignService = service.NewCampaignService(
		srv.db,
		srv.deliveryService,
		templateService,
	)

	srv.templateService = service.NewTemplateService(srv.db)

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
	s.adminRouter.Get("/swagger/*", httpSwagger.WrapHandler)

	listHandler := admin.NewListHandler(s.listService)
	campaignHandler := admin.NewCampaignHandler(s.campaignService)
	deliveryHandler := admin.NewDeliveryHandler(s.deliveryService)
	subscriberHandler := admin.NewSubscriberHandler(s.listService)
	templateHandler := admin.NewTemplateHandler(s.templateService)

	s.adminRouter.Route("/api", func(r chi.Router) {
		listHandler.RegisterRoutes(r)
		campaignHandler.RegisterRoutes(r)
		deliveryHandler.RegisterRoutes(r)
		subscriberHandler.RegisterRoutes(r)
		templateHandler.RegisterRoutes(r)
	})
}

func (s *Server) registerPublicRoutes() {
	// Public tracking routes (open / click)
	public.RegisterRoutes(s.publicRouter, s.db, s.cfg)
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

	// start background scheduler and worker
	q := s.db.QueueRepository()
	// start scheduler: enqueue scheduled deliveries every minute
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			processed := true
			for processed {
				processed = s.enqueueDueDeliveries() > 0
			}

			<-ticker.C
		}
	}()

	// start worker(s) - single worker for now
	worker := NewWorker(s.db, q)
	_ = worker.SetHandler("delivery", s.deliveryService.HandleDeliveryQueuedItem)
	hostname, _ := os.Hostname()
	go worker.Start(context.Background(), hostname+":"+uuid.NewString())

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

// enqueueDueDeliveries finds scheduled deliveries whose scheduled_at <= now and enqueues them.
func (s *Server) enqueueDueDeliveries() int {
	ctx := context.Background()
	now := time.Now().Unix()
	// find scheduled deliveries directly from repository by timestamp
	deliveries, err := s.db.DeliveryRepository().ListScheduledBefore(ctx, now, 1000)
	if err != nil {
		log.Printf("scheduler: failed to list scheduled deliveries: %v", err)
		return 0
	}
	for _, d := range deliveries {
		unique := "delivery:" + d.ID
		payloadMap := map[string]string{"delivery_id": d.ID}
		b, _ := json.Marshal(payloadMap)
		item := &queue.QueueItem{
			ID:        uuid.New().String(),
			Type:      "delivery",
			Payload:   b,
			UniqueKey: &unique,
			Status:    queue.StatusPending,
			CreatedAt: time.Now().Unix(),
		}
		if err := s.db.QueueRepository().Enqueue(ctx, item); err != nil {
			log.Printf("scheduler: enqueue failed for delivery %s: %v", d.ID, err)
		}
	}
	return len(deliveries)
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
