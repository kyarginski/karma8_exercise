package web

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

type HTTPServer struct {
	log    *slog.Logger
	server *http.Server
}

// New creates new HTTP server app.
func New(log *slog.Logger, port int, router *mux.Router) (*HTTPServer, error) {
	cfgAddress := fmt.Sprintf(":%d", port)

	srv := &http.Server{
		Addr:    cfgAddress,
		Handler: router,
	}

	return &HTTPServer{
		log:    log,
		server: srv,
	}, nil
}

func (s *HTTPServer) Start() {
	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				s.log.Error("failed to start server", "error", err)
				panic(err)
			}
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	s.log.Info("started http server", "port", s.server.Addr)

	<-done
	s.log.Info("stopping http server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		s.log.Error("failed to stop http server", "error", err)
	}

	s.log.Info("http server stopped")
}
