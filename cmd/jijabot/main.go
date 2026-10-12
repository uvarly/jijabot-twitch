package main

import (
	"context"
	"fmt"
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
	log := logger.NewSlogLogger(logger.WithTextFormat(), logger.WithSource())

	if err := run(log); err != nil {
		log.Error("exited with error", "error", err)
		wait()
		os.Exit(1)
	}
}

func run(log logger.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.NewConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	log = logger.NewSlogLogger(logger.WithLevel(logger.SlogLevel(cfg.Logging.Level)), logger.WithTextFormat(), logger.WithSource())

	db, err := database.Open(ctx, cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
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
			return fmt.Errorf("failed to bootstrap oauth token: %w", err)
		}
	}

	refresher := oauth.NewRefresher(cfg.TwitchBot.ClientID, cfg.TwitchBot.ClientSecret, oauth.TwitchTokenEndpoint, http.DefaultClient)
	tokenSource := oauth.NewSource(store, refresher)
	bot, err := twitchbot.NewTwitchBot(cfg, bus, tokenSource, log.With("component", "twitchbot"))
	if err != nil {
		return fmt.Errorf("failed to create twitch bot: %w", err)
	}

	streamerChecker := authorization.NewStreamerChecker(cfg.JijaBot.StreamerTwitchUserID)

	phraseBook, err := phrasebook.NewPhrasebook(cfg.Phrasebook.Path)
	if err != nil {
		return fmt.Errorf("failed to load phrasebook: %w", err)
	}

	if err := phraseBook.ValidateEntries(cfg.Phrasebook.RequiredEntries...); err != nil {
		return fmt.Errorf("failed to validate phrasebook: %w", err)
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
		cfg.JijaBot.Commands.ArenaTop.Limit,
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
		return fmt.Errorf("failed to run app: %w", err)
	case <-ctx.Done():
		log.Info("shutdown signal received, shutting down...")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := application.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown app: %w", err)
	}

	return nil
}
