package handler // HTTP 请求处理层：漫画任务相关接口

import (
	"net/http" // HTTP 状态码（统一返回 200，业务码在 JSON body 中）

	"github.com/ai-comic-generator/server/internal/common" // 统一响应封装、错误码、角色常量
	"github.com/ai-comic-generator/server/internal/middleware" // 从 Context 读取 AuthCheck 注入的登录用户
	"github.com/ai-comic-generator/server/internal/model"  // 请求/响应数据结构
	"github.com/ai-comic-generator/server/internal/service" // 漫画业务逻辑层
	"github.com/gin-gonic/gin" // Web 框架，提供 Context、路由、JSON 绑定
)

// ComicHandler 漫画任务 HTTP 入口
type ComicHandler struct {
	svc *service.ComicService // 漫画任务业务服务（创建、查询、分页）
}

// NewComicHandler 创建漫画 Handler，注入业务服务依赖
func NewComicHandler(svc *service.ComicService) *ComicHandler {
	return &ComicHandler{svc: svc} // 保存服务引用供各接口方法调用
}

// Create 创建漫画生成任务（异步六步流水线）
func (h *ComicHandler) Create(c *gin.Context) {
	loginUser, ok := middleware.GetLoginUserFromContext(c) // 从 AuthCheck 中间件注入的 Context 读取登录用户
	if !ok { // 未挂中间件或登录态异常（正常流程下 AuthCheck 已拦截）
		c.JSON(http.StatusOK, common.Error(common.ErrNotLogin)) // 防御性返回未登录错误
		return // 终止处理
	}

	var req model.CreateComicRequest                               // 声明创建请求体变量
	if err := c.ShouldBindJSON(&req); err != nil {                 // 将 JSON 请求体绑定到结构体并校验 required 字段
		c.JSON(http.StatusOK, common.Error(common.ErrParams)) // 参数缺失或格式错误
		return // 终止处理
	}

	taskID, err := h.svc.Create(loginUser.ID, &req) // 调用业务层创建任务，后台 goroutine 跑六步流水线
	if err != nil { // 创建失败（如数据库写入错误）
		handleError(c, err) // 将业务错误映射为统一 JSON 响应
		return // 终止处理
	}
	c.JSON(http.StatusOK, common.Success(gin.H{"taskId": taskID})) // 返回新任务的 UUID 供前端轮询
}

// Get 查询任务详情
func (h *ComicHandler) Get(c *gin.Context) {
	loginUser, ok := middleware.GetLoginUserFromContext(c) // 从 AuthCheck 中间件注入的 Context 读取登录用户
	if !ok { // 未挂中间件或登录态异常
		c.JSON(http.StatusOK, common.Error(common.ErrNotLogin)) // 防御性返回未登录错误
		return // 终止处理
	}
	taskID := c.Query("taskId") // 从 URL 查询参数获取任务 ID
	if taskID == "" { // 未传 taskId
		c.JSON(http.StatusOK, common.Error(common.ErrParams)) // 返回参数错误
		return // 终止处理
	}

	isAdmin := isAdminUser(loginUser)                              // 根据用户角色判断是否管理员
	info, err := h.svc.GetForUser(taskID, loginUser.ID, isAdmin) // 查询详情；非管理员只能看自己的任务
	if err != nil { // 任务不存在或无权限
		handleError(c, err) // 映射为 NOT_FOUND / NO_AUTH 等错误响应
		return // 终止处理
	}
	c.JSON(http.StatusOK, common.Success(info)) // 返回漫画任务详情（含各步产物 JSON）
}

// ListPage 分页查询漫画任务
func (h *ComicHandler) ListPage(c *gin.Context) {
	loginUser, ok := middleware.GetLoginUserFromContext(c) // 从 AuthCheck 中间件注入的 Context 读取登录用户
	if !ok { // 未挂中间件或登录态异常
		c.JSON(http.StatusOK, common.Error(common.ErrNotLogin)) // 防御性返回未登录错误
		return // 终止处理
	}

	var req model.QueryComicRequest                        // 声明分页查询请求体
	if err := c.ShouldBindJSON(&req); err != nil {         // 绑定 JSON（pageNum、pageSize、筛选条件等）
		c.JSON(http.StatusOK, common.Error(common.ErrParams)) // 参数格式错误
		return // 终止处理
	}

	if !isAdminUser(loginUser) { // 普通用户不能查他人任务
		req.UserID = &loginUser.ID // 强制限定为当前用户 ID
	}

	page, err := h.svc.ListByPage(&req) // 调用业务层分页查询
	if err != nil { // 数据库查询失败
		handleError(c, err) // 映射为系统错误响应
		return // 终止处理
	}
	c.JSON(http.StatusOK, common.Success(page)) // 返回分页结果（total、records、pageNum、pageSize）
}

// isAdminUser 根据中间件已加载的用户对象判断是否为管理员（无需再次查库）
func isAdminUser(user *model.User) bool {
	return user.UserRole == common.AdminRole // 角色等于 admin 则为管理员
}
