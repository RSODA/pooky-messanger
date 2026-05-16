package chat

import (
	"context"
	"pooky-messanger/internal/ws"
	"time"
)

type Service interface {
	CreateConversation(ctx context.Context, userID string, targetUsername string) (string, error)
	GetConversationMembers(ctx context.Context, conversationID string) ([]string, error)
	SaveMessage(ctx context.Context, req *SendMessageRequest) error
	GetMessages(ctx context.Context, req *GetMessageRequest) ([]*GetMessageResponse, error)
	GetConversations(ctx context.Context, userID string) ([]string, error)
}

type service struct {
	r ChatRepository
	h *ws.Hub
}

func NewService(repo ChatRepository, hub *ws.Hub) Service {
	return &service{
		r: repo,
		h: hub,
	}
}

func (s *service) CreateConversation(ctx context.Context, userID, targetUsername string) (string, error) {
	if len(userID) == 0 || len(targetUsername) == 0 {
		return "", ErrInvalidArguments
	}

	targetID, err := s.r.GetUserByUsername(ctx, targetUsername)
	if err != nil {
		return "", err
	}

	resp, err := s.r.CreateConversation(ctx, userID, targetID)
	if err != nil {
		return "", err
	}

	s.h.SendToUser(targetID, ws.Event{
		EventType: "new_message",
		Payload:   ws.MessagePayload{},
	})

	return resp, nil
}

func (s *service) GetConversationMembers(ctx context.Context, conversationID string) ([]string, error) {
	if len(conversationID) == 0 {
		return nil, ErrInvalidArguments
	}

	res, err := s.r.GetConversationMembers(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	return res, err
}

func (s *service) SaveMessage(ctx context.Context, req *SendMessageRequest) error {
	if req == nil {
		return ErrInvalidArguments
	}

	err := s.r.SaveMessage(ctx, req)
	if err != nil {
		return err
	}

	members, err := s.r.GetConversationMembers(ctx, req.ConversationId)
	if err != nil {
		return err
	}

	for _, v := range members {
		s.h.SendToUser(v, ws.Event{
			EventType: "new_message",
			Payload: ws.MessagePayload{
				From:           req.SenderID,
				Message:        req.Content,
				ConversationID: req.ConversationId,
				CreatedAt:      time.Now(),
			},
		})
	}

	return nil
}

func (s *service) GetMessages(ctx context.Context, req *GetMessageRequest) ([]*GetMessageResponse, error) {
	if req == nil {
		return nil, ErrInvalidArguments
	}

	resp, err := s.r.GetMessages(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *service) GetConversations(ctx context.Context, userID string) ([]string, error) {
	if len(userID) == 0 {
		return nil, ErrInvalidArguments
	}

	res, err := s.r.GetConversations(ctx, userID)
	if err != nil {
		return nil, err
	}

	return res, nil
}
