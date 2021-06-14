package main

import (
	"fmt"
	"net/http"

	"github.com/KompiTech/go-toolkit/natswatcher"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/adding"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/listing"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/updating"
	"github.com/KompiTech/itsm-commenting-service/pkg/event"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest/auth"
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
	s := couchdb.NewStorage(logger, couchdb.Config{
		CaPath: viper.GetString("CouchDBCaPath"),
		Host:         viper.GetString("CouchDBHost"),
		Port:         viper.GetString("CouchDBPort"),
		Username:     viper.GetString("CouchDBUsername"),
		Passwd:       viper.GetString("CouchDBPasswd"),
		Validator:    v,
		EventService: event.NewService(nc),
	})

	// User service fetches user data from external service
	userService, err := rest.NewUserService()
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
		Addr:              viper.GetString("HTTPBindAddress"),
		URISchema:         "http://",
		Logger:            logger,
		AuthService:       auth.NewService(logger),
		UserService:       userService,
		AddingService:     adder,
		ListingService:    lister,
		UpdatingService:   updater,
		RepositoryService: s,
		PayloadValidator:  pv,
	})

	// TODO add graceful shutdown

	logger.Info(fmt.Sprintf("starting server at %s...", server.Addr))
	logger.Fatal("server start failed", zap.Error(http.ListenAndServe(server.Addr, server)))
}

// loadEnvConfiguration loads environment variables
func loadEnvConfiguration() {
	// HTTP server
	viper.SetDefault("HTTPBindAddress", "localhost:8080")
	_ = viper.BindEnv("HTTPBindAddress", "HTTP_BIND_ADDRESS")

	// NATS connection
	viper.SetDefault("NATSQueueAddress", "127.0.0.1")
	_ = viper.BindEnv("NATSQueueAddress", "NATS_QUEUE_ADDRESS")
	viper.SetDefault("NATSQueuePort", "4222")
	_ = viper.BindEnv("NATSQueuePort", "NATS_QUEUE_PORT")

	// NATS certificates
	viper.SetDefault("NATSQueueCaPath", "./certs/ca.pem")
	_ = viper.BindEnv("NATSQueueCaPath", "NATS_QUEUE_CA_PATH")
	viper.SetDefault("NATSQueueCertPath", "./certs/cert.pem")
	_ = viper.BindEnv("NATSQueueCertPath", "NATS_QUEUE_CERT_PATH")
	viper.SetDefault("NATSQueueKeyPath", "./certs/key.pem")
	_ = viper.BindEnv("NATSQueueKeyPath", "NATS_QUEUE_KEY_PATH")

	// Couch DB
	viper.SetDefault("CouchDBHost", "localhost")
	_ = viper.BindEnv("CouchDBHost", "COUCHDB_HOST")
	viper.SetDefault("CouchDBPort", "5984")
	_ = viper.BindEnv("CouchDBPort", "COUCHDB_PORT")
	viper.SetDefault("CouchDBCaPath", "./certs_couchdb/ca.crt")
	_ = viper.BindEnv("CouchDBCaPath", "COUCHDB_CA_PATH")
	viper.SetDefault("CouchDBUsername", "admin")
	_ = viper.BindEnv("CouchDBUsername", "COUCHDB_USERNAME")
	viper.SetDefault("CouchDBPasswd", "admin")
	_ = viper.BindEnv("CouchDBPasswd", "COUCHDB_PASSWD")

	// User service
	viper.SetDefault("UserServiceGRPCDialTarget", "localhost:50051")
	_ = viper.BindEnv("UserServiceGRPCDialTarget", "USER_SERVICE_GRPC_DIAL_TARGET")
}
