package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"jijabot/internal/app"
	"jijabot/internal/commands"
	"jijabot/internal/config"
	"jijabot/internal/eventbus"
	"jijabot/internal/logger"
	"jijabot/internal/twitchbot"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := logger.NewLogger()
	cfg, err := config.NewConfig()
	if err != nil {
		logger.Error("failed to load config: %v", err)
		return
	}

	bus := eventbus.NewEventBus()

	bot, err := twitchbot.NewTwitchBot(cfg, bus)
	if err != nil {
		logger.Error("failed to create twitch bot: %v", err)
		return
	}

	router := commands.NewRouter(bot, logger)
	router.Register(commands.NewHiCommand())
	router.Register(commands.NewPokeCommand())

	bus.Subscribe(eventbus.EventMessage, router.HandleMessage)

	application := app.NewApp(bot)
	appRunErr := make(chan error, 1)
	go func() { appRunErr <- application.Run(ctx) }()

	select {
	case err := <-appRunErr:
		logger.Error("failed to run app: %v", err)
		return
	case <-ctx.Done():
		logger.Info("shutdown signal received, shutting down...")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := application.Shutdown(shutdownCtx); err != nil {
		logger.Error("failed to shutdown app: %v", err)
	}
}
