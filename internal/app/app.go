package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"karma8/internal/app/handler"
	"karma8/internal/app/health"
	"karma8/internal/app/services"
	"karma8/internal/app/web"
	"karma8/internal/lib/middleware"

	"github.com/gorilla/mux"
)

type App struct {
	HTTPServer *web.HTTPServer
	service    services.IService
}

// NewServiceA создает новый экземпляр сервиса A.
func NewServiceA(
	log *slog.Logger,
	connectString string,
	port int,
	useTracing bool,
	tracingAddress string,
	serviceName string,
) (*App, error) {
	const op = "app.NewServiceA"
	ctx := context.Background()

	srv, err := services.NewServiceA(log, connectString)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	telemetryMiddleware, err := addTelemetryMiddleware(ctx, useTracing, tracingAddress, serviceName)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	router := mux.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(telemetryMiddleware)

	router.HandleFunc("/live", health.LivenessHandler(srv)).Methods("GET")
	router.HandleFunc("/ready", health.ReadinessHandler(srv)).Methods("GET")

	router.HandleFunc("/api/file/{id}", handler.GetFileItem(srv)).Methods("GET")
	router.HandleFunc("/api/file", handler.PutFileItem(srv)).Methods("PUT")
	server, err := web.New(log, port, router)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Запуск фоновой задачи по очистке кэша.
	go srv.ClearCache(3 * time.Minute) // TODO: Передавать значение из конфига.

	return &App{
		HTTPServer: server,
		service:    srv,
	}, nil
}

// NewServiceB создает новый экземпляр сервиса B.
func NewServiceB(
	log *slog.Logger,
	connectString string,
	port int,
	redisDB int,
	useTracing bool,
	tracingAddress string,
	serviceName string,
) (*App, error) {
	const op = "app.NewServiceB"
	ctx := context.Background()

	srv, err := services.NewServiceB(log, connectString, redisDB)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	telemetryMiddleware, err := addTelemetryMiddleware(ctx, useTracing, tracingAddress, serviceName)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	router := mux.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(telemetryMiddleware)

	router.HandleFunc("/live", health.LivenessHandler(srv)).Methods("GET")
	router.HandleFunc("/ready", health.ReadinessHandler(srv)).Methods("GET")

	router.HandleFunc("/api/filepart/{id}", handler.GetBucketItem(srv)).Methods("GET")
	router.HandleFunc("/api/filepart", handler.PutBucketItem(srv)).Methods("PUT")
	server, err := web.New(log, port, router)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &App{
		HTTPServer: server,
		service:    srv,
	}, nil
}

// Start запускает приложение.
func (a *App) Start() {
	a.HTTPServer.Start()
}

// Stop останавливает приложение.
func (a *App) Stop() {
	if a != nil && a.service != nil {
		err := a.service.Close()
		if err != nil {
			fmt.Println("An error occurred closing service" + err.Error())

			return
		}
	}
}

func (a *App) ClearCacheAll() error {
	return a.service.ClearCacheAll()
}

func addTelemetryMiddleware(ctx context.Context, useTracing bool, tracingAddress string, serviceName string) (mux.MiddlewareFunc, error) {
	var telemetryMiddleware mux.MiddlewareFunc
	var err error
	if useTracing {
		telemetryMiddleware, err = handler.AddTelemetryMiddleware(ctx, tracingAddress, serviceName)
		if err != nil {
			return nil, err
		}
	}

	return telemetryMiddleware, nil
}
