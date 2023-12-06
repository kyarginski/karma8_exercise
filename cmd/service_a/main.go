// Service A for karma8.
//
// # Description of the REST API of the service A for working with New Super Amazon S3 competitor.
//
// Consumes:
// - application/json
//
// Produces:
// - application/json
//
// Schemes: http, https
// Host: localhost
// Version: 1.0.0
//
// swagger:meta
package main

import (
	"fmt"
	"log/slog"
	"os"

	"karma8/internal/app"
	"karma8/internal/config"
	"karma8/internal/lib/logger/sl"

	_ "github.com/lib/pq"
)

const (
	serviceName = "service_a"
)

func main() {
	cfg := config.MustLoad(serviceName)
	log := sl.SetupLogger(cfg.Env)
	log.Info(
		"starting server "+serviceName,
		slog.String("env", cfg.Env),
		slog.String("version", cfg.Version),
		slog.Bool("use_tracing", cfg.UseTracing),
	)

	if err := run(log, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(2)
	}
}

func run(log *slog.Logger, cfg *config.Config) error {
	log.Debug("starting db connect ", "connect", cfg.DBConnect)

	application, err := app.NewServiceA(log, cfg.DBConnect, cfg.Port, cfg.UseTracing, cfg.TracingAddress, serviceName)
	defer application.Stop()
	if err != nil {
		return err
	}

	application.Start()

	return nil
}
