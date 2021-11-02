package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KompiTech/go-toolkit/natswatcher"
	"github.com/KompiTech/go-toolkit/tracing"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/adding"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/listing"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/updating"
	"github.com/KompiTech/itsm-commenting-service/pkg/event"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest/auth"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest/usersvc"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest/validation"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository/couchdb"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer func(logger *zap.Logger) {
		_ = logger.Sync()
	}(logger)

	logger.Info("app starting...")

	loadEnvConfiguration()

	// DB schema validator
	v, err := couchdb.NewValidator()
	if err != nil {
		logger.Fatal("could not create couchDB validation service", zap.Error(err))
	}

	// NATS client for event service
	nc, err := natswatcher.NewWatcher(&natswatcher.Config{
		NATS: natswatcher.NatsConfig{
			Address: viper.GetString("NATSQueueAddress"),
			Port:    viper.GetString("NATSQueuePort"),
			TLS: &natswatcher.TLS{
				CAPath:   viper.GetString("NATSQueueCaPath"),
				CertPath: viper.GetString("NATSQueueCertPath"),
				KeyPath:  viper.GetString("NATSQueueKeyPath"),
			},
		},
		Instance: "stan-blits",
		ClientID: uuid.New().String(),
	})
	if err != nil {
		logger.Fatal("could not create NATS client", zap.Error(err))
	}

	// Couch DB
	s := couchdb.NewStorage(context.Background(), logger, couchdb.Config{
		CaPath:       viper.GetString("CouchDBCaPath"),
		Host:         viper.GetString("CouchDBHost"),
		Port:         viper.GetString("CouchDBPort"),
		Username:     viper.GetString("CouchDBUsername"),
		Passwd:       viper.GetString("CouchDBPasswd"),
		Validator:    v,
		EventService: event.NewService(nc),
	})

	// User service fetches user data from external service
	userService, err := usersvc.NewService()
	if err != nil {
		logger.Fatal("could not create user service", zap.Error(err))
	}

	// Auth service provides ACL functionality
	authService := auth.NewService(logger)

	adder := adding.NewService(s)
	lister := listing.NewService(s)
	updater := updating.NewService(s)

	// Request payload validator
	pv, err := validation.NewPayloadValidator()
	if err != nil {
		logger.Fatal("could not create payload validation service", zap.Error(err))
	}

	// HTTP server
	server := rest.NewServer(rest.Config{
		Addr:                    viper.GetString("HTTPBindAddress"),
		URISchema:               "http://",
		Logger:                  logger,
		AuthService:             authService,
		UserService:             userService,
		AddingService:           adder,
		ListingService:          lister,
		UpdatingService:         updater,
		RepositoryService:       s,
		PayloadValidator:        pv,
		ExternalLocationAddress: viper.GetString("ExternalLocationAddress"),
	})

	{ // Setup tracing
		logger.Info("setting up tracing")
		tracer, err := tracing.NewZipkinTracer(viper.GetString("TracingCollectorEndpoint"),
			"blits-itsm-commenting-service",
			viper.GetString("HTTPBindPort"),
			"1", 3)
		if err != nil {
			logger.Fatal(err.Error())
		}

		openTracer := zipkinot.Wrap(tracer)
		opentracing.SetGlobalTracer(openTracer)
	}

	srv := &http.Server{
		Addr:    server.Addr,
		Handler: server,
	}

	// Graceful shutdown
	idleConnsClosed := make(chan struct{})
	go func() {
		// Trap sigterm or interrupt and gracefully shutdown the server
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		sig := <-sigint
		logger.Info(fmt.Sprintf("got signal: %s", sig))
		// We received a signal, shut down.

		// Gracefully shutdown the server, waiting max 'timeout' seconds for current operations to complete
		timeout := viper.GetInt("HTTPShutdownTimeoutInSeconds")
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
		defer cancel()

		logger.Info("shutting down HTTP server...")
		if err := srv.Shutdown(ctx); err != nil {
			// Error from closing listeners, or context timeout:
			logger.Error("HTTP server Shutdown", zap.Error(err))
		}
		logger.Info("HTTP server shutdown finished successfully")

		// Close connection to external user service
		logger.Info("closing UserService client")
		if err := userService.Close(); err != nil {
			logger.Error("error closing UserService client", zap.Error(err))
		}

		// Close connection to external auth service
		logger.Info("closing AuthService client")
		if err := authService.Close(); err != nil {
			logger.Error("error closing AuthService client", zap.Error(err))
		}

		// Close database client
		logger.Info("closing database client")
		if err := s.Client().Close(context.Background()); err != nil {
			logger.Error("error closing database client", zap.Error(err))
		}

		// Unsubscribe NATS client from all subscriptions and close the connection to the cluster
		logger.Info("closing NATS client")
		if err := nc.Close(); err != nil {
			logger.Error("error closing NATS client", zap.Error(err))
		}

		close(idleConnsClosed)
	}()

	// Start the server
	logger.Info(fmt.Sprintf("starting HTTP server at %s", server.Addr))
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		logger.Fatal("HTTP server ListenAndServe", zap.Error(err))
	}

	// Block until a signal is received and graceful shutdown completed.
	<-idleConnsClosed

	logger.Info("exiting")
	_ = logger.Sync()
}
