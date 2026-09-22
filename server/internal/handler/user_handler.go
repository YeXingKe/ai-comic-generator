package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/ai-comic-generator/server/internal/common"
	"github.com/ai-comic-generator/server/internal/model"
	"github.com/ai-comic-generator/server/internal/service"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// 作用：HTTP 接口层（Handler / Controller） 文件：负责接收前端请求、调用 UserService、返回统一 JSON

// UserHandler 用户处理器
type UserHandler struct {
   svc *service.UserService
}
  
// NewUserHandler 创建用户处理器
func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// Register 用户注册
// @Summary      用户注册
// @Description  注册新用户，账号至少 4 位，密码至少 8 位
// @Tags         用户
// @Accept       json
// @Produce      json
// @Param        body  body      model.RegisterRequest  true  "注册参数"
// @Success      200   {object}  common.BaseResponse{data=int64}  "data 为新用户 ID"
// @Failure      200   {object}  common.BaseResponse  "业务错误（参数错误、账号重复等）"
// @Router       /user/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, common.Error(common.ErrParams))
		return
	}

	userID, err := h.svc.Register(&req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, common.Success(userID))
}

// Login 用户登录
// @Summary      用户登录
// @Description  登录成功后写入 Session Cookie（session），后续受保护接口需携带
// @Tags         用户
// @Accept       json
// @Produce      json
// @Param        body  body      model.LoginRequest  true  "登录参数"
// @Success      200   {object}  common.BaseResponse{data=model.LoginUser}
// @Failure      200   {object}  common.BaseResponse  "业务错误（账号或密码错误等）"
// @Router       /user/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	req, err := bindLoginRequest(c)
	if err != nil {
		c.JSON(http.StatusOK, common.Error(common.ErrParams.WithMessage("请提交 JSON，字段为 userAccount、userPassword")))
		return
	}

	session := sessions.Default(c) // 当前请求的 Session
	loginUser, err := h.svc.Login(req, session) // 调用 UserService 的 Login 方法
	if err != nil {
		handleError(c, err) // 统一错误处理
		return
	}

	c.JSON(http.StatusOK, common.Success(loginUser))
}

// GetLoginUser 获取当前登录用户
// @Summary      获取当前登录用户
// @Description  从 Session 读取当前登录用户信息，含配额等
// @Tags         用户
// @Produce      json
// @Success      200  {object}  common.BaseResponse{data=model.LoginUser}
// @Failure      200  {object}  common.BaseResponse  "未登录"
// @Security     SessionCookie
// @Router       /user/info [get]
func (h *UserHandler) GetLoginUser(c *gin.Context) {
	session := sessions.Default(c)
	user, err := h.svc.GetLoginUser(session)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, common.Success(user.ToLoginUser()))
}

// Logout 用户注销
// @Summary      用户注销
// @Description  清除 Session，退出登录
// @Tags         用户
// @Produce      json
// @Success      200  {object}  common.BaseResponse{data=bool}
// @Failure      200  {object}  common.BaseResponse  "业务错误"
// @Security     SessionCookie
// @Router       /user/logout [post]
func (h *UserHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	if err := h.svc.Logout(session); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, common.Success(true))
}

// UpdateProfile 更新当前登录用户资料
// @Summary      更新个人资料
// @Tags         用户
// @Accept       json
// @Produce      json
// @Param        body  body      model.UpdateProfileRequest  true  "昵称/头像/简介"
// @Success      200   {object}  common.BaseResponse{data=model.LoginUser}
// @Failure      200   {object}  common.BaseResponse
// @Security     SessionCookie
// @Router       /user/profile/update [post]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req model.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, common.Error(common.ErrParams))
		return
	}

	session := sessions.Default(c)
	loginUser, err := h.svc.UpdateProfile(session, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, common.Success(loginUser))
}

// UpdatePassword 修改当前登录用户密码
// @Summary      修改密码
// @Tags         用户
// @Accept       json
// @Produce      json
// @Param        body  body      model.UpdatePasswordRequest  true  "原密码与新密码"
// @Success      200   {object}  common.BaseResponse{data=bool}
// @Failure      200   {object}  common.BaseResponse
// @Security     SessionCookie
// @Router       /user/password/update [post]
func (h *UserHandler) UpdatePassword(c *gin.Context) {
	var req model.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, common.Error(common.ErrParams))
		return
	}

	session := sessions.Default(c)
	if err := h.svc.UpdatePassword(session, &req); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, common.Success(true))
}

