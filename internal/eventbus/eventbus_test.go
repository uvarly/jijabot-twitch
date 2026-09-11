package eventbus_test

import (
	"context"
	"jijabot/internal/eventbus"
	"testing"
	"time"
)

const waitTimeout = 100 * time.Millisecond

var (
	_ eventbus.Publisher  = (*eventbus.EventBus)(nil)
	_ eventbus.Subscriber = (*eventbus.EventBus)(nil)
	_ eventbus.Bus        = (*eventbus.EventBus)(nil)
)

func TestEventBus_Publish_NoSubscribers(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewEventBus()

	bus.Publish(context.Background(), eventbus.Event{})
}

func TestEventBus_Publish_DeliversToSubscriber(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewEventBus()
	received := make(chan eventbus.Event, 1)

	bus.Subscribe(eventbus.EventMessage, func(_ context.Context, e eventbus.Event) {
		received <- e
	})

	want := eventbus.Event{
		Type: eventbus.EventMessage,
		Payload: eventbus.MessagePayload{
			User:   "foo",
			UserID: "bar",
			Text:   "baz",
		},
	}

	bus.Publish(context.Background(), want)

	select {
	case got := <-received:
		if got.Type != want.Type {
			t.Errorf("Event type mismatch: got %v, want %v", got.Type, want.Type)
		}

		gotPayload, ok := got.Payload.(eventbus.MessagePayload)
		if !ok {
			t.Errorf("Event payload type mismatch: got %T, want %T", got.Payload, want.Payload)
		}

		wantPayload, _ := want.Payload.(eventbus.MessagePayload)

		if gotPayload != wantPayload {
			t.Errorf("Event payload mismatch: got %v, want %v", gotPayload, wantPayload)
		}
	case <-time.After(waitTimeout):
		t.Fatal("timed out waiting for event to be received")
	}
}

func TestEventBus_Publish_MultipleSubscribersAllInvoked(t *testing.T) {
	t.Parallel()

	// bus := eventbus.NewEventBus()
}

func TestEventBus_Publish_OnlyDeliversToMatchingSubscribers(t *testing.T) {
	t.Parallel()

	// bus := eventbus.NewEventBus()
}

func TestEventBus_Publish_PropagatesContext(t *testing.T) {
	t.Parallel()

	// bus := eventbus.NewEventBus()
}

func TestEventBus_Subscribe_DoesNotReceivePastEvents(t *testing.T) {
	t.Parallel()

	// bus := eventbus.NewEventBus()
}

func TestEventBus_ConcurrentSubscribeAndPublish(t *testing.T) {
	t.Parallel()

	// bus := eventbus.NewEventBus()
}
