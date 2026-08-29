package twitchbot

import twitchirc "github.com/gempir/go-twitch-irc/v4"

type IRCCLient interface {
	Connect() error
	Disconnect() error
	Join(channels ...string)
	Say(channel string, text string)
	OnConnect(callback func())
	OnPrivateMessage(callback func(message twitchirc.PrivateMessage))
	OnUserJoinMessage(callback func(message twitchirc.UserJoinMessage))
}
