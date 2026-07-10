package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github-star-sync/internal/config"
	gh "github-star-sync/internal/github"
	syncctrl "github-star-sync/internal/sync"
)

func main() {
	configPath := flag.String("config", "config.toml", "path to config.toml")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Error("load config failed", "err", err)
		os.Exit(1)
	}

	token := config.TokenFromEnv(cfg.TokenEnv)
	if token == "" {
		logger.Info("running without token (unauthenticated rate limit applies)", "token_env", cfg.TokenEnv)
	} else {
		logger.Info("using token from env", "token_env", cfg.TokenEnv)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	client := gh.New(token)
	runner := syncctrl.New(cfg, client, logger)
	if err := runner.Run(ctx); err != nil {
		logger.Error("run failed", "err", err)
		os.Exit(1)
	}
	fmt.Println("done")
}
