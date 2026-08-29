package commands

import "context"

type Payload struct {
	User string
	Text string
	Args []string
}

type Command interface {
	Name() string
	Execute(ctx context.Context, p Payload, r Responder) error
}

type Responder interface {
	Say(message string) error
}
