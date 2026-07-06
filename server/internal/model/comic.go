package model // 模型包：定义漫画任务相关的数据库实体、API 响应结构与编排状态

import "time" // 时间类型，用于创建/完成时间等字段

// Comic 漫画生成任务实体（对应数据库表 comic）
type Comic struct {
	ID              int64      `gorm:"primaryKey;autoIncrement" json:"id"`                                      // 主键 ID，自增
	TaskID          string     `gorm:"column:taskId;uniqueIndex:uk_taskId" json:"taskId"`                       // 任务 ID（UUID），全局唯一标识一次漫画生成
	UserID          int64      `gorm:"column:userId;index:idx_userId" json:"userId"`                              // 所属用户 ID，关联 user 表
	Topic           string     `gorm:"column:topic" json:"topic"`                                                 // 创作主题/关键词（用户输入）
	UserDescription *string    `gorm:"column:userDescription;type:text" json:"userDescription"`                   // 用户补充描述（可选，*string 表示可为 NULL）
	Title           *string    `gorm:"column:title" json:"title"`                                                 // 漫画标题（故事构思阶段确定后写入）
	CoverImage      *string    `gorm:"column:coverImage" json:"coverImage"`                                       // 封面图 URL（排版合成后写入）
	Style           string     `gorm:"column:style;default:cartoon" json:"style"`                                 // 漫画风格：cartoon / realistic / chibi 等，默认 cartoon

	// 六步产物：数据库以 JSON 字符串存储，返回 API 时由 ToComicInfo 解析为结构体
	TitleOptions   *string `gorm:"column:titleOptions;type:json" json:"titleOptions"`     // 0. 标题推荐列表（JSON）
	StoryIdeation  *string `gorm:"column:storyIdeation;type:json" json:"storyIdeation"`   // 1. 故事构思结果（JSON）
	Characters     *string `gorm:"column:characters;type:json" json:"characters"`         // 2. 角色设定列表（JSON 数组）
	Storyboard     *string `gorm:"column:storyboard;type:json" json:"storyboard"`         // 3. 分镜脚本（JSON）
	PanelImages    *string `gorm:"column:panelImages;type:json" json:"panelImages"`       // 4. 各分镜格图片（JSON 数组）
	ComposedLayout *string `gorm:"column:composedLayout;type:json" json:"composedLayout"` // 5. 排版合成结果（JSON）
	PublishResult  *string `gorm:"column:publishResult;type:json" json:"publishResult"`   // 6. 公众号发布结果（JSON）

	Status        string     `gorm:"column:status;default:PENDING;index:idx_status" json:"status"`              // 任务总状态：PENDING / PROCESSING / COMPLETED / FAILED
	Phase         string     `gorm:"column:phase;default:PENDING" json:"phase"`                                 // 当前流水线阶段（见 ComicPhase 常量）
	ErrorMessage  *string    `gorm:"column:errorMessage;type:text" json:"errorMessage"`                         // 失败时的错误信息（成功时为 nil）
	CreateTime    time.Time  `gorm:"column:createTime;autoCreateTime;index:idx_createTime" json:"createTime"`   // 任务创建时间，GORM 自动写入
	CompletedTime *time.Time `gorm:"column:completedTime" json:"completedTime"`                                 // 任务完成时间（未完成时为 nil）
	UpdateTime    time.Time  `gorm:"column:updateTime;autoUpdateTime" json:"updateTime"`                        // 最后更新时间，GORM 自动维护
	IsDelete      int        `gorm:"column:isDelete;default:0" json:"-"`                                        // 软删除：0 正常，1 已删除；json:"-" 不返回前端
}

// TableName 指定 GORM 映射的表名
func (Comic) TableName() string {
	return "comic" // 表名 comic
}

// ComicStatus 漫画任务总状态常量
const (
	ComicStatusPending         = "PENDING"          // 等待开始
	ComicStatusProcessing      = "PROCESSING"       // 生成进行中
	ComicStatusAwaitingConfirm = "AWAITING_CONFIRM" // 等待用户确认（如标题选择）
	ComicStatusTitleConfirmed  = "TITLE_CONFIRMED"  // 标题已确认，等待用户启动流水线
	ComicStatusCompleted       = "COMPLETED"        // 全部步骤完成
	ComicStatusFailed          = "FAILED"           // 某步骤失败，任务终止
)

// ComicPhase 漫画生成流水线阶段（标题推荐 + 六步 + 初始态）
const (
	ComicPhasePending          = "PENDING"           // 初始：任务已创建，尚未进入第一步
	ComicPhaseTitleGeneration  = "TITLE_GENERATION"  // 第 0 步：AI 生成标题推荐
	ComicPhaseTitleSelecting   = "TITLE_SELECTING"   // 第 0 步：等待用户选择/编辑标题
	ComicPhaseStoryIdeation    = "STORY_IDEATION"    // 第 1 步：故事构思
	ComicPhaseCharacterDesign  = "CHARACTER_DESIGN"  // 第 2 步：角色设定
	ComicPhaseStoryboardScript = "STORYBOARD_SCRIPT" // 第 3 步：分镜脚本
	ComicPhaseImageGeneration  = "IMAGE_GENERATION"  // 第 4 步：图片生成
	ComicPhaseLayoutCompose    = "LAYOUT_COMPOSE"    // 第 5 步：排版合成
	ComicPhaseWechatPublish    = "WECHAT_PUBLISH"    // 第 6 步：公众号发布
)

