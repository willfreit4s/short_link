package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/willfreit4s/short_link/configs"
	"github.com/willfreit4s/short_link/internal/bootstrap"
	"github.com/willfreit4s/short_link/pkg/database"
)

func main() {
	cfg := configs.LoadConfig()

	log, err := initLogger(cfg)
	if err != nil {
		panic(err)
	}

	db, err := database.InitDatabase(cfg, log)
	if err != nil {
		log.Error("database init", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	router := bootstrap.NewRouter(log, db)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ServerPort),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("listen", "err", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server Shutdown", "err", err)
	}
	log.Info("Server exiting")
}

func initLogger(cfg *configs.Config) (*slog.Logger, error) {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	log.Info(
		"Starting application",
		"service_name", cfg.ServiceName,
		"environment", cfg.Environment,
		"server_port", cfg.ServerPort,
	)

	return log, nil
}
