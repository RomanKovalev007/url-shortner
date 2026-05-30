package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RomanKovalev007/url-shortner/internal/config"
	"github.com/RomanKovalev007/url-shortner/internal/migrator"
	inmemory "github.com/RomanKovalev007/url-shortner/internal/repository/in-memory"
	postgresrepo "github.com/RomanKovalev007/url-shortner/internal/repository/postgres"
	"github.com/RomanKovalev007/url-shortner/internal/service"
	httptransport "github.com/RomanKovalev007/url-shortner/internal/transport/http"
	httphandler "github.com/RomanKovalev007/url-shortner/internal/transport/http/handler"
	"github.com/RomanKovalev007/url-shortner/pkg/logger"
	"github.com/RomanKovalev007/url-shortner/pkg/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	slog.SetDefault(logger.New(cfg.LogLevel))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var repo service.UrlRepo
	switch cfg.DbFlag {
	case config.FlagPostgres:
		poolCfg := postgres.PoolConfig{
			MaxConns:        cfg.Postgres.MaxConns,
			MinConns:        cfg.Postgres.MinConns,
			MaxConnLifetime: cfg.Postgres.MaxConnLifetime,
			MaxConnIdleTime: cfg.Postgres.MaxConnIdleTime,
		}
		pool, err := postgres.New(ctx, cfg.Postgres.DSN(), poolCfg)
		if err != nil {
			log.Fatalf("connect to postgres: %v", err)
		}
		defer pool.Close()

		if err := migrator.Run(cfg.Postgres.DSN()); err != nil {
			log.Fatalf("run migrations: %v", err)
		}

		repo = postgresrepo.New(pool)
	case config.FlagInMemory:
		repo = inmemory.New()
	default:
		log.Fatalf("unknown db flag: %s", cfg.DbFlag)
	}
	
	svc := service.NewService(repo)
	h := httphandler.NewHandler(svc, cfg.BaseURl)
	router := httptransport.NewRouter(h)

	srv := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	slog.Info("starting server", "addr", cfg.HTTP.Addr)

	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
}
