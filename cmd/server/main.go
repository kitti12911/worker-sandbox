package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	async "github.com/kitti12911/lib-async"
	"github.com/kitti12911/lib-monitor/profiling"
	"github.com/kitti12911/lib-monitor/tracing"
	libconfig "github.com/kitti12911/lib-util/v3/config"
	"github.com/kitti12911/lib-util/v3/logger"

	"worker-sandbox/internal/config"
	"worker-sandbox/internal/worker"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx := context.Background()

	cfg, err := libconfig.Load[config.Config]("config.yml")
	if err != nil {
		slog.ErrorContext(ctx, "failed to load config", "error", err)
		return 1
	}

	if cfg.Service.ShutdownTimeout == 0 {
		cfg.Service.ShutdownTimeout = 10 * time.Second
	}

	logger.NewFromConfig(cfg.Logging, cfg.Service.Name)

	profiler, err := profiling.NewFromConfig(cfg.Service.Name, cfg.Profiling)
	if err != nil {
		slog.ErrorContext(ctx, "failed to init profiling", "error", err)
		return 1
	}
	defer func() {
		if shutdownErr := profiling.Shutdown(profiler); shutdownErr != nil {
			slog.ErrorContext(ctx, "failed to stop profiling", "error", shutdownErr)
		}
	}()

	tp, err := tracing.NewFromConfig(ctx, cfg.Service.Name, cfg.Tracing)
	if err != nil {
		slog.ErrorContext(ctx, "failed to init tracing", "error", err)
		return 1
	}
	defer func() {
		if shutdownErr := tracing.Shutdown(ctx, tp); shutdownErr != nil {
			slog.ErrorContext(ctx, "failed to stop tracing", "error", shutdownErr)
		}
	}()

	bus, err := async.NewNATS(cfg.NATS, nil)
	if err != nil {
		slog.ErrorContext(ctx, "failed to connect to nats", "error", err)
		return 1
	}
	defer func() {
		if closeErr := bus.Close(); closeErr != nil {
			slog.ErrorContext(ctx, "failed to close nats bus", "error", closeErr)
		}
	}()

	workerCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	handler := worker.NewHandler()
	consumeErr := make(chan error, 1)
	go func() {
		consumeErr <- async.Consume(
			workerCtx,
			bus.Subscriber(),
			async.JSONCodec{},
			cfg.Worker.Topic,
			handler.Handle,
			async.WithErrorHandler(func(ctx context.Context, msg async.Envelope[[]byte], err error) {
				slog.ErrorContext(ctx, "failed to process worker message", "message_uuid", msg.UUID, "error", err)
			}),
		)
	}()

	slog.InfoContext(ctx, "worker started", "topic", cfg.Worker.Topic)

	consumeStopped := false
	select {
	case <-workerCtx.Done():
	case err := <-consumeErr:
		consumeStopped = true
		if err != nil {
			slog.ErrorContext(ctx, "worker stopped with error", "error", err)
			return 1
		}
	}

	slog.InfoContext(ctx, "shutting down worker")
	stop()

	if err := waitForWorkerStop(ctx, cfg.Service.ShutdownTimeout, consumeStopped, consumeErr); err != nil {
		slog.ErrorContext(ctx, "worker stopped with error", "error", err)
		return 1
	}

	slog.InfoContext(ctx, "worker stopped")

	return 0
}

func waitForWorkerStop(
	ctx context.Context,
	timeout time.Duration,
	alreadyStopped bool,
	consumeErr <-chan error,
) error {
	if alreadyStopped {
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	select {
	case err := <-consumeErr:
		return err
	case <-shutdownCtx.Done():
		slog.WarnContext(ctx, "worker shutdown timed out")
		return nil
	}
}
