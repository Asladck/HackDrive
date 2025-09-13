package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Error struct {
	Message string `json:"message"`
}

type statusResponse struct {
	Status string `json:"status"`
}

type statusFloat struct {
	Status float64 `json:"status"`
}

// NewErrorResponse logs error and returns error response
func NewErrorResponse(c *gin.Context, statusCode int, message string) {
	logrus.Error()
	c.AbortWithStatusJSON(statusCode, Error{message})
}
