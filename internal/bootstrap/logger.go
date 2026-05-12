package bootstrap

import (
	"log/slog"
	"os"
)

func NewLogger(cfg *Config) *slog.Logger {
	level := slog.LevelInfo
	if cfg.App.Env != "prod" {
		level = slog.LevelDebug
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return slog.New(handler)
}
