package handler

import (
	"net/http"

	"github.com/ai-comic-generator/server/internal/common"
	"github.com/ai-comic-generator/server/internal/middleware"
	"github.com/ai-comic-generator/server/internal/model"
	"github.com/ai-comic-generator/server/internal/service"
	"github.com/gin-gonic/gin"
)

type PayHandler struct{ svc *service.PayService }

func NewPayHandler(svc *service.PayService) *PayHandler { return &PayHandler{svc: svc} }

func (h *PayHandler) ListPackages(c *gin.Context) {
	c.JSON(http.StatusOK, common.Success(h.svc.GetCatalog()))
}

func (h *PayHandler) ListPage(c *gin.Context) {
	u, ok := middleware.GetLoginUserFromContext(c)
	if !ok {
		c.JSON(http.StatusOK, common.Error(common.ErrNotLogin))
		return
	}
	var req model.PayOrderPageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, common.Error(common.ErrParams))
		return
	}
	result, err := h.svc.ListMine(u.ID, req.PageNum, req.PageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, common.Success(result))
}

func (h *PayHandler) Create(c *gin.Context) {
	u, ok := middleware.GetLoginUserFromContext(c)
	if !ok {
		c.JSON(http.StatusOK, common.Error(common.ErrNotLogin))
		return
	}
	var req model.CreatePayOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, common.Error(common.ErrParams))
		return
	}
	vo, err := h.svc.CreateOrder(u.ID, &req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, common.Success(vo))
}

func (h *PayHandler) Get(c *gin.Context) {
	u, ok := middleware.GetLoginUserFromContext(c)
	if !ok {
		c.JSON(http.StatusOK, common.Error(common.ErrNotLogin))
		return
	}
	orderNo := c.Query("orderNo")
	o, err := h.svc.GetMine(u.ID, orderNo)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, common.Success(o.ToVO()))
}

func (h *PayHandler) MockPay(c *gin.Context) {
	u, ok := middleware.GetLoginUserFromContext(c)
	if !ok {
		c.JSON(http.StatusOK, common.Error(common.ErrNotLogin))
		return
	}
	var req struct {
		OrderNo string `json:"orderNo"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, common.Error(common.ErrParams))
		return
	}
	if err := h.svc.MockPay(u.ID, req.OrderNo); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, common.Success(true))
}

// AlipayNotify 必须返回纯文本 success，不要套 {code,data,message}
func (h *PayHandler) AlipayNotify(c *gin.Context) {
	if err := h.svc.HandleAlipayNotify(c.Request); err != nil {
		c.String(http.StatusOK, "fail")
		return
	}
	c.String(http.StatusOK, "success")
}

func (h *PayHandler) AdminListPackagePlans(c *gin.Context) {
	var req model.PayPackagePlanPageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, common.Error(common.ErrParams))
		return
	}
	result, err := h.svc.ListPackagePlans(req.PageNum, req.PageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, common.Success(result))
}

func (h *PayHandler) AdminAddPackagePlan(c *gin.Context) {
	var req model.AddPayPackagePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, common.Error(common.ErrParams))
		return
	}
	id, err := h.svc.AddPackagePlan(&req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, common.Success(id))
}

func (h *PayHandler) AdminUpdatePackagePlan(c *gin.Context) {
	var req model.UpdatePayPackagePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, common.Error(common.ErrParams))
		return
	}
	if err := h.svc.UpdatePackagePlan(&req); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, common.Success(true))
}

func (h *PayHandler) AdminDeletePackagePlan(c *gin.Context) {
	var req model.DeletePayPackagePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, common.Error(common.ErrParams))
		return
	}
	if err := h.svc.DeletePackagePlan(req.ID); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, common.Success(true))
}

func (h *PayHandler) AdminListOrders(c *gin.Context) {
	var req model.AdminPayOrderPageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, common.Error(common.ErrParams))
		return
	}
	result, err := h.svc.ListAdminOrders(&req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, common.Success(result))
}

func (h *PayHandler) AdminListReceipts(c *gin.Context) {
	var req model.AdminPayOrderPageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, common.Error(common.ErrParams))
		return
	}
	result, err := h.svc.ListAdminReceipts(&req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, common.Success(result))
}