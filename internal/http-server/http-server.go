package httpServer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	createPc "smart-pc-pc-service/internal/http-server/handlers/pcs/create-pc"
	getCommands "smart-pc-pc-service/internal/http-server/handlers/pcs/id/commands/get-commands"
	getParameters "smart-pc-pc-service/internal/http-server/handlers/pcs/id/commands/id/get-parameters"
	deletePc "smart-pc-pc-service/internal/http-server/handlers/pcs/id/delete-pc"
	getPcLogs "smart-pc-pc-service/internal/http-server/handlers/pcs/id/logs/get-pc-logs"
	"smart-pc-pc-service/internal/http-server/middlewares/request"
	"smart-pc-pc-service/internal/http-server/middlewares/uuidmw/commands"
	"smart-pc-pc-service/internal/http-server/middlewares/uuidmw/pcs"
	"smart-pc-pc-service/internal/storage/postgres"

	"smart-pc-pc-service/internal/config"
	getPcs "smart-pc-pc-service/internal/http-server/handlers/pcs/get-pcs"
	getPc "smart-pc-pc-service/internal/http-server/handlers/pcs/id/get-pc"
	updatePc "smart-pc-pc-service/internal/http-server/handlers/pcs/id/update-pc"
	"smart-pc-pc-service/internal/http-server/middlewares/auth"
	mwLogger "smart-pc-pc-service/internal/http-server/middlewares/logger"
	"smart-pc-pc-service/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
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
	storage *postgres.Storage,
) *Server {
	r := chi.NewRouter()
	r.Use(
		middleware.RequestID,
		middleware.Recoverer,
		mwLogger.New(log),
	)

	v := validator.New()

	r.Route(fmt.Sprintf("/u/{%s}/pcs", auth.UIDURLParam), func(r chi.Router) {
		r.Use(
			auth.NewAuthMiddleware(log),
			auth.NewParseUIDMiddleware(log),
			auth.NewStrictUIDMiddleware(log),
		)

		r.Get("/", getPcs.New(log, storage.Pcs))
		r.With(request.New[createPc.Request](log, v)).Post("/", createPc.New(log, storage.Pcs))

		r.Route(fmt.Sprintf("/{%s}", pcs.PcIDURLParam), func(r chi.Router) {
			r.Use(pcs.NewParsePcIDMiddleware(log))

			r.Get("/", getPc.New(log, storage.Pcs))
			r.Patch("/", updatePc.New(log, storage.Pcs))
			r.Delete("/", deletePc.New(log, storage.Pcs))

			r.Route("/logs", func(r chi.Router) {
				r.Get("/", getPcLogs.New(log, storage.PcLogs))
			})

			r.Route("/commands", func(r chi.Router) {
				r.Get("/", getCommands.New(log, storage.PcCommands))

				r.Route(fmt.Sprintf("/{%s}", commands.PcCommandIDURLParam), func(r chi.Router) {
					r.Use(commands.NewParsePcCommandIDMiddleware(log))

					r.Get("/parameters", getParameters.New(log, storage.PcCommandParameters))
				})
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
