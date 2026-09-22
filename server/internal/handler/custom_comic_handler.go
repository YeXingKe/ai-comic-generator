package handler

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

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
// 支持 application/json 或 multipart/form-data（字段 references：角色参考图，可多张）
// @Summary      创建自定义漫画
// @Tags         自定义漫画
// @Accept       json,multipart/form-data
// @Produce      json
// @Success      200  {object}  common.BaseResponse
// @Security     SessionCookie
// @Router       /comic/custom/create [post]
func (h *CustomComicHandler) Create(c *gin.Context) {
	loginUser, ok := middleware.GetLoginUserFromContext(c)
	if !ok {
		c.JSON(http.StatusOK, common.Error(common.ErrNotLogin))
		return
	}

	req, refs, err := bindCreateCustomComic(c)
	if err != nil {
		c.JSON(http.StatusOK, common.Error(common.ErrParams.WithMessage(err.Error())))
		return
	}

	taskID, err := h.svc.Create(loginUser.ID, req, refs)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, common.Success(gin.H{"taskId": taskID}))
}

func bindCreateCustomComic(c *gin.Context) (*model.CreateCustomComicRequest, []*multipart.FileHeader, error) {
	ct := c.ContentType()
	if strings.Contains(ct, "multipart/form-data") {
		req := &model.CreateCustomComicRequest{
			Prompt:       c.PostForm("prompt"),
			AspectRatio:  c.PostForm("aspectRatio"),
			ImageBackend: c.PostForm("imageBackend"),
		}
		if v := c.PostForm("panelCount"); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil {
				return nil, nil, fmt.Errorf("panelCount 无效")
			}
			req.PanelCount = n
		}
		form, err := c.MultipartForm()
		if err != nil {
			return nil, nil, fmt.Errorf("解析上传表单失败")
		}
		var refs []*multipart.FileHeader
		if form != nil {
			refs = form.File["references"]
		}
		return req, refs, nil
	}

	var req model.CreateCustomComicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, nil, fmt.Errorf("参数错误")
	}
	return &req, nil, nil
}

// Get GET /comic/custom/get?taskId=
// @Summary      查询自定义任务
// @Tags         自定义漫画
// @Produce      json
// @Param        taskId  query  string  true  "任务 ID"
// @Success      200     {object}  common.BaseResponse
// @Security     SessionCookie
// @Router       /comic/custom/get [get]
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
// @Summary      自定义任务分页
// @Tags         自定义漫画
// @Accept       json
// @Produce      json
// @Param        body  body  model.QueryCustomComicRequest  true  "分页与筛选"
// @Success      200   {object}  common.BaseResponse{data=model.CustomComicPageResult}
// @Security     SessionCookie
// @Router       /comic/custom/page [post]
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
// @Summary      下载分镜 ZIP
// @Tags         自定义漫画
// @Produce      application/zip
// @Param        taskId  query  string  true  "任务 ID"
// @Success      200     {file}   binary  "ZIP 文件"
// @Security     SessionCookie
// @Router       /comic/custom/download [get]
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

// RegeneratePanel POST /comic/custom/regenerate-panel  重绘某一格
// @Summary      自定义单格重绘
// @Tags         自定义漫画
// @Accept       json
// @Produce      json
// @Param        body  body  model.RegenerateCustomPanelRequest  true  "任务与分镜序号"
// @Success      200   {object}  common.BaseResponse
// @Security     SessionCookie
// @Router       /comic/custom/regenerate-panel [post]
func (h *CustomComicHandler) RegeneratePanel(c *gin.Context) {
	loginUser, ok := middleware.GetLoginUserFromContext(c)
	if !ok {
		c.JSON(http.StatusOK, common.Error(common.ErrNotLogin))
		return
	}
	var req model.RegenerateCustomPanelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, common.Error(common.ErrParams))
		return
	}
	info, err := h.svc.RegeneratePanel(loginUser.ID, &req, isAdminUser(loginUser))
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, common.Success(info))
}
