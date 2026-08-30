package commands

import (
	"context"
	"fmt"
)

type PokeCommand struct{}

func NewPokeCommand() *PokeCommand {
	return &PokeCommand{}
}

func (c *PokeCommand) Name() string {
	return "!poke"
}

func (c *PokeCommand) Execute(_ context.Context, p Payload, r Responder) error {
	fmt.Println("PokeCommand executed with args:", p.Args)

	if len(p.Args) == 0 {
		fmt.Println("No arguments provided for !poke command.")
		return nil
	}

	fmt.Println("Poking user:", p.Args[0])

	return r.Say(fmt.Sprintf("Тебя ткнули, @%s!", p.Args[0]))
}
