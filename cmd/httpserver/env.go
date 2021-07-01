package main

import "github.com/spf13/viper"

// loadEnvConfiguration loads environment variables
func loadEnvConfiguration() {
	// HTTP server
	viper.SetDefault("HTTPBindAddress", "localhost:8080")
	_ = viper.BindEnv("HTTPBindAddress", "HTTP_BIND_ADDRESS")

	viper.SetDefault("ExternalLocationAddress", "http://localhost:8080")
	_ = viper.BindEnv("ExternalLocationAddress", "EXTERNAL_LOCATION_ADDRESS")

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
	viper.SetDefault("CouchDBCaPath", "")
	_ = viper.BindEnv("CouchDBCaPath", "COUCHDB_CA_PATH")
	viper.SetDefault("CouchDBUsername", "admin")
	_ = viper.BindEnv("CouchDBUsername", "COUCHDB_USERNAME")
	viper.SetDefault("CouchDBPasswd", "admin")
	_ = viper.BindEnv("CouchDBPasswd", "COUCHDB_PASSWD")

	// User service
	viper.SetDefault("UserServiceGRPCDialTarget", "localhost:50051")
	_ = viper.BindEnv("UserServiceGRPCDialTarget", "USER_SERVICE_GRPC_DIAL_TARGET")

	// Authorization service
	viper.SetDefault("AuthServiceAddress", "localhost:8081")
	_ = viper.BindEnv("AuthServiceAddress", "AUTH_SERVICE_ADDRESS")
}
