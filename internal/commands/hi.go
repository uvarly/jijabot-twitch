package commands

import (
	"context"
	"fmt"
)

type HiCommand struct{}

func (c HiCommand) Name() string {
	return "hi"
}

func (c HiCommand) Execute(_ context.Context, p Payload, r Responder) error {
	return r.Say(fmt.Sprintf("Hi, @%s!", p.User))
}
