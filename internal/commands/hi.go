package commands

import (
	"context"
	"fmt"
	"strings"
)

type HiCommand struct{}

func NewHiCommand() *HiCommand {
	return &HiCommand{}
}

func (c *HiCommand) Name() string {
	return "!hi"
}

func (c *HiCommand) Execute(_ context.Context, p Payload, r Responder) error {
	response := fmt.Sprintf("Привет, @%s!", p.User)

	if p.User == strings.ToLower(streamerNickname) {
		return r.Say("Приветствую, Владыка!")
	}

	return r.Say(response)
}
