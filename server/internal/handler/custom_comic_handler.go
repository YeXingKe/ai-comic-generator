package handler

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/ai-comic-generator/server/internal/common"
	"github.com/ai-comic-generator/server/internal/middleware"
	"github.com/ai-comic-generator/server/internal/model"
	"github.com/ai-comic-generator/server/internal/service"
	"github.com/gin-gonic/gin"
)

// CustomComicHandler 自定义创作 HTTP 入口
type CustomComicHandler struct {
	svc *service.CustomComicService
}

func NewCustomComicHandler(svc *service.CustomComicService) *CustomComicHandler {
	return &CustomComicHandler{svc: svc}
}

// Create POST /comic/custom/create
func (h *CustomComicHandler) Create(c *gin.Context) {
	loginUser, ok := middleware.GetLoginUserFromContext(c)
	if !ok {
		c.JSON(http.StatusOK, common.Error(common.ErrNotLogin))
		return
	}

	var req model.CreateCustomComicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, common.Error(common.ErrParams))
		return
	}

	taskID, err := h.svc.Create(loginUser.ID, &req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, common.Success(gin.H{"taskId": taskID}))
}

// Get GET /comic/custom/get?taskId=
func (h *CustomComicHandler) Get(c *gin.Context) {
	loginUser, ok := middleware.GetLoginUserFromContext(c)
	if !ok {
		c.JSON(http.StatusOK, common.Error(common.ErrNotLogin))
		return
	}
	taskID := c.Query("taskId")
	if taskID == "" {
		c.JSON(http.StatusOK, common.Error(common.ErrParams))
		return
	}

	info, err := h.svc.GetForUser(taskID, loginUser.ID, isAdminUser(loginUser))
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, common.Success(info))
}

// ListPage POST /comic/custom/page
func (h *CustomComicHandler) ListPage(c *gin.Context) {
	loginUser, ok := middleware.GetLoginUserFromContext(c)
	if !ok {
		c.JSON(http.StatusOK, common.Error(common.ErrNotLogin))
		return
	}

	var req model.QueryCustomComicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, common.Error(common.ErrParams))
		return
	}

	if !isAdminUser(loginUser) {
		req.UserID = &loginUser.ID
	}

	page, err := h.svc.ListByPage(&req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, common.Success(page))
}

// DownloadZip GET /comic/custom/download?taskId=  打包全部分镜为 zip
func (h *CustomComicHandler) DownloadZip(c *gin.Context) {
	loginUser, ok := middleware.GetLoginUserFromContext(c)
	if !ok {
		c.JSON(http.StatusOK, common.Error(common.ErrNotLogin))
		return
	}
	taskID := c.Query("taskId")
	if taskID == "" {
		c.JSON(http.StatusOK, common.Error(common.ErrParams))
		return
	}

	var buf bytes.Buffer
	filename, err := h.svc.WritePanelsZip(&buf, taskID, loginUser.ID, isAdminUser(loginUser))
	if err != nil {
		handleError(c, err)
		return
	}
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("Content-Length", fmt.Sprintf("%d", buf.Len()))
	c.Data(http.StatusOK, "application/zip", buf.Bytes())
}
