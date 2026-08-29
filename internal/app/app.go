package app

import (
	"context"
	"fmt"
)

type TwitchBot interface {
	Connect(ctx context.Context) error
	Disconnect() error
}

type App struct {
	bot TwitchBot
}

func NewApp(bot TwitchBot) *App {
	return &App{
		bot: bot,
	}
}

func (a *App) Run(ctx context.Context) error {
	return a.bot.Connect(ctx)
}

func (a *App) Shutdown(ctx context.Context) error {
	botDisconnectErr := make(chan error, 1)
	go func() { botDisconnectErr <- a.bot.Disconnect() }()

	select {
	case err := <-botDisconnectErr:
		if err != nil {
			return fmt.Errorf("failed to disconnect bot: %w", err)
		}
		return nil
	case <-ctx.Done():
		return fmt.Errorf("failed to disconnect bot: %w", ctx.Err())
	}
}
