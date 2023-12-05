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
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"karma8/internal/app"
	"karma8/internal/config"
	"karma8/internal/lib/logger/sl"
	"karma8/internal/trace"

	_ "github.com/lib/pq"

	"go.opentelemetry.io/otel"
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

	if cfg.UseTracing {
		tp, err := trace.InitJaegerTracer(cfg.TracingAddress, serviceName, cfg.Env, cfg.UseTracing)
		if err != nil {
			log.Error(err.Error())
			os.Exit(1)
		}
		// Register our TracerProvider as the global so any imported
		// instrumentation in the future will default to using it.
		otel.SetTracerProvider(tp)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Cleanly shutdown and flush telemetry when the application exits.
		defer func(ctx context.Context) {
			// Do not make the application hang when it is shutdown.
			ctx, cancel = context.WithTimeout(ctx, time.Second*5)
			defer cancel()
			if err := tp.Shutdown(ctx); err != nil {
				log.Error(err.Error())
			}
		}(ctx)
	}

	application, err := app.NewServiceA(log, cfg.DBConnect, cfg.Port)
	defer application.Stop()
	if err != nil {
		return err
	}

	application.Start()

	return nil
}
