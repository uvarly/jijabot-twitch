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
	"jijabot/internal/authorization"
	"jijabot/internal/commands"
	"jijabot/internal/config"
	"jijabot/internal/database"
	"jijabot/internal/duelling"
	"jijabot/internal/eventbus"
	"jijabot/internal/gambling"
	"jijabot/internal/grant"
	"jijabot/internal/logger"
	"jijabot/internal/oauth"
	"jijabot/internal/phrasebook"
	"jijabot/internal/reward"
	"jijabot/internal/twitchbot"
	"jijabot/internal/users"
	"jijabot/internal/wallet"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log := logger.NewSlogLogger(logger.WithLevel(slog.LevelDebug), logger.WithTextFormat(), logger.WithSource())
	// log := logger.NewSlogLogger(logger.WithLevel(slog.LevelInfo), logger.WithTextFormat())
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

	streamerChecker := authorization.NewStreamerChecker(cfg.JijaBot.StreamerTwitchUserID)

	phraseBook, err := phrasebook.NewPhrasebook(cfg.Phrasebook.Path)
	if err != nil {
		log.Error("failed to load phrasebook", "error", err)
		return
	}

	if err := phraseBook.ValidateEntries(cfg.Phrasebook.RequiredEntries...); err != nil {
		log.Error("failed to validate phrasebook", "error", err)
		return
	}

	phrasePicker := phrasebook.NewPicker(phraseBook)

	balanceGetter := wallet.NewBalanceGetter(
		db,
		users.NewSQLiteRepository(db),
		wallet.NewSQLiteRepository(db),
	)

	dailyRedeemer := reward.NewDailyRedeemer(
		db,
		users.NewSQLiteRepository(db),
		reward.NewSQLiteRedeemHistoryRepository(db),
		wallet.NewSQLiteRepository(db),
		cfg.JijaBot.Commands.Daily.JijaCoinAmount,
		cfg.JijaBot.Commands.Daily.ResetHourUTC,
	)

	betPlacer := gambling.NewBetPlacer(
		db,
		users.NewSQLiteRepository(db),
		gambling.NewSQLiteBetHistoryRepository(db),
		wallet.NewSQLiteRepository(db),
		cfg.JijaBot.Commands.Bet.BaseProbability,
		cfg.JijaBot.Commands.Bet.PayoutMultiple,
		cfg.JijaBot.Commands.Bet.DailyLimit,
		cfg.JijaBot.Commands.Bet.ResetHourUTC,
	)

	granter := grant.NewGranter(
		db,
		users.NewSQLiteRepository(db),
		grant.NewSQLiteGrantHistoryRepository(db),
		wallet.NewSQLiteRepository(db),
	)

	statsProvider := duelling.NewStatsProvider(
		db,
		users.NewSQLiteRepository(db),
		duelling.NewSQLiteDuelRepository(db),
		duelling.NewSQLiteMMRRepository(db),
		duelling.NewSQLiteMMRHistoryRepository(db),
		cfg.JijaBot.Commands.Duel.MMR.ELO.DefaultRating,
	)

	var ranks = make([]duelling.Rank, 0, len(cfg.JijaBot.Ranks))

	for _, rank := range cfg.JijaBot.Ranks {
		ranks = append(ranks, duelling.Rank{Name: rank.Name, MinRating: rank.MinRating})
	}

	rankProvider := duelling.NewRankProvider(ranks)

	duellist := duelling.NewDuellist(
		db,
		users.NewSQLiteRepository(db),
		wallet.NewSQLiteRepository(db),
		duelling.NewSQLiteDuelRepository(db),
		duelling.NewSQLiteStakeHistoryRepository(db),
		duelling.NewSQLiteMMRRepository(db),
		duelling.NewSQLiteMMRHistoryRepository(db),
		cfg.JijaBot.Commands.Duel.DailyLimit,
		cfg.JijaBot.Commands.Duel.ResetHourUTC,
		cfg.JijaBot.Commands.Duel.ExpiryDuration,
		cfg.JijaBot.Commands.Duel.MMR.ELO.DefaultRating,
		cfg.JijaBot.Commands.Duel.MMR.ELO.KFactor,
	)

	router := commands.NewRouter(bot, log)
	router.Register(commands.NewHiCommand(streamerChecker, phrasePicker, log))
	router.Register(commands.NewBalanceCommand(balanceGetter, phrasePicker, log))
	router.Register(commands.NewDailyCommand(dailyRedeemer, phrasePicker, log))
	router.Register(commands.NewBetCommand(betPlacer, phrasePicker, log))
	router.Register(commands.NewGrantCommand(streamerChecker, granter, phrasePicker, log))
	router.Register(commands.NewArenaCommand(phrasePicker, log))
	router.Register(commands.NewArenaTopCommand(statsProvider, phrasePicker, log))
	router.Register(commands.NewArenaStatsCommand(statsProvider, phrasePicker, log))
	router.Register(commands.NewArenaRankCommand(rankProvider, statsProvider, phrasePicker, log))
	router.Register(commands.NewDuelCommand(duellist, phrasePicker, log))
	router.Register(commands.NewDuelAcceptCommand(duellist, phrasePicker, log))
	router.Register(commands.NewDuelDeclineCommand(duellist, phrasePicker, log))
	router.Register(commands.NewDuelCancelCommand(duellist, phrasePicker, log))

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
