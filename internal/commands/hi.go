package commands

import (
	"context"
	"fmt"
	"math/rand/v2"
)

type HiCommand struct{}

func NewHiCommand() *HiCommand {
	return &HiCommand{}
}

func (c HiCommand) Name() string {
	return "!hi"
}

func (c HiCommand) Execute(_ context.Context, p Payload, r Responder) error {
	roll := rand.Float64()
	if roll < 0.5 {
		return r.Say("Ща уебу")
	}

	return r.Say(fmt.Sprintf("Hi, @%s!", p.User))
}
