package twitchbot

import (
	"context"
	"fmt"

	twitchirc "github.com/gempir/go-twitch-irc/v4"

	"jijabot/internal/config"
	"jijabot/internal/eventbus"
)

type MessageSender interface {
	Say(channel string, text string)
}

type TwitchBot struct {
	ctx     context.Context
	client  IRCCLient
	bus     eventbus.Publisher
	channel string
}

func NewTwitchBot(cfg config.Config, bus eventbus.Publisher) (*TwitchBot, error) {
	var (
		client = twitchirc.NewClient(cfg.Twitch.Username, cfg.Twitch.Oauth)
		bot    = &TwitchBot{
			client:  client,
			bus:     bus,
			channel: cfg.Twitch.Channel,
		}
	)

	client.OnConnect(func() {
		fmt.Printf("Connected to Twitch chat\n")
		bot.bus.Publish(bot.ctx, eventbus.Event{Type: eventbus.EventConnected})
	})

	client.OnPrivateMessage(func(message twitchirc.PrivateMessage) {
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
		fmt.Printf("User %s joined the channel\n", message.User)
		bot.bus.Publish(bot.ctx, eventbus.Event{Type: eventbus.EventJoin})
	})

	return bot, nil
}

func (tb *TwitchBot) Connect(ctx context.Context) error {
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
