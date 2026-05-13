package config

import (
	"time"

	async "github.com/kitti12911/lib-async"
	"github.com/kitti12911/lib-monitor/profiling"
	"github.com/kitti12911/lib-monitor/tracing"
	"github.com/kitti12911/lib-util/v3/logger"
)

type Config struct {
	Service   Service          `mapstructure:"service" validate:"required"`
	Logging   logger.Config    `mapstructure:"logging"`
	Tracing   tracing.Config   `mapstructure:"tracing"`
	Profiling profiling.Config `mapstructure:"profiling"`
	NATS      async.NATSConfig `mapstructure:"nats"    validate:"required"`
	Worker    Worker           `mapstructure:"worker"  validate:"required"`
}

type Service struct {
	Name            string        `mapstructure:"name"             env:"SERVICE_NAME"      validate:"required"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" env:"SHUTDOWN_TIMEOUT"`
}

type Worker struct {
	Topic string `mapstructure:"topic" env:"WORKER_TOPIC" validate:"required"`
}
