package model

// CreateComicRequest 创建漫画任务
type CreateComicRequest struct {
	Topic           string  `json:"topic" binding:"required" example:"哪吒闹海"`
	UserDescription *string `json:"userDescription" example:"搞笑风格，保持角色一致性"`
	Style           string  `json:"style" example:"cartoon" enums:"cartoon,realistic,chibi,animal"`
	ImageBackend    string  `json:"imageBackend" example:"hunyuan" enums:"hunyuan,openai_image_1k,openai_image_4k"`
	CaptionTextMode string  `json:"captionTextMode" example:"top" enums:"none,top,bubble"`
	PanelCount      int     `json:"panelCount" example:"4"` // 分镜格数：4 / 6 / 8
}

// QueryComicRequest 查询漫画任务
type QueryComicRequest struct {
	UserID   *int64  `json:"userId"`
	Status   *string `json:"status" enums:"PENDING,PROCESSING,AWAITING_CONFIRM,TITLE_CONFIRMED,AWAITING_STORYBOARD,COMPLETED,FAILED"`
	Phase    *string `json:"phase"`
	PageNum  int64   `json:"pageNum" example:"1"`
	PageSize int64   `json:"pageSize" example:"10"`
}

// ConfirmTitleRequest 确认/编辑标题（不启动流水线）
type ConfirmTitleRequest struct {
	TaskID string `json:"taskId" binding:"required"`
	Title  string `json:"title" binding:"required"`
}

// StartComicRequest 正式启动后续五步流水线
type StartComicRequest struct {
	TaskID string `json:"taskId" binding:"required"`
}

// ConfirmStoryboardRequest 确认/编辑分镜后启动生图与排版
type ConfirmStoryboardRequest struct {
	TaskID     string            `json:"taskId" binding:"required"`
	Storyboard []StoryboardPanel `json:"storyboard" binding:"required"`
}

// RetryComicRequest 失败任务从当前步骤重试
type RetryComicRequest struct {
	TaskID string `json:"taskId" binding:"required"`
}

// RegeneratePanelRequest 修改单格台词/场景后重新生图（可顺带重排版）
type RegeneratePanelRequest struct {
	TaskID    string   `json:"taskId" binding:"required"`
	PanelNo   int      `json:"panelNo" binding:"required"`
	Scene     *string  `json:"scene"`
	Dialogue  []string `json:"dialogue"`
	Narration *string  `json:"narration"`
}

// PublishComicRequest 在历史列表中手动触发公众号发布
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
