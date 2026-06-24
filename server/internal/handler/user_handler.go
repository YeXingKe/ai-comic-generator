package handler

import (
	"net/http"

	"github.com/ai-comic-generator/server/internal/common"
	"github.com/ai-comic-generator/server/internal/model"
	"github.com/ai-comic-generator/server/internal/service"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Check(c *gin.Context) {
	c.JSON(http.StatusOK, common.Success(gin.H{"status": "ok"}))
}

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req model.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, common.Error(40000, "请求参数错误"))
		return
	}

	id, err := h.userService.Register(&req)
	if err != nil {
		switch err {
		case service.ErrUserExists:
			c.JSON(http.StatusOK, common.Error(40001, "账号已存在"))
		case service.ErrInvalidParams:
			c.JSON(http.StatusOK, common.Error(40000, "两次密码不一致"))
		default:
			c.JSON(http.StatusOK, common.Error(50000, "注册失败"))
		}
		return
	}
	c.JSON(http.StatusOK, common.Success(id))
}

func (h *UserHandler) Login(c *gin.Context) {
	var req model.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, common.Error(40000, "请求参数错误"))
		return
	}

	vo, err := h.userService.Login(&req)
	if err != nil {
		switch err {
		case service.ErrUserNotFound, service.ErrPasswordMismatch:
			c.JSON(http.StatusOK, common.Error(40100, "账号或密码错误"))
		default:
			c.JSON(http.StatusOK, common.Error(50000, "登录失败"))
		}
		return
	}

	session := sessions.Default(c)
	session.Set(common.SessionUserKey, vo.ID)
	_ = session.Save()

	c.JSON(http.StatusOK, common.Success(vo))
}

func (h *UserHandler) GetLoginUser(c *gin.Context) {
	userID, ok := GetLoginUserID(c)
	if !ok {
		c.JSON(http.StatusOK, common.Error(40100, "未登录"))
		return
	}

	vo, err := h.userService.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusOK, common.Error(40100, "未登录"))
		return
	}
	c.JSON(http.StatusOK, common.Success(vo))
}

func (h *UserHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete(common.SessionUserKey)
	_ = session.Save()
	c.JSON(http.StatusOK, common.Success(true))
}

func (h *UserHandler) ListPageVO(c *gin.Context) {
	var req model.UserQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, common.Error(40000, "请求参数错误"))
		return
	}

	result, err := h.userService.ListPage(&req)
	if err != nil {
		c.JSON(http.StatusOK, common.Error(50000, "查询失败"))
		return
	}
	c.JSON(http.StatusOK, common.Success(result))
}

func (h *UserHandler) Delete(c *gin.Context) {
	var req model.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, common.Error(40000, "请求参数错误"))
		return
	}

	if err := h.userService.Delete(req.ID); err != nil {
		c.JSON(http.StatusOK, common.Error(50000, "删除失败"))
		return
	}
	c.JSON(http.StatusOK, common.Success(true))
}

func GetLoginUserID(c *gin.Context) (uint, bool) {
	session := sessions.Default(c)
	value := session.Get(common.SessionUserKey)
	if value == nil {
		return 0, false
	}
	switch id := value.(type) {
	case uint:
		return id, true
	case int:
		return uint(id), true
	case int64:
		return uint(id), true
	case float64:
		return uint(id), true
	default:
		return 0, false
	}
}
