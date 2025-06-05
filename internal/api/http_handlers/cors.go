package http_handlers

import (
	"github.com/gin-gonic/gin"
)

type CorsHandler struct {
}

func NewCorsHandler() *CorsHandler {
	return &CorsHandler{}
}

func (h CorsHandler) Allowed(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type")

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(204)
		return
	}

	c.Next()
}
