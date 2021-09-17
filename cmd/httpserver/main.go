package main

import (
	"context"
	"fmt"
	"net/http"

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
		AuthService:             auth.NewService(logger),
		UserService:             userService,
		AddingService:           adder,
		ListingService:          lister,
		UpdatingService:         updater,
		RepositoryService:       s,
		PayloadValidator:        pv,
		ExternalLocationAddress: viper.GetString("ExternalLocationAddress"),
	})

	// TODO add graceful shutdown

	{
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

	logger.Info(fmt.Sprintf("starting server at %s...", server.Addr))
	logger.Fatal("server start failed", zap.Error(http.ListenAndServe(server.Addr, server)))
}
