package handler

import (
	"net/http"

	"github.com/ai-comic-generator/server/internal/common"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Check(c *gin.Context) {
	c.JSON(http.StatusOK, common.Success(gin.H{"status": "ok"}))
}
