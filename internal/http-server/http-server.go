package httpServer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	getPcLogs "smart-pc-pc-service/internal/http-server/handlers/pcs/id/logs/get-pc-logs"

	"smart-pc-pc-service/internal/config"
	getPcs "smart-pc-pc-service/internal/http-server/handlers/pcs/get-pcs"
	getPc "smart-pc-pc-service/internal/http-server/handlers/pcs/id/get-pc"
	updatePc "smart-pc-pc-service/internal/http-server/handlers/pcs/id/update-pc"
	"smart-pc-pc-service/internal/http-server/middlewares/auth"
	mwLogger "smart-pc-pc-service/internal/http-server/middlewares/logger"
	"smart-pc-pc-service/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	HTTPServer *http.Server
	log        *slog.Logger
	cfg        config.HTTPServer

	done chan struct{}
}

func New(
	log *slog.Logger,
	cfg config.HTTPServer,
	pcsGetter getPcs.PcGetter,
	pcGetter getPc.PcGetter,
	updater updatePc.PcUpdater,
	pcLogsGetter getPcLogs.PcLogsGetter,
) *Server {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(mwLogger.New(log))
	r.Use(middleware.Recoverer)

	r.Route("/u/{uid}/pcs", func(r chi.Router) {
		r.Use(auth.NewAuthMiddleware(log))
		r.Use(auth.NewUserIdVerifierMiddleware(log))

		r.Get("/", getPcs.New(log, pcsGetter))
		r.Route("/{pc_id}", func(r chi.Router) {
			r.Get("/", getPc.New(log, pcGetter))
			r.Patch("/", updatePc.New(log, updater))

			r.Route("/logs", func(r chi.Router) {
				r.Get("/", getPcLogs.New(log, pcLogsGetter))
			})
		})
	})

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      r,
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
	log := s.log.With(sl.Op(op))

	defer close(s.done)

	log.Info("starting http server", slog.String("address", s.HTTPServer.Addr))
	errorChan := make(chan error, 1)
	go func() {
		if err := s.start(); err != nil {
			log.Error("failed to start http server", sl.Err(err))
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

		log.Info("shutting down http server")
		if err := s.stop(stopCtx); err != nil {
			log.Error("failed to stop http server", sl.Err(err))
			return fmt.Errorf("%s: error stopping http server: %w", op, err)
		}
		log.Info("http server stopped")
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
