package e2e

import (
	"net/http/httptest"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/adding"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/listing"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/updating"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest/validation"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository/couchdb"
	"github.com/KompiTech/itsm-commenting-service/testutils"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var testChannelID = "3fc9958f-bace-4b1f-afcf-edf6670f91a9"
var bearerToken = "Bearer token - fake"

// StartServer starts new test server and returns it. The caller should call Close when finished, to shut it down.
func StartServer() (*httptest.Server, *couchdb.DBStorage) {
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
	viper.SetDefault("TestCouchDBHost", "localhost")
	_ = viper.BindEnv("TestCouchDBHost", "TEST_COUCHDB_HOST")
	viper.SetDefault("TestCouchDBPort", "5984")
	_ = viper.BindEnv("TestCouchDBPort", "TEST_COUCHDB_PORT")
	viper.SetDefault("TestCouchDBUsername", "test")
	_ = viper.BindEnv("TestCouchDBUsername", "TEST_COUCHDB_USERNAME")
	viper.SetDefault("TestCouchDBPasswd", "test")
	_ = viper.BindEnv("TestCouchDBPasswd", "TEST_COUCHDB_PASSWD")

	// set log level for couchDB initialization to see log output
	origLevel := cfg.Level.Level()
	cfg.Level.SetLevel(zap.InfoLevel)

	// Couch DB
	storage := couchdb.NewStorage(logger, couchdb.Config{
		Host:         viper.GetString("TestCouchDBHost"),
		Port:         viper.GetString("TestCouchDBPort"),
		Username:     viper.GetString("TestCouchDBUsername"),
		Passwd:       viper.GetString("TestCouchDBPasswd"),
		Validator:    v,
		EventService: new(EventServiceStub),
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
	s.Start()

	return s, storage
}
