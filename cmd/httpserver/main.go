package main

import (
	"net/http"

	"github.com/KompiTech/go-toolkit/natswatcher"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/adding"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/listing"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/updating"
	"github.com/KompiTech/itsm-commenting-service/pkg/event"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest/validation"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository/couchdb"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer func(logger *zap.Logger) {
		_ = logger.Sync()
	}(logger)

	// HTTP server
	viper.SetDefault("HTTPBindAddress", "localhost:8080")
	_ = viper.BindEnv("HTTPBindAddress", "HTTP_BIND_ADDRESS")

	// NATS
	viper.SetDefault("NATSQueueAddress", "127.0.0.1")
	_ = viper.BindEnv("NATSQueueAddress", "NATS_QUEUE_ADDRESS")
	viper.SetDefault("NATSQueuePort", "4222")
	_ = viper.BindEnv("NATSQueuePort", "NATS_QUEUE_PORT")

	// Couch DB
	viper.SetDefault("CouchDBHost", "localhost")
	_ = viper.BindEnv("CouchDBHost", "COUCHDB_HOST")
	viper.SetDefault("CouchDBPort", "5984")
	_ = viper.BindEnv("CouchDBPort", "COUCHDB_PORT")
	viper.SetDefault("CouchDBUsername", "admin")
	_ = viper.BindEnv("CouchDBUsername", "COUCHDB_USERNAME")
	viper.SetDefault("CouchDBPasswd", "admin")
	_ = viper.BindEnv("CouchDBPasswd", "COUCHDB_PASSWD")

	v, err := couchdb.NewValidator()
	if err != nil {
		logger.Fatal("could not create couchDB validation service", zap.Error(err))
	}

	nc, err := natswatcher.NewWatcher(&natswatcher.Config{
		NATS: natswatcher.NatsConfig{
			Address: viper.GetString("NATSQueueAddress"),
			Port:    viper.GetString("NATSQueuePort"),
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
		Host:         viper.GetString("CouchDBHost"),
		Port:         viper.GetString("CouchDBPort"),
		Username:     viper.GetString("CouchDBUsername"),
		Passwd:       viper.GetString("CouchDBPasswd"),
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
		Addr:              viper.GetString("HTTPBindAddress"),
		URISchema:         "http://",
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
