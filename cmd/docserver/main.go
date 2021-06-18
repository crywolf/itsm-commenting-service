package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	l, _ := zap.NewProduction()
	defer func(l *zap.Logger) {
		_ = l.Sync()
	}(l)

	logger := l.Sugar()

	r := httprouter.New()

	// API documentation
	opts := middleware.RedocOpts{Path: "/", SpecURL: "/swagger.yaml", Title: "Commenting service API documentation"}
	docsHandler := middleware.Redoc(opts, nil)
	// handlers for API documentation
	r.Handler(http.MethodGet, "/", docsHandler)
	r.Handler(http.MethodGet, "/swagger.yaml", http.FileServer(http.Dir("./pkg/http/rest")))

	viper.SetDefault("HTTPBindPort", "3001")
	_ = viper.BindEnv("HTTPBindPort", "HTTP_DOCS_PORT")

	flag.String("port", "3001", "http server port")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	_ = viper.BindPFlags(pflag.CommandLine)

	port := viper.GetString("port")

	addr := fmt.Sprintf("localhost:%s", port)
	logger.Info(fmt.Sprintf("starting API documentation server at %s...", addr))
	logger.Fatal(http.ListenAndServe(addr, r))
}
