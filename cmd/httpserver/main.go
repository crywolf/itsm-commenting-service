package main

import (
	"net/http"
	"os"

	"github.com/KompiTech/go-toolkit/natswatcher"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/adding"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/listing"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/updating"
	"github.com/KompiTech/itsm-commenting-service/pkg/event"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest/validation"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository/couchdb"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer func(logger *zap.Logger) {
		_ = logger.Sync()
	}(logger)

	bindAddress := os.Getenv("BIND_ADDRESS")
	if bindAddress == "" {
		bindAddress = "localhost:8080"
	}

	v, err := couchdb.NewValidator()
	if err != nil {
		logger.Fatal("could not create couchDB validation service", zap.Error(err))
	}

	nc, err := natswatcher.NewWatcher(&natswatcher.Config{
		NATS: natswatcher.NatsConfig{
			Address: "127.0.0.1",
			Port:    "4222",
			TLS: &natswatcher.TLS{
				CAPath:   "./certs/ca.pem",
				CertPath: "./certs/cert.pem",
				KeyPath:  "./certs/key.pem",
			},
		},
		Instance: "stan-blits",
		ClientID: uuid.New().String(),
	})
	if err != nil {
		logger.Fatal("could not create NATS client", zap.Error(err))
	}

	s := couchdb.NewStorage(logger, couchdb.Config{
		Host:         "localhost",
		Port:         "5984",
		Username:     "admin",
		Passwd:       "admin",
		Validator:    v,
		EventService: event.NewService(nc),
	})

	userService := rest.NewUserService()
	adder := adding.NewService(s)
	lister := listing.NewService(s)
	updater := updating.NewService(s)

	pv, err := validation.NewPayloadValidator()
	if err != nil {
		logger.Fatal("could not create payload validation service", zap.Error(err))
	}

	server := rest.NewServer(rest.Config{
		Addr:              bindAddress,
		Logger:            logger,
		UserService:       userService,
		AddingService:     adder,
		ListingService:    lister,
		UpdatingService:   updater,
		RepositoryService: s,
		PayloadValidator:  pv,
	})

	logger.Info("starting server...")
	logger.Fatal("server start failed", zap.Error(http.ListenAndServe(server.Addr, server)))
}
