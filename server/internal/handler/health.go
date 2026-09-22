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

// Check 健康检查
// @Summary      健康检查
// @Tags         系统
// @Produce      json
// @Success      200  {object}  common.BaseResponse
// @Router       /health [get]
func (h *HealthHandler) Check(c *gin.Context) {
	c.JSON(http.StatusOK, common.Success(gin.H{"status": "ok"}))
}
