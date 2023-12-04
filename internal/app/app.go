package app

import (
	"fmt"
	"log/slog"
	"time"

	"karma8/internal/app/handler"
	"karma8/internal/app/services"
	"karma8/internal/app/web"

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
) (*App, error) {
	const op = "app.NewServiceA"

	srv, err := services.NewServiceA(log, connectString)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	router := mux.NewRouter()

	router.HandleFunc("/api/file/{id}", handler.GetFileItem(srv)).Methods("GET")
	router.HandleFunc("/api/file", handler.PutFileItem(srv)).Methods("PUT")
	server := web.New(log, port, router)

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
) (*App, error) {
	const op = "app.NewServiceB"

	srv, err := services.NewServiceB(log, connectString, redisDB)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	router := mux.NewRouter()

	router.HandleFunc("/api/filepart/{id}", handler.GetBucketItem(srv)).Methods("GET")
	router.HandleFunc("/api/filepart", handler.PutBucketItem(srv)).Methods("PUT")
	server := web.New(log, port, router)

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
		a.service.Close()
	}
}

func (a *App) ClearCacheAll() error {
	return a.service.ClearCacheAll()
}
