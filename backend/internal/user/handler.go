package user

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler interface {
	GetMe(c *gin.Context)
	GetUser(c *gin.Context)
	UpdateMe(c *gin.Context)
}

type userHandler struct {
	s UserService
}

func NewUserHandler(s UserService) UserHandler {
	return &userHandler{
		s: s,
	}
}

func (h *userHandler) GetMe(c *gin.Context) {
	v, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id not found in context"})
		return
	}

	resp, err := h.s.GetMe(c, &GetMeRequest{Userid: v.(string)})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *userHandler) GetUser(c *gin.Context) {
	var req GetUserRequest

	req.Username = c.Query("username")

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Userid = c.Param("user_id")

	res, err := h.s.GetUser(c, &req)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		} else if errors.Is(err, ErrInvalidToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *userHandler) UpdateMe(c *gin.Context) {
	var req UpdateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	v, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id not found in context"})
		return
	}

	req.Userid = v.(string)

	err := h.s.UpdateMe(c, &req)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		} else if errors.Is(err, ErrInvalidToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"error": nil})
}