// ---------- 分步 JSON 结构（序列化后存入 Comic 对应 JSON 列） ----------

// TitleOption 单个标题推荐方案
type TitleOption struct {
	Title    string `json:"title"`              // 漫画标题，不超过 20 字
	Subtitle string `json:"subtitle,omitempty"` // 副标题/卖点，可选
}

// TitleOptionsResult 标题推荐阶段输出
type TitleOptionsResult struct {
	Options []TitleOption `json:"options"` // 3-5 个标题方案
}

// StoryIdeationResult 故事构思阶段的结构化输出
type StoryIdeationResult struct {
	Synopsis    string   `json:"synopsis"`    // 故事梗概（一段话摘要）
	Theme       string   `json:"theme"`       // 主题（如成长、友情）
	Tone        string   `json:"tone"`        // 基调（搞笑 / 热血 / 治愈等）
	Title       string   `json:"title"`       // AI 推荐的漫画标题
	KeyConflict string   `json:"keyConflict"` // 核心冲突（推动剧情的主矛盾）
	Highlights  []string `json:"highlights"`  // 亮点情节列表
}

// ComicCharacter 单个角色的设定信息
type ComicCharacter struct {
	Name        string `json:"name"`        // 角色名称
	Role        string `json:"role"`        // 角色类型：protagonist 主角 / antagonist 反派 / supporting 配角
	Appearance  string `json:"appearance"`  // 外貌描述（供后续生图 Prompt 使用）
	Personality string `json:"personality"` // 性格特征
	AvatarURL   string `json:"avatarUrl"`   // 角色立绘 URL（角色设定阶段生成，可为空）
}

// StoryboardPanel 分镜脚本中的单格（一格漫画）
type StoryboardPanel struct {
	PanelNo     int      `json:"panelNo"`     // 分镜格序号，从 1 开始
	Scene       string   `json:"scene"`       // 场景与画面描述
	Dialogue    []string `json:"dialogue"`    // 角色台词列表（可多气泡）
	Narration   string   `json:"narration"`   // 旁白文字
	Camera      string   `json:"camera"`      // 镜头类型（特写 / 中景 / 全景等）
	ImagePrompt string   `json:"imagePrompt"` // 该格 AI 生图的英文/中文 Prompt
}

// StoryboardResult 完整分镜脚本
type StoryboardResult struct {
	PageCount int               `json:"pageCount"` // 预计页数
	Panels    []StoryboardPanel `json:"panels"`  // 全部分镜格列表
}

// PanelImageResult 某一格生成完成后的图片结果
type PanelImageResult struct {
	PanelNo     int    `json:"panelNo"`     // 对应分镜格序号
	URL         string `json:"url"`         // 图片访问地址
	Method      string `json:"method" enums:"AI_GENERATE,UPLOAD"` // 图片来源：AI 生成 或 用户上传
	ImagePrompt string `json:"imagePrompt"` // 实际使用的生图 Prompt（便于追溯）
}

// ComposedLayoutResult 排版合成后的成品信息
type ComposedLayoutResult struct {
	Format     string   `json:"format" example:"long_image"` // 排版格式：long_image 长图 / grid 宫格 / page_pdf 分页 PDF
	PreviewURL string   `json:"previewUrl"`                  // 合成预览图 URL
	AssetURLs  []string `json:"assetUrls"`                   // 各页或各资源的最终文件 URL 列表
	CoverImage string   `json:"coverImage"`                  // 封面图 URL（用于列表展示）
}

// PublishResult 公众号发布结果
type PublishResult struct {
	Platform    string     `json:"platform" example:"WECHAT_MP"`            // 发布平台，当前为微信公众号 WECHAT_MP
	Title       string     `json:"title"`                                   // 发布标题（来自故事构思）
	MediaID     string     `json:"mediaId"`                                 // 微信素材库中的 media_id
	ArticleURL  string     `json:"articleUrl"`                              // 发布后的文章链接（若有）
	PublishedAt *time.Time `json:"publishedAt"`                             // 发布时间（未发布时为 nil）
	Status      string     `json:"status" enums:"DRAFT,PUBLISHED,FAILED"`   // 发布状态：草稿 / 已发布 / 失败
}

// ---------- API 响应（JSON 字段已解析为 Go 结构体，非字符串） ----------

