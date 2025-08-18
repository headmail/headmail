// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	http_swagger "github.com/headmail/headmail/internal/http-swagger"
	"github.com/headmail/headmail/internal/mail/imap"
	"github.com/headmail/headmail/internal/mail/smtp"
	"github.com/headmail/headmail/pkg/mailer"
	"github.com/headmail/headmail/pkg/receiver"
	"github.com/headmail/headmail/pkg/template"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/headmail/headmail/pkg/api/admin"
	"github.com/headmail/headmail/pkg/api/public"
	"github.com/headmail/headmail/pkg/config"
	"github.com/headmail/headmail/pkg/db"
	"github.com/headmail/headmail/pkg/repository"
	"github.com/headmail/headmail/pkg/service"

	// Import providers to register them
	_ "github.com/headmail/headmail/internal/db/sqlite"
)

// Server holds the dependencies for an HTTP server.
type Server struct {
	cfg *config.Config
	db  repository.DB

	mailer   mailer.Mailer
	receiver receiver.Receiver

	adminRouter  *chi.Mux
	publicRouter *chi.Mux
	adminServer  *http.Server
	publicServer *http.Server

	startTime time.Time
	promReg   *prometheus.Registry

	// Services
	listService     service.ListServiceProvider
	campaignService service.CampaignServiceProvider
	deliveryService service.DeliveryServiceProvider
	templateService service.TemplateServiceProvider
	trackingService service.TrackingServiceProvider
}

// Option defines a function that configures a Server.
type Option func(*Server)

// WithDB is an option to set the database for the server.
func WithDB(db repository.DB) Option {
	return func(s *Server) {
		s.db = db
	}
}

func WithMailer(m mailer.Mailer) Option {
	return func(s *Server) {
		s.mailer = m
	}
}

func WithReceiver(r receiver.Receiver) Option {
	return func(s *Server) {
		s.receiver = r
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

	if srv.mailer == nil && len(cfg.SMTP.Host) > 0 {
		srv.mailer = smtp.NewMailer(cfg.SMTP)
	}
	if srv.receiver == nil && len(cfg.IMAP.Host) > 0 {
		srv.receiver = imap.NewReceiver(&cfg.IMAP)
	}

	// Initialize services
	// create queue from DB provider and pass into delivery service
	q := srv.db.QueueRepository()
	// create mailer implementation from config
	trackingHost := cfg.Server.Public.URL
	maxAttempts := cfg.SMTP.Send.Attempts

	templateService := template.NewService()
	srv.listService = service.NewListService(srv.db)
	srv.deliveryService = service.NewDeliveryService(srv.db, templateService, q, srv.mailer, trackingHost, maxAttempts)
	srv.campaignService = service.NewCampaignService(
		srv.db,
		srv.deliveryService,
	)

	srv.templateService = service.NewTemplateService(srv.db)

	srv.trackingService = service.NewTrackingService(srv.db)

	srv.startTime = time.Now()
	srv.promReg = NewPrometheusRegistry()

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
	s.adminRouter.Get("/swagger/*", http_swagger.WrapHandler)

	listHandler := admin.NewListHandler(s.listService)
	campaignHandler := admin.NewCampaignHandler(s.campaignService)
	deliveryHandler := admin.NewDeliveryHandler(s.deliveryService, s.templateService)
	subscriberHandler := admin.NewSubscriberHandler(s.listService)
	templateHandler := admin.NewTemplateHandler(s.templateService)

	s.adminRouter.Route("/api", func(r chi.Router) {
		// register monitoring (health + prometheus metrics) using helper functions
		RegisterMetricsHandler(r, s.promReg)
		RegisterHealthHandler(r, s.startTime)

		listHandler.RegisterRoutes(r)
		campaignHandler.RegisterRoutes(r)
		deliveryHandler.RegisterRoutes(r)
		subscriberHandler.RegisterRoutes(r)
		templateHandler.RegisterRoutes(r)
	})
}

func (s *Server) registerPublicRoutes() {
	trackingHandler := public.NewTrackingHandler(&s.cfg.Tracking, s.trackingService)

	// Public tracking routes (open / click)
	trackingHandler.RegisterRoutes(s.publicRouter)
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
		ticker := time.NewTicker(time.Second * 30)
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

	if s.receiver != nil {
		ctx := context.Background()
		events, err := s.receiver.Start(ctx)
		if err != nil {
			log.Printf("imap receiver failed to start: %v", err)
		} else {
			go func() {
				for {
					event, ok := <-events
					if !ok {
						break
					}
					log.Printf("MAIL RECEIVED: %+v", event)
					if err := s.deliveryService.HandleBouncedMail(ctx, event); err != nil {
						log.Printf("HandleBouncedMail failed: %+v", err)
					}
				}
			}()
		}
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

// enqueueDueDeliveries finds scheduled deliveries whose scheduled_at <= now and enqueues them.
func (s *Server) enqueueDueDeliveries() int {
	ctx := context.Background()
	now := time.Now().Unix()

	// Ensure campaign-level scheduled deliveries are released first (sets send time on deliveries)
	if n, err := s.campaignService.ReleaseDueDeliveries(ctx, now); err != nil {
		log.Printf("scheduler: ReleaseDueDeliveries failed: %v", err)
	} else if n > 0 {
		log.Printf("scheduler: released %d deliveries for due campaigns", n)
	}

	// find scheduled deliveries directly from repository by timestamp
	deliveries, err := s.db.DeliveryRepository().ListScheduledBefore(ctx, now, 100)
	if err != nil {
		log.Printf("scheduler: failed to list scheduled deliveries: %v", err)
		return 0
	}

	if err := repository.Transactional0(s.db, ctx, func(txCtx context.Context) error {
		for _, d := range deliveries {
			if err := s.deliveryService.EnqueueDelivery(txCtx, d); err != nil {
				log.Printf("scheduler: enqueue failed for delivery %s: %v", d.ID, err)
			}
		}
		return nil
	}); err != nil {
		log.Printf("scheduler: enqueue deliveries failed: %+v", err)
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
		// Return first error; could be aggregated if desired.
		return errs[0]
	}
	return nil
}
