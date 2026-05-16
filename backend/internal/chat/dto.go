package chat

import "time"

type CreateConversationRequest struct {
	TargetUsername string `json:"target_username"`
}

type CreateConversationResponse struct {
	ConversationId string `json:"conversation_id"`
}

type SendMessageRequest struct {
	ConversationId string `json:"conversation_id"`
	Content        string `json:"content"`
	SenderID       string `json:"sender_id"`
}

type GetMessageRequest struct {
	ConversationId string `json:"conversation_id"`
	UserID         string `json:"user_id"`
}

type GetMessageResponse struct {
	ID             string    `json:"id"`
	ConversationId string    `json:"conversation_id"`
	SenderID       string    `json:"sender_id"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
}
