package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"jijabot/internal/app"
	"jijabot/internal/commands"
	"jijabot/internal/config"
	"jijabot/internal/eventbus"
	"jijabot/internal/logger"
	"jijabot/internal/oauth"
	"jijabot/internal/twitchbot"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log := logger.NewSlogLogger(logger.WithLevel(slog.LevelDebug), logger.WithTextFormat())
	cfg, err := config.NewConfig()
	if err != nil {
		log.Error("failed to load config", "error", err)
		return
	}

	bus := eventbus.NewEventBus()
	store := oauth.NewFileStore(cfg.Oauth.TokenFile)
	_, err = store.Load(ctx)
	if err != nil {
		log.Info("no token file found, starting bot authorization flow...")

		deviceAuthorizer := oauth.NewDeviceAuthorizer(
			cfg.Twitch.ClientID,
			[]string{"chat:read", "chat:edit"},
			http.DefaultClient,
		)

		if _, err = oauth.Bootstrap(ctx, os.Stdout, deviceAuthorizer, store); err != nil {
			log.Error("failed to bootstrap oauth token", "error", err)
			return
		}
	}

	refresher := oauth.NewRefresher(cfg.Twitch.ClientID, cfg.Twitch.ClientSecret, oauth.TwitchTokenEndpoint, http.DefaultClient)
	tokenSource := oauth.NewSource(store, refresher)
	bot, err := twitchbot.NewTwitchBot(cfg, bus, tokenSource, log.With("component", "twitchbot"))
	if err != nil {
		log.Error("failed to create twitch bot", "error", err)
		return
	}

	router := commands.NewRouter(bot, logger.NoOp())
	router.Register(commands.NewHiCommand())
	router.Register(commands.NewPokeCommand())

	bus.Subscribe(eventbus.EventMessage, router.HandleMessage)

	application := app.NewApp(bot)
	appRunErr := make(chan error, 1)
	go func() { appRunErr <- application.Run(ctx) }()

	select {
	case err := <-appRunErr:
		log.Error("failed to run app", "error", err)
		return
	case <-ctx.Done():
		log.Info("shutdown signal received, shutting down...")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := application.Shutdown(shutdownCtx); err != nil {
		log.Error("failed to shutdown app", "error", err)
	}
}
