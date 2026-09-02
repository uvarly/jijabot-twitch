package twitchbot

import (
	"context"
	"fmt"

	twitchirc "github.com/gempir/go-twitch-irc/v4"

	"jijabot/internal/config"
	"jijabot/internal/eventbus"
	"jijabot/internal/logger"
	"jijabot/internal/oauth"
)

type MessageSender interface {
	Say(channel string, text string)
}

type TwitchBot struct {
	ctx           context.Context
	client        IRCCLient
	bus           eventbus.Publisher
	tokenProvider oauth.TokenProvider
	log           logger.Logger
	channel       string
}

func NewTwitchBot(cfg config.Config, bus eventbus.Publisher, tokenProvider oauth.TokenProvider, log logger.Logger) (*TwitchBot, error) {
	var (
		client = twitchirc.NewClient(cfg.Twitch.Username, "")
		bot    = &TwitchBot{
			client:        client,
			bus:           bus,
			tokenProvider: tokenProvider,
			log:           log,
			channel:       cfg.Twitch.Channel,
		}
	)

	client.OnConnect(func() {
		bot.log.Info("connected to Twitch chat", "channel", bot.channel)
		bot.bus.Publish(bot.ctx, eventbus.Event{Type: eventbus.EventConnected})
	})

	client.OnPrivateMessage(func(message twitchirc.PrivateMessage) {
		bot.log.Debug("received message", "user", message.User.Name, "message", message.Message)
		fmt.Printf("Received private message from %s: %s\n", message.User.Name, message.Message)
		bot.bus.Publish(bot.ctx, eventbus.Event{
			Type: eventbus.EventMessage,
			Payload: eventbus.MessagePayload{
				User: message.User.Name,
				Text: message.Message,
			},
		})
	})

	client.OnUserJoinMessage(func(message twitchirc.UserJoinMessage) {
		bot.log.Debug("user joined channel", "user", message.User)
		bot.bus.Publish(bot.ctx, eventbus.Event{Type: eventbus.EventJoin})
	})

	return bot, nil
}

func (tb *TwitchBot) Connect(ctx context.Context) error {
	token, err := tb.tokenProvider.Token(ctx)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	tb.client.SetIRCToken("oauth:" + token.AccessToken)
	tb.client.Join(tb.channel)

	return tb.client.Connect()
}

func (tb *TwitchBot) Disconnect() error {
	return tb.client.Disconnect()
}

func (tb *TwitchBot) Say(message string) error {
	tb.client.Say(tb.channel, message)
	return nil
}
