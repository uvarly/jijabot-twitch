package commands

import (
	"context"
	"jijabot/internal/eventbus"
	"jijabot/internal/logger"
	"strings"
	"sync"
)

const commandPrefix = "!"

type Router struct {
	mu        sync.RWMutex
	responder Responder
	commands  map[string]Command
	logger    logger.Logger
}

func NewRouter(r Responder, l logger.Logger) *Router {
	return &Router{
		responder: r,
		commands:  make(map[string]Command),
	}
}

func (r *Router) Register(c Command) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.commands[c.Name()] = c
}

func (r *Router) HandleMessage(ctx context.Context, e eventbus.Event) {
	p, ok := e.Payload.(eventbus.MessagePayload)
	if !ok {
		return
	}

	if !strings.HasPrefix(p.Text, commandPrefix) {
		return
	}

	fields := strings.Fields(p.Text)
	if len(fields) == 0 {
		return
	}

	cmdName, args := fields[0], fields[1:]

	r.mu.RLock()
	cmd, ok := r.commands[cmdName]
	r.mu.Unlock()

	if !ok {
		return
	}

	payload := Payload{
		User: p.User,
		Text: p.Text,
		Args: args,
	}

	if err := cmd.Execute(ctx, payload, r.responder); err != nil {
		r.logger.Error("failed to execute command %q: %v", cmdName, err)
	}
}
