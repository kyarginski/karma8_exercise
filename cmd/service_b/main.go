// Service B for karma8.
//
// # Description of the REST API of the service B for working with New Super Amazon S3 competitor.
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

	_ "github.com/lib/pq"

	"karma8/internal/app"
	"karma8/internal/config"
	"karma8/internal/lib/logger/sl"
)

func main() {
	cfg := config.MustLoad("service_b")
	log := sl.SetupLogger(cfg.Env)
	log.Info(
		"starting service B server",
		slog.String("env", cfg.Env),
		slog.String("version", cfg.Version),
	)

	if err := run(log, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(2)
	}
}

func run(log *slog.Logger, cfg *config.Config) error {
	log.Debug("starting db connect ",
		"connect", cfg.DBConnect,
		"port", cfg.Port,
		"redis_db", cfg.RedisDB,
	)

	application, err := app.NewServiceB(log, cfg.DBConnect, cfg.Port, cfg.RedisDB)
	defer application.Stop()
	if err != nil {
		return err
	}

	application.Start()

	return nil
}
