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

// Dashboard 管理端统计看板
// @Summary      管理端统计看板
// @Tags         统计
// @Produce      json
// @Param        range  query  string  false  "7d|30d|90d，默认 30d"
// @Success      200    {object}  common.BaseResponse{data=model.StatDashboard}
// @Failure      200    {object}  common.BaseResponse
// @Security     SessionCookie
// @Router       /stat/dashboard [get]
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