// ComicInfo 漫画任务详情，供 Handler 返回给前端
type ComicInfo struct {
	ID              int64                 `json:"id"`              // 主键 ID
	TaskID          string                `json:"taskId"`          // 任务 UUID
	UserID          int64                 `json:"userId"`          // 所属用户 ID
	Topic           string                `json:"topic"`           // 创作主题
	UserDescription *string               `json:"userDescription"` // 用户补充描述
	Title           *string               `json:"title"`           // 漫画标题
	CoverImage      *string               `json:"coverImage"`      // 封面图 URL
	Style           string                `json:"style"`           // 漫画风格
	TitleOptions    *TitleOptionsResult   `json:"titleOptions"`    // 标题推荐列表（已解析）
	StoryIdeation   *StoryIdeationResult  `json:"storyIdeation"`   // 故事构思（已解析）
	Characters      []ComicCharacter      `json:"characters"`      // 角色列表（已解析）
	Storyboard      *StoryboardResult     `json:"storyboard"`      // 分镜脚本（已解析）
	PanelImages     []PanelImageResult    `json:"panelImages"`     // 分镜图片列表（已解析）
	ComposedLayout  *ComposedLayoutResult `json:"composedLayout"`  // 排版结果（已解析）
	PublishResult   *PublishResult        `json:"publishResult"`   // 发布结果（已解析）
	Status          string                `json:"status"`          // 任务总状态
	Phase           string                `json:"phase"`           // 当前阶段
	ErrorMessage    *string               `json:"errorMessage"`    // 错误信息
	CreateTime      time.Time             `json:"createTime"`      // 创建时间
	CompletedTime   *time.Time            `json:"completedTime"`   // 完成时间
}

// ToComicInfo 将数据库实体 Comic 转为 API 响应 ComicInfo，并解析 JSON 列
func (c *Comic) ToComicInfo() *ComicInfo {
	if c == nil { // 空指针保护，避免 panic
		return nil
	}
	// 先拷贝标量字段与可直接赋值的字段
	info := &ComicInfo{
		ID:              c.ID,
		TaskID:          c.TaskID,
		UserID:          c.UserID,
		Topic:           c.Topic,
		UserDescription: c.UserDescription,
		Title:           c.Title,
		CoverImage:      c.CoverImage,
		Style:           c.Style,
		Status:          c.Status,
		Phase:           c.Phase,
		ErrorMessage:    c.ErrorMessage,
		CreateTime:      c.CreateTime,
		CompletedTime:   c.CompletedTime,
	}
	// 以下将 *string JSON 列反序列化为对应结构体/切片
	if c.TitleOptions != nil {
		var v TitleOptionsResult
		parseJSON(*c.TitleOptions, &v)
		info.TitleOptions = &v
	}
	if c.StoryIdeation != nil { // 有故事构思 JSON 时才解析
		var v StoryIdeationResult
		parseJSON(*c.StoryIdeation, &v) // 调用 util.go 中的 JSON 解析
		info.StoryIdeation = &v
	}
	if c.Characters != nil { // 有角色 JSON 时才解析
		parseJSON(*c.Characters, &info.Characters)
	}
	if c.Storyboard != nil { // 有分镜 JSON 时才解析
		var v StoryboardResult
		parseJSON(*c.Storyboard, &v)
		info.Storyboard = &v
	}
	if c.PanelImages != nil { // 有图片 JSON 时才解析
		parseJSON(*c.PanelImages, &info.PanelImages)
	}
	if c.ComposedLayout != nil { // 有排版 JSON 时才解析
		var v ComposedLayoutResult
		parseJSON(*c.ComposedLayout, &v)
		info.ComposedLayout = &v
	}
	if c.PublishResult != nil { // 有发布 JSON 时才解析
		var v PublishResult
		parseJSON(*c.PublishResult, &v)
		info.PublishResult = &v
	}
	return info // 返回完整 API 视图
}

// ComicState 漫画生成过程中的内存状态，供 Chain/Service 多步编排传递（不写库，或按需同步到 Comic）
type ComicState struct {
	TaskID          string                `json:"taskId"`          // 任务 UUID
	UserID          int64                 `json:"userId"`          // 用户 ID
	Topic           string                `json:"topic"`           // 创作主题
	UserDescription string                `json:"userDescription"` // 用户描述（编排时用 string，空则为 ""）
	Style           string                `json:"style"`           // 漫画风格
	Phase           string                `json:"phase"`           // 当前执行到的阶段
	SelectedTitle   string                `json:"selectedTitle"`   // 用户确认的标题
	TitleOptions    *TitleOptionsResult   `json:"titleOptions"`    // 标题推荐（内存态）
	StoryIdeation   *StoryIdeationResult  `json:"storyIdeation"`   // 故事构思（内存态）
	Characters      []ComicCharacter      `json:"characters"`      // 角色列表（内存态）
	Storyboard      *StoryboardResult     `json:"storyboard"`      // 分镜脚本（内存态）
	PanelImages     []PanelImageResult    `json:"panelImages"`     // 分镜图片（内存态）
	ComposedLayout  *ComposedLayoutResult `json:"composedLayout"`  // 排版结果（内存态）
	PublishResult   *PublishResult        `json:"publishResult"`   // 发布结果（内存态）
}
