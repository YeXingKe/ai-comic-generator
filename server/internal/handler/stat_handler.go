package handler

import (
	"net/http"

	"github.com/ai-comic-generator/server/internal/common"
	"github.com/ai-comic-generator/server/internal/model"
	"github.com/ai-comic-generator/server/internal/service"
	"github.com/gin-gonic/gin"
)

type StatHandler struct {
	svc *service.StatService
}

func NewStatHandler(svc *service.StatService) *StatHandler {
	return &StatHandler{svc: svc}
}

// Dashboard godoc
// GET /stat/dashboard?range=7d|30d|90d
func (h *StatHandler) Dashboard(c *gin.Context) {
	var req model.StatQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusOK, common.Error(common.ErrParams))
		return
	}
	if !req.Range.IsValid() {
		req.Range = model.StatRange30d
	}

	data, err := h.svc.GetDashboard(&req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, common.Success(data))
}
