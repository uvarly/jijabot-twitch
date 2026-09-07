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
	"jijabot/internal/bet"
	"jijabot/internal/commands"
	"jijabot/internal/config"
	"jijabot/internal/database"
	"jijabot/internal/eventbus"
	"jijabot/internal/gambling"
	"jijabot/internal/logger"
	"jijabot/internal/oauth"
	"jijabot/internal/redeem"
	"jijabot/internal/reward"
	"jijabot/internal/twitchbot"
	"jijabot/internal/users"
	"jijabot/internal/wallet"
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

	db, err := database.Open(ctx, cfg.Database.Path)
	if err != nil {
		log.Error("failed to open database", "error", err)
		return
	}
	defer db.Close()

	bus := eventbus.NewEventBus()
	store := oauth.NewFileStore(cfg.Oauth.TokenFilePath)
	_, err = store.Load(ctx)
	if err != nil {
		log.Info("no token file found, starting bot authorization flow...")

		deviceAuthorizer := oauth.NewDeviceAuthorizer(
			cfg.TwitchBot.ClientID,
			[]string{"chat:read", "chat:edit"},
			http.DefaultClient,
		)

		if _, err = oauth.Bootstrap(ctx, os.Stdout, deviceAuthorizer, store); err != nil {
			log.Error("failed to bootstrap oauth token", "error", err)
			return
		}
	}

	refresher := oauth.NewRefresher(cfg.TwitchBot.ClientID, cfg.TwitchBot.ClientSecret, oauth.TwitchTokenEndpoint, http.DefaultClient)
	tokenSource := oauth.NewSource(store, refresher)
	bot, err := twitchbot.NewTwitchBot(cfg, bus, tokenSource, log.With("component", "twitchbot"))
	if err != nil {
		log.Error("failed to create twitch bot", "error", err)
		return
	}

	claimer := reward.NewDailyClaimer(
		db,
		users.NewSQLiteRepository(db),
		redeem.NewSQLiteRepository(db),
		wallet.NewSQLiteRepository(db),
		cfg.JijaBot.Daily.JijaCoinAmount,
		cfg.JijaBot.Daily.ResetHourUTC,
	)

	betPlacer := gambling.NewBetPlacer(
		db,
		users.NewSQLiteRepository(db),
		bet.NewSQLiteRepository(db),
		wallet.NewSQLiteRepository(db),
		cfg.JijaBot.Bet.BaseProbability,
		cfg.JijaBot.Bet.PayoutMultiple,
		cfg.JijaBot.Bet.DailyLimit,
		cfg.JijaBot.Bet.ResetHourUTC,
	)

	router := commands.NewRouter(bot, log)
	router.Register(commands.NewHiCommand(log))
	router.Register(commands.NewDailyCommand(claimer, log))
	router.Register(commands.NewBetCommand(betPlacer, log))

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
