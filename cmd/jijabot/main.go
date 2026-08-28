package main

import (
	"context"
	"log"
	"os/signal"
	"time"

	"jijabot/internal/jijabot"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background())
	defer stop()

	app := jijabot.NewApp()
	if err := app.Run(ctx); err != nil {
		log.Fatalf("failed to run app: %v", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("failed to shutdown app: %v", err)
	}
}
