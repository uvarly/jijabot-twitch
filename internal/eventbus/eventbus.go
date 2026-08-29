package eventbus

type Publisher interface{}

type EventBus struct {
}

func NewEventBus() *EventBus {
	return &EventBus{}
}
