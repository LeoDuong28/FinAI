package main

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/nghiaduong/finai/internal/config"
	"github.com/nghiaduong/finai/internal/router"
)

func main() {
	// Structured logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if os.Getenv("ENVIRONMENT") != "production" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load and validate config (fail-fast)
	cfg, err := config.Load(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	log.Info().
		Str("env", cfg.Server.Environment).
		Str("addr", cfg.Server.Addr()).
		Bool("demo_mode", cfg.Features.DemoMode).
		Msg("starting FinAI server")

	// Connect to database with retry
	pool, err := connectDB(ctx, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer pool.Close()

	// Build router with all dependencies
	r, routerCleanup := router.New(cfg, pool)
	defer routerCleanup()

	// HTTP server
	srv := &http.Server{
		Addr:         cfg.Server.Addr(),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.ReadTimeout + 10*time.Second, // extra time for response writing
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info().Str("addr", cfg.Server.Addr()).Msg("server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	// Wait for shutdown signal
	sig := <-shutdown
	log.Info().Str("signal", sig.String()).Msg("shutting down")

	// Give active connections 10 seconds to finish
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("forced shutdown")
	}

	log.Info().Msg("server stopped")
}

// connectDB retries database connection with exponential backoff.
// Handles Neon.tech cold starts (1-3s) and Docker Compose race conditions.
func connectDB(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	maxAttempts := 5

	poolCfg, err := pgxpool.ParseConfig(cfg.Database.URL)
	if err != nil {
		return nil, fmt.Errorf("parsing database URL: %w", err)
	}

	poolCfg.MaxConns = cfg.Database.MaxConns
	poolCfg.MinConns = cfg.Database.MinConns
	poolCfg.MaxConnLifetime = cfg.Database.MaxConnLifetime
	poolCfg.MaxConnIdleTime = cfg.Database.MaxConnIdleTime
	poolCfg.HealthCheckPeriod = 30 * time.Second

	var lastErr error
	for i := 0; i < maxAttempts; i++ {
		pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
		if err != nil {
			lastErr = err
		} else {
			if pingErr := pool.Ping(ctx); pingErr != nil {
				lastErr = pingErr
				pool.Close()
			} else {
				log.Info().Msg("connected to database")
				return pool, nil
			}
		}

		log.Warn().
			Int("attempt", i+1).
			Err(lastErr).
			Msg("database not ready")

		// Don't sleep after the last failed attempt
		if i < maxAttempts-1 {
			backoff := time.Duration(math.Pow(2, float64(i))) * time.Second
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled while connecting to database: %w", ctx.Err())
			}
		}
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxAttempts, lastErr)
}
