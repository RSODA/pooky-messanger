package ws

import "time"

type EventType string

const (
	EventNewMessage  EventType = "new_message"
	EventUserOnline  EventType = "user_online"
	EventUserOffline EventType = "user_offline"
)

type Event struct {
	EventType EventType      `json:"event_type"`
	Payload   MessagePayload `json:"payload"`
}

type MessagePayload struct {
	From           string    `json:"from"`
	Message        string    `json:"content"`
	ConversationID string    `json:"conversation_id"`
	CreatedAt      time.Time `json:"created_at"`
}
