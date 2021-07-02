package e2e

import (
	"context"
	"net/http/httptest"

	"github.com/KompiTech/go-toolkit/natswatcher"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/adding"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/listing"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/updating"
	"github.com/KompiTech/itsm-commenting-service/pkg/event"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest/validation"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository/couchdb"
	"github.com/KompiTech/itsm-commenting-service/testutils"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var testChannelID = "3fc9958f-bace-4b1f-afcf-edf6670f91a9"
var bearerToken = "Bearer token - fake"

// StartServer starts new test server and returns it. The caller should call Close when finished, to shut it down.
func StartServer() (*httptest.Server, *couchdb.DBStorage, *natswatcher.Watcher) {
	logger, cfg := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	as := new(AuthServiceStub)
	us := new(UserServiceStub)

	// DB schema validator
	v, err := couchdb.NewValidator()
	if err != nil {
		panic(err)
	}

	// Couch DB - configuration for tests
	viper.SetDefault("CouchDBHost", "localhost")
	_ = viper.BindEnv("CouchDBHost", "TEST_COUCHDB_HOST")
	viper.SetDefault("CouchDBPort", "5984")
	_ = viper.BindEnv("CouchDBPort", "TEST_COUCHDB_PORT")
	viper.SetDefault("CouchDBUsername", "test")
	_ = viper.BindEnv("CouchDBUsername", "TEST_COUCHDB_USERNAME")
	viper.SetDefault("CouchDBPasswd", "test")
	_ = viper.BindEnv("CouchDBPasswd", "TEST_COUCHDB_PASSWD")

	// NATS connection
	viper.SetDefault("NATSQueueAddress", "127.0.0.1")
	_ = viper.BindEnv("NATSQueueAddress", "TEST_NATS_QUEUE_ADDRESS")
	viper.SetDefault("NATSQueuePort", "4222")
	_ = viper.BindEnv("NATSQueuePort", "TEST_NATS_QUEUE_PORT")

	// NATS client for event service
	nc, err := natswatcher.NewWatcher(&natswatcher.Config{
		NATS: natswatcher.NatsConfig{
			Address: viper.GetString("NATSQueueAddress"),
			Port:    viper.GetString("NATSQueuePort"),
		},
		Instance: "stan-blits",
		ClientID: uuid.New().String(),
	})
	if err != nil {
		logger.Fatal("could not create NATS client", zap.Error(err))
	}

	// set log level for couchDB initialization to see log output
	origLevel := cfg.Level.Level()
	cfg.Level.SetLevel(zap.InfoLevel)

	// Couch DB
	storage := couchdb.NewStorage(context.Background(), logger, couchdb.Config{
		Host:         viper.GetString("CouchDBHost"),
		Port:         viper.GetString("CouchDBPort"),
		Username:     viper.GetString("CouchDBUsername"),
		Passwd:       viper.GetString("CouchDBPasswd"),
		Validator:    v,
		EventService: event.NewService(nc),
	})

	cfg.Level.SetLevel(origLevel) // restore orig log level

	adder := adding.NewService(storage)
	lister := listing.NewService(storage)
	updater := updating.NewService(storage)

	pv, err := validation.NewPayloadValidator()
	if err != nil {
		panic(err)
	}

	server := rest.NewServer(rest.Config{
		Logger:            logger,
		AuthService:       as,
		UserService:       us,
		AddingService:     adder,
		ListingService:    lister,
		UpdatingService:   updater,
		RepositoryService: storage,
		PayloadValidator:  pv,
	})

	s := httptest.NewUnstartedServer(server)

	// set address of the newly started listener to original handler
	server.Addr = s.Listener.Addr().String()
	server.ExternalLocationAddress = "http://" + server.Addr
	s.Start()

	return s, storage, nc
}
