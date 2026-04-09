package httpServer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"smart-pc-pc-service/internal/config"
	mwLogger "smart-pc-pc-service/internal/http-server/middlewares/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	HTTPServer *http.Server
	log        *slog.Logger
	cfg        *config.HTTPServer

	done chan struct{}
}

func New(log *slog.Logger, cfg *config.HTTPServer) *Server {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return &Server{
		HTTPServer: srv,
		log:        log,
		cfg:        cfg,
		done:       make(chan struct{}),
	}
}

func (s *Server) Done() <-chan struct{} {
	return s.done
}

func (s *Server) Run(ctx context.Context) error {
	const op = "http-server.Run"

	defer close(s.done)

	errorChan := make(chan error, 1)
	go func() {
		if err := s.start(); err != nil {
			errorChan <- err
			return
		}
	}()

	select {
	case err := <-errorChan:
		return fmt.Errorf("%s: error starting http server: %w", op, err)
	case <-ctx.Done():
		stopCtx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
		defer cancel()

		if err := s.stop(stopCtx); err != nil {
			return fmt.Errorf("%s: error stopping http server: %w", op, err)
		}
	}

	return nil
}

func (s *Server) start() error {
	const op = "http-server.Start"

	if err := s.HTTPServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("%s: failed to start server: %w", op, err)
	}

	return nil
}

func (s *Server) stop(ctx context.Context) error {
	const op = "http-server.Stop"

	if err := s.HTTPServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("%s: failed to stop server: %w", op, err)
	}

	return nil
}
