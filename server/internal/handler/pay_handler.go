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

// ListPackages 充值套餐与支付能力说明
// @Summary      充值套餐列表
// @Tags         支付
// @Produce      json
// @Success      200  {object}  common.BaseResponse
// @Security     SessionCookie
// @Router       /pay/packages [get]
func (h *PayHandler) ListPackages(c *gin.Context) {
	c.JSON(http.StatusOK, common.Success(h.svc.GetCatalog()))
}

// ListPage 我的充值订单分页
// @Summary      我的充值订单分页
// @Tags         支付
// @Accept       json
// @Produce      json
// @Param        body  body  model.PayOrderPageRequest  true  "分页"
// @Success      200   {object}  common.BaseResponse{data=model.PageResult}
// @Security     SessionCookie
// @Router       /pay/order/page [post]
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

// Create 创建充值订单
// @Summary      创建充值订单
// @Tags         支付
// @Accept       json
// @Produce      json
// @Param        body  body  model.CreatePayOrderRequest  true  "套餐 code 等"
// @Success      200   {object}  common.BaseResponse
// @Security     SessionCookie
// @Router       /pay/order [post]
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

// Get 查询单笔订单（仅读库，不打支付宝）
// @Summary      查询充值订单
// @Tags         支付
// @Produce      json
// @Param        orderNo  query  string  true  "订单号"
// @Success      200      {object}  common.BaseResponse
// @Security     SessionCookie
// @Router       /pay/order [get]
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

// Sync 主动查支付宝并尝试入账（支付回跳 / 扫码确认后调用）
// @Summary      同步支付宝订单状态
// @Tags         支付
// @Accept       json
// @Produce      json
// @Param        body  body  object{orderNo=string}  true  "订单号"
// @Success      200   {object}  common.BaseResponse
// @Security     SessionCookie
// @Router       /pay/order/sync [post]
func (h *PayHandler) Sync(c *gin.Context) {
	u, ok := middleware.GetLoginUserFromContext(c)
	if !ok {
		c.JSON(http.StatusOK, common.Error(common.ErrNotLogin))
		return
	}
	var req struct {
		OrderNo string `json:"orderNo" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, common.Error(common.ErrParams))
		return
	}
	o, err := h.svc.SyncMine(u.ID, req.OrderNo)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, common.Success(o.ToVO()))
}

// MockPay 模拟支付成功（开发）
// @Summary      模拟支付成功
// @Tags         支付
// @Accept       json
// @Produce      json
// @Param        body  body  object{orderNo=string}  true  "订单号"
// @Success      200   {object}  common.BaseResponse{data=bool}
// @Security     SessionCookie
// @Router       /pay/mock-pay [post]
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

// AlipayNotify 支付宝异步通知（纯文本 success/fail）
// @Summary      支付宝支付回调
// @Tags         支付
// @Accept       x-www-form-urlencoded
// @Produce      plain
// @Success      200  {string}  string  "success"
// @Router       /pay/notify/alipay [post]
func (h *PayHandler) AlipayNotify(c *gin.Context) {
	if err := h.svc.HandleAlipayNotify(c.Request); err != nil {
		c.String(http.StatusOK, "fail")
		return
	}
	c.String(http.StatusOK, "success")
}

// AdminListPackagePlans 管理端充值方案分页
// @Summary      管理端充值方案分页
// @Tags         支付管理
// @Accept       json
// @Produce      json
// @Param        body  body  model.PayPackagePlanPageRequest  true  "分页"
// @Success      200   {object}  common.BaseResponse{data=model.PageResult}
// @Security     SessionCookie
// @Router       /pay/admin/package/page [post]
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

// @Summary      新增充值方案
// @Tags         支付管理
// @Accept       json
// @Produce      json
// @Param        body  body  model.AddPayPackagePlanRequest  true  "方案"
// @Success      200   {object}  common.BaseResponse{data=int64}
// @Security     SessionCookie
// @Router       /pay/admin/package/add [post]
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

// @Summary      更新充值方案
// @Tags         支付管理
// @Accept       json
// @Produce      json
// @Param        body  body  model.UpdatePayPackagePlanRequest  true  "方案"
// @Success      200   {object}  common.BaseResponse{data=bool}
// @Security     SessionCookie
// @Router       /pay/admin/package/update [post]
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

// @Summary      删除充值方案
// @Tags         支付管理
// @Accept       json
// @Produce      json
// @Param        body  body  model.DeletePayPackagePlanRequest  true  "id"
// @Success      200   {object}  common.BaseResponse{data=bool}
// @Security     SessionCookie
// @Router       /pay/admin/package/delete [post]
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

// @Summary      全站支付订单分页
// @Tags         支付管理
// @Accept       json
// @Produce      json
// @Param        body  body  model.AdminPayOrderPageRequest  true  "筛选与分页"
// @Success      200   {object}  common.BaseResponse{data=model.PageResult}
// @Security     SessionCookie
// @Router       /pay/admin/order/page [post]
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

// @Summary      收款记录分页（已支付）
// @Tags         支付管理
// @Accept       json
// @Produce      json
// @Param        body  body  model.AdminPayOrderPageRequest  true  "筛选与分页"
// @Success      200   {object}  common.BaseResponse{data=model.PageResult}
// @Security     SessionCookie
// @Router       /pay/admin/receipt/page [post]
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