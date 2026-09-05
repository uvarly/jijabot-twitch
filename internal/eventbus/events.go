package eventbus

type EventType string

const (
	EventConnected = "connected"
	EventMessage   = "message"
	EventJoin      = "join"
)

type Event struct {
	Type    EventType
	Payload any
}

type MessagePayload struct {
	User   string
	UserID string
	Text   string
}
