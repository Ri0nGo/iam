package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"iam/internal/bootstrap"
)

func main() {
	app, err := bootstrap.NewApp()
	if err != nil {
		slog.Error("bootstrap app failed", "error", err)
		os.Exit(1)
	}

	go func() {
		if err := app.Run(); err != nil {
			app.Logger.Error("server exited with error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Shutdown(ctx); err != nil {
		app.Logger.Error("shutdown failed", "error", err)
		os.Exit(1)
	}
	app.Logger.Info("server stopped")
}
