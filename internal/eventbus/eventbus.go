package eventbus

import (
	"context"
	"sync"
)

type Handler func(ctx context.Context, e Event)

type Subscriber interface {
	Subscribe(et EventType, h Handler)
}

type Publisher interface {
	Publish(ctx context.Context, e Event)
}

type Bus interface {
	Subscriber
	Publisher
}

type EventBus struct {
	mu       sync.Mutex
	handlers map[EventType][]Handler
}

func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[EventType][]Handler),
	}
}

func (eb *EventBus) Subscribe(eventType EventType, handler Handler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

func (eb *EventBus) Publish(ctx context.Context, event Event) {
	eb.mu.Lock()
	handlers := append([]Handler{}, eb.handlers[event.Type]...)
	eb.mu.Unlock()

	for _, h := range handlers {
		go h(ctx, event)
	}
}