// Add 创建用户（管理员）
// @Summary      创建用户
// @Description  管理员创建用户，需 admin 角色
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        body  body      model.AddUserRequest  true  "创建参数"
// @Success      200   {object}  common.BaseResponse{data=int64}  "data 为新用户 ID"
// @Failure      200   {object}  common.BaseResponse  "业务错误（无权限、参数错误等）"
// @Security     SessionCookie
// @Router       /user/add [post]
func (h *UserHandler) Add(c *gin.Context) {
	var req model.AddUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, common.Error(common.ErrParams))
		return
	}

	userID, err := h.svc.Create(&req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, common.Success(userID))
}

// Get 根据 ID 获取用户（管理员）；当前未在 router 注册，仅供内部或后续启用。
func (h *UserHandler) Get(c *gin.Context) {
	var req struct {
		ID int64 `form:"id" binding:"required,gt=0"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusOK, common.Error(common.ErrParams))
		return
	}

	user, err := h.svc.GetByID(req.ID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, common.Success(user))
}

// GetVO 根据 ID 获取用户脱敏信息；当前未在 router 注册。
func (h *UserHandler) GetVO(c *gin.Context) {
	var req struct {
		ID int64 `form:"id" binding:"required,gt=0"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusOK, common.Error(common.ErrParams))
		return
	}

	user, err := h.svc.GetByID(req.ID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, common.Success(user.ToUserInfo()))
}

// Delete 删除用户（管理员）
// @Summary      删除用户
// @Description  管理员逻辑删除用户，需 admin 角色
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        body  body      model.DeleteRequest  true  "删除参数"
// @Success      200   {object}  common.BaseResponse{data=bool}
// @Failure      200   {object}  common.BaseResponse  "业务错误"
// @Security     SessionCookie
// @Router       /user/delete [post]
func (h *UserHandler) Delete(c *gin.Context) {
	var req model.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, common.Error(common.ErrParams))
		return
	}

	if err := h.svc.Delete(req.ID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, common.Success(true))
}

// Update 更新用户（管理员）
// @Summary      更新用户
// @Description  管理员更新用户信息，需 admin 角色
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        body  body      model.UpdateUserRequest  true  "更新参数"
// @Success      200   {object}  common.BaseResponse{data=bool}
// @Failure      200   {object}  common.BaseResponse  "业务错误"
// @Security     SessionCookie
// @Router       /user/update [post]
func (h *UserHandler) Update(c *gin.Context) {
	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, common.Error(common.ErrParams))
		return
	}

	if err := h.svc.Update(&req); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, common.Success(true))
}

// ListPageVO 分页查询用户列表（管理员）
// @Summary      分页查询用户列表
// @Description  管理员分页查询用户，支持多条件筛选与排序，需 admin 角色
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        body  body      model.QueryUserRequest  true  "查询参数"
// @Success      200   {object}  common.BaseResponse{data=model.PageResult}
// @Failure      200   {object}  common.BaseResponse  "业务错误"
// @Security     SessionCookie
// @Router       /user/page/vo [post]
func (h *UserHandler) ListPageVO(c *gin.Context) {
	var req model.QueryUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, common.Error(common.ErrParams))
		return
	}

	// 设置默认值
	if req.PageNum <= 0 {
		req.PageNum = common.DefaultPageNum
	}
	if req.PageSize <= 0 {
		req.PageSize = common.DefaultPageSize
	}
	if req.PageSize > common.MaxPageSize {
		req.PageSize = common.MaxPageSize
	}

	page, err := h.svc.ListByPage(&req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, common.Success(page))
}

// loginPayload 兼容 JSON / 表单，以及 username、password 等常见别名
type loginPayload struct {
	UserAccount  string `json:"userAccount" form:"userAccount"`
	UserPassword string `json:"userPassword" form:"userPassword"`
	Username     string `json:"username" form:"username"`
	Account      string `json:"account" form:"account"`
	Password     string `json:"password" form:"password"`
}

func (p loginPayload) toRequest() *model.LoginRequest {
	account := firstNonEmpty(p.UserAccount, p.Username, p.Account)
	password := firstNonEmpty(p.UserPassword, p.Password)
	return &model.LoginRequest{UserAccount: account, UserPassword: password}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

func bindLoginRequest(c *gin.Context) (*model.LoginRequest, error) {
	var payload loginPayload
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))

	trim := bytes.TrimSpace(body)
	if len(trim) > 0 && trim[0] == '{' {
		if err := json.Unmarshal(trim, &payload); err != nil {
			return nil, err
		}
		return payload.toRequest(), nil
	}

	if err := c.ShouldBind(&payload); err != nil {
		return nil, err
	}
	return payload.toRequest(), nil
}

// handleError 统一错误处理
func handleError(c *gin.Context, err error) {
	if appErr, ok := err.(*common.AppError); ok {
		c.JSON(http.StatusOK, common.Error(appErr))
	} else {
		c.JSON(http.StatusOK, common.Error(common.ErrSystem))
	}
}
