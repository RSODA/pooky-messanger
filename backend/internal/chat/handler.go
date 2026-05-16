package chat

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ChatHandler interface {
	CreateConversation(c *gin.Context)
	GetConversations(c *gin.Context)
	GetMessages(c *gin.Context)
	SendMessage(c *gin.Context)
}

type chatHandler struct {
	s Service
}

func NewChatHandler(s Service) ChatHandler {
	return &chatHandler{s: s}
}

func (h *chatHandler) CreateConversation(c *gin.Context) {
	var req CreateConversationRequest

	v, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token invalid"})
		return
	}

	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad arguments"})
		return
	}

	res, err := h.s.CreateConversation(c, v.(string), req.TargetUsername)
	if err != nil {
		if errors.Is(err, ErrInvalidArguments) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

func (h *chatHandler) GetConversations(c *gin.Context) {
	v, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token invalid"})
		return
	}

	conversations, err := h.s.GetConversations(c, v.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, conversations)
}

func (h *chatHandler) GetMessages(c *gin.Context) {
	var req GetMessageRequest

	v, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token invalid"})
		return
	}

	req.UserID = v.(string)
	req.ConversationId = c.Param("id")

	messages, err := h.s.GetMessages(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}

func (h *chatHandler) SendMessage(c *gin.Context) {
	var req SendMessageRequest

	v, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token invalid"})
		return
	}

	req.ConversationId = c.Param("id")
	req.SenderID = v.(string)

	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad arguments"})
		return
	}

	err := h.s.SaveMessage(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message saved"})
}
