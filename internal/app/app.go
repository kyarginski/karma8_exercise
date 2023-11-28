package app

import (
	"fmt"
	"log/slog"

	"karma8/internal/app/handler"
	"karma8/internal/app/services"
	"karma8/internal/app/web"

	"github.com/gorilla/mux"
)

type App struct {
	HTTPServer *web.HTTPServer
	service    services.IService
}

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

	return &App{
		HTTPServer: server,
		service:    srv,
	}, nil
}

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

func (a *App) Start() {
	a.HTTPServer.Start()
}

func (a *App) Stop() {
	if a != nil && a.service != nil {
		a.service.Close()
	}
}
