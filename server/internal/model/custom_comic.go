package model

import (
	"encoding/json"
	"time"
)

// CustomComic 自定义创作任务（独立于自动化 comic 流水线）
type CustomComic struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID       string    `gorm:"column:taskId;uniqueIndex:uk_taskId" json:"taskId"`
	UserID       int64     `gorm:"column:userId;index:idx_userId" json:"userId"`
	Prompt       string    `gorm:"column:prompt;type:text" json:"prompt"`
	AspectRatio  string    `gorm:"column:aspectRatio;default:16:9" json:"aspectRatio"`
	ImageBackend string    `gorm:"column:imageBackend;default:hunyuan" json:"imageBackend"`
	PanelCount   int       `gorm:"column:panelCount;default:4" json:"panelCount"`
	PanelImages  *string   `gorm:"column:panelImages;type:json" json:"panelImages"`
	Status       string    `gorm:"column:status;default:PENDING;index:idx_status" json:"status"`
	ErrorMessage *string   `gorm:"column:errorMessage;type:text" json:"errorMessage"`
	CreateTime   time.Time `gorm:"column:createTime;autoCreateTime" json:"createTime"`
	UpdateTime   time.Time `gorm:"column:updateTime;autoUpdateTime" json:"updateTime"`
	IsDelete     int       `gorm:"column:isDelete;default:0" json:"-"`
}

func (CustomComic) TableName() string {
	return "custom_comic"
}

const (
	CustomComicStatusPending    = "PENDING"
	CustomComicStatusProcessing = "PROCESSING"
	CustomComicStatusCompleted  = "COMPLETED"
	CustomComicStatusFailed     = "FAILED"
)

const (
	AspectRatio1x1  = "1:1"
	AspectRatio16x9 = "16:9"
	AspectRatio9x16 = "9:16"
	AspectRatio2x3  = "2:3" // 小红书竖版常见比例
)

// CreateCustomComicRequest 创建自定义创作任务
type CreateCustomComicRequest struct {
	Prompt       string `json:"prompt" binding:"required"`
	AspectRatio  string `json:"aspectRatio" example:"16:9" enums:"1:1,16:9,9:16,2:3"`
	ImageBackend string `json:"imageBackend" example:"hunyuan" enums:"hunyuan,openai_image_1k,openai_image_4k"`
	PanelCount   int    `json:"panelCount" example:"4"`
}

// QueryCustomComicRequest 分页查询自定义创作
type QueryCustomComicRequest struct {
	UserID   *int64  `json:"userId"`
	Status   *string `json:"status" enums:"PENDING,PROCESSING,COMPLETED,FAILED"`
	PageNum  int64   `json:"pageNum" example:"1"`
	PageSize int64   `json:"pageSize" example:"10"`
}

// CustomComicPageResult 自定义创作分页
type CustomComicPageResult struct {
	Total    int64             `json:"total"`
	Records  []CustomComicInfo `json:"records"`
	PageNum  int64             `json:"pageNum"`
	PageSize int64             `json:"pageSize"`
}

// CustomComicInfo API 响应
type CustomComicInfo struct {
	ID           int64              `json:"id"`
	TaskID       string             `json:"taskId"`
	UserID       int64              `json:"userId"`
	UserAccount  string             `json:"userAccount,omitempty"`
	UserName     *string            `json:"userName,omitempty"`
	Prompt       string             `json:"prompt"`
	AspectRatio  string             `json:"aspectRatio"`
	ImageBackend string             `json:"imageBackend"`
	PanelCount   int                `json:"panelCount"`
	PanelImages  []PanelImageResult `json:"panelImages"`
	Status       string             `json:"status"`
	ErrorMessage *string            `json:"errorMessage"`
	CreateTime   time.Time          `json:"createTime"`
	UpdateTime   time.Time          `json:"updateTime"`
}

// ToCustomComicInfo 实体转 API 结构
func (c *CustomComic) ToCustomComicInfo() *CustomComicInfo {
	if c == nil {
		return nil
	}
	info := &CustomComicInfo{
		ID:           c.ID,
		TaskID:       c.TaskID,
		UserID:       c.UserID,
		Prompt:       c.Prompt,
		AspectRatio:  c.AspectRatio,
		ImageBackend: c.ImageBackend,
		PanelCount:   c.PanelCount,
		PanelImages:  []PanelImageResult{},
		Status:       c.Status,
		ErrorMessage: c.ErrorMessage,
		CreateTime:   c.CreateTime,
		UpdateTime:   c.UpdateTime,
	}
	if c.PanelImages != nil && *c.PanelImages != "" {
		var panels []PanelImageResult
		if err := json.Unmarshal([]byte(*c.PanelImages), &panels); err == nil {
			info.PanelImages = panels
		}
	}
	return info
}
