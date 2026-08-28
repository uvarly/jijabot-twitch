package jijabot

import "context"

type App struct{}

func NewApp() *App {
	return &App{}
}

func (a *App) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (a *App) Shutdown(_ context.Context) error {
	return nil
}
