package main

import (
	"net/http"
	"os"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/adding"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/listing"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/user"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository/couchdb"
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

	s := couchdb.NewStorage(logger, couchdb.Config{
		Host:     "localhost",
		Port:     "5984",
		Username: "admin",
		Passwd:   "admin",
	})

	userService := user.NewService()
	adder := adding.NewService(s)
	lister := listing.NewService(s)

	server := rest.NewServer(rest.Config{
		Addr:           bindAddress,
		Logger:         logger,
		UserService:    userService,
		AddingService:  adder,
		ListingService: lister,
	})

	logger.Info("starting server...")
	logger.Fatal("server start failed", zap.Error(http.ListenAndServe(server.Addr, server)))
}
