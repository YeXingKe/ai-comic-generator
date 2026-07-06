package model

// CreateComicRequest 创建漫画任务
type CreateComicRequest struct {
	Topic           string  `json:"topic" binding:"required" example:"哪吒闹海"`
	UserDescription *string `json:"userDescription" example:"四格漫画，搞笑风格"`
	Style           string  `json:"style" example:"cartoon" enums:"cartoon,realistic,chibi"`
}

// QueryComicRequest 查询漫画任务
type QueryComicRequest struct {
	UserID   *int64  `json:"userId"`
	Status   *string `json:"status" enums:"PENDING,PROCESSING,COMPLETED,FAILED"`
	Phase    *string `json:"phase"`
	PageNum  int64   `json:"pageNum" example:"1"`
	PageSize int64   `json:"pageSize" example:"10"`
}

// ConfirmTitleRequest 确认/编辑标题后继续流水线
type ConfirmTitleRequest struct {
	TaskID string `json:"taskId" binding:"required"`
	Title  string `json:"title" binding:"required"`
}

// ConfirmStoryboardRequest 确认/编辑分镜（可选 HITL）
type ConfirmStoryboardRequest struct {
	TaskID     string            `json:"taskId" binding:"required"`
	Storyboard []StoryboardPanel `json:"storyboard" binding:"required"`
}

// PublishComicRequest 触发公众号发布
type PublishComicRequest struct {
	TaskID   string `json:"taskId" binding:"required"`
	Platform string `json:"platform" example:"WECHAT_MP" enums:"WECHAT_MP"`
}

// ComicPageResult 漫画分页
type ComicPageResult struct {
	Total    int64       `json:"total"`
	Records  []ComicInfo `json:"records"`
	PageNum  int64       `json:"pageNum"`
	PageSize int64       `json:"pageSize"`
}