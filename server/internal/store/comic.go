package store // 漫画任务数据访问层

import (
	"encoding/json" // 将结构体序列化为 JSON 字符串存入数据库
	"time"          // 记录任务完成时间

	"github.com/ai-comic-generator/server/internal/model" // 漫画实体与请求模型
	"gorm.io/gorm" // ORM，操作 MySQL
)

// ComicStore 漫画任务数据访问
type ComicStore struct {
	db *gorm.DB // 注入的 GORM 数据库连接
}

// NewComicStore 创建漫画 Store
func NewComicStore(db *gorm.DB) *ComicStore {
	return &ComicStore{db: db} // 保存数据库连接引用
}

// Create 插入一条新的漫画任务记录
func (s *ComicStore) Create(comic *model.Comic) error {
	return s.db.Create(comic).Error // GORM 插入，失败返回底层错误
}

// GetByTaskID 按任务 UUID 查询漫画（排除软删除）
func (s *ComicStore) GetByTaskID(taskID string) (*model.Comic, error) {
	var comic model.Comic                                                                      // 声明接收查询结果的实体
	err := s.db.Where("taskId = ? AND isDelete = 0", taskID).First(&comic).Error // 按 taskId 查首条未删除记录
	if err != nil { // 未找到或数据库错误
		return nil, err // 返回 nil 与错误
	}
	return &comic, nil // 返回查询到的漫画指针
}

// UpdatePhase 更新任务状态与当前流水线阶段（步骤开始前调用）
func (s *ComicStore) UpdatePhase(taskID, status, phase string) error {
	return s.db.Model(&model.Comic{}). // 指定更新 comic 表
						Where("taskId = ? AND isDelete = 0", taskID). // 限定目标任务
						Updates(map[string]interface{}{ // 批量更新字段
			"status": status, // 任务总状态（如 PROCESSING）
			"phase":  phase,  // 当前步骤阶段（如 STORY_IDEATION）
		}).Error // 返回更新错误（若有）
}

// SyncState 将流水线内存态中的各步产物同步到数据库 JSON 列
func (s *ComicStore) SyncState(state *model.ComicState) error {
	updates := map[string]interface{}{ // 待更新字段集合
		"phase": state.Phase, // 始终同步当前阶段
	}
	if state.TitleOptions != nil {
		updates["titleOptions"] = toJSON(state.TitleOptions)
	}
	if state.SelectedTitle != "" {
		updates["title"] = state.SelectedTitle
	}
	if state.StoryIdeation != nil { // 故事构思步骤已完成
		j := toJSON(state.StoryIdeation) // 序列化为 JSON 字符串
		updates["storyIdeation"] = j     // 写入 storyIdeation 列
		if state.StoryIdeation.Title != "" { // AI 生成了标题
			updates["title"] = state.StoryIdeation.Title // 同步到 title 列便于列表展示
		}
	}
	if len(state.Characters) > 0 { // 角色设定步骤已完成
		updates["characters"] = toJSON(state.Characters) // 写入 characters JSON 列
	}
	if state.Storyboard != nil { // 分镜脚本步骤已完成
		updates["storyboard"] = toJSON(state.Storyboard) // 写入 storyboard JSON 列
	}
	if len(state.PanelImages) > 0 { // 画面生成步骤已完成
		updates["panelImages"] = toJSON(state.PanelImages) // 写入 panelImages JSON 列
	}
	if state.ComposedLayout != nil { // 排版合成步骤已完成
		updates["composedLayout"] = toJSON(state.ComposedLayout) // 写入 composedLayout JSON 列
		if state.ComposedLayout.CoverImage != "" { // 有封面图 URL
			updates["coverImage"] = state.ComposedLayout.CoverImage // 同步到 coverImage 列
		}
	}
	if state.PublishResult != nil { // 公众号发布步骤已完成
		updates["publishResult"] = toJSON(state.PublishResult) // 写入 publishResult JSON 列
	}
	return s.db.Model(&model.Comic{}).Where("taskId = ?", state.TaskID).Updates(updates).Error
}

// MarkAwaitingTitleConfirm 标题推荐完成，等待用户选择
func (s *ComicStore) MarkAwaitingTitleConfirm(state *model.ComicState) error {
	updates := map[string]interface{}{
		"status": model.ComicStatusAwaitingConfirm,
		"phase":  model.ComicPhaseTitleSelecting,
	}
	if state.TitleOptions != nil {
		updates["titleOptions"] = toJSON(state.TitleOptions)
	}
	return s.db.Model(&model.Comic{}).Where("taskId = ?", state.TaskID).Updates(updates).Error
}

// BuildStateFromComic 从数据库实体恢复流水线内存态
func (s *ComicStore) BuildStateFromComic(c *model.Comic) *model.ComicState {
	state := &model.ComicState{
		TaskID: c.TaskID,
		UserID: c.UserID,
		Topic:  c.Topic,
		Style:  c.Style,
		Phase:  c.Phase,
	}
	if c.UserDescription != nil {
		state.UserDescription = *c.UserDescription
	}
	if c.Title != nil {
		state.SelectedTitle = *c.Title
	}
	if c.TitleOptions != nil {
		var v model.TitleOptionsResult
		parseJSONInto(*c.TitleOptions, &v)
		state.TitleOptions = &v
	}
	if c.StoryIdeation != nil {
		var v model.StoryIdeationResult
		parseJSONInto(*c.StoryIdeation, &v)
		state.StoryIdeation = &v
	}
	if c.Characters != nil {
		parseJSONInto(*c.Characters, &state.Characters)
	}
	if c.Storyboard != nil {
		var v model.StoryboardResult
		parseJSONInto(*c.Storyboard, &v)
		state.Storyboard = &v
	}
	if c.PanelImages != nil {
		parseJSONInto(*c.PanelImages, &state.PanelImages)
	}
	if c.ComposedLayout != nil {
		var v model.ComposedLayoutResult
		parseJSONInto(*c.ComposedLayout, &v)
		state.ComposedLayout = &v
	}
	if c.PublishResult != nil {
		var v model.PublishResult
		parseJSONInto(*c.PublishResult, &v)
		state.PublishResult = &v
	}
	return state
}

func parseJSONInto(raw string, target interface{}) {
	_ = json.Unmarshal([]byte(raw), target)
}

func (s *ComicStore) MarkFailed(taskID, phase, errMsg string) error {
	return s.db.Model(&model.Comic{}). // 指定更新 comic 表
						Where("taskId = ? AND isDelete = 0", taskID). // 限定目标任务
						Updates(map[string]interface{}{ // 批量更新失败相关字段
			"status":       model.ComicStatusFailed, // 总状态改为 FAILED
			"phase":        phase,                 // 记录失败时所在阶段
			"errorMessage": errMsg,                // 记录错误详情供前端展示
		}).Error // 返回更新错误（若有）
}

// MarkCompleted 标记任务全部步骤完成
func (s *ComicStore) MarkCompleted(taskID string) error {
	now := time.Now() // 取当前时间作为完成时间
	return s.db.Model(&model.Comic{}). // 指定更新 comic 表
						Where("taskId = ? AND isDelete = 0", taskID). // 限定目标任务
						Updates(map[string]interface{}{ // 批量更新完成相关字段
			"status":        model.ComicStatusCompleted,    // 总状态改为 COMPLETED
			"phase":         model.ComicPhaseWechatPublish, // 最终阶段为公众号发布
			"completedTime": &now,                          // 写入完成时间戳
		}).Error // 返回更新错误（若有）
}

// ListByPage 分页查询漫画任务列表
func (s *ComicStore) ListByPage(req *model.QueryComicRequest) (*model.ComicPageResult, error) {
	q := s.db.Model(&model.Comic{}).Where("isDelete = 0") // 基础查询：未软删除
	if req.UserID != nil { // 按用户 ID 筛选（普通用户只看自己的）
		q = q.Where("userId = ?", *req.UserID) // 追加 userId 条件
	}
	if req.Status != nil && *req.Status != "" { // 按任务状态筛选
		q = q.Where("status = ?", *req.Status) // 追加 status 条件
	}
	if req.Phase != nil && *req.Phase != "" { // 按流水线阶段筛选
		q = q.Where("phase = ?", *req.Phase) // 追加 phase 条件
	}

	var total int64                             // 符合条件的总记录数
	if err := q.Count(&total).Error; err != nil { // 统计总数
		return nil, err // 统计失败则返回错误
	}

	pageNum, pageSize := req.PageNum, req.PageSize // 读取请求中的分页参数
	if pageNum <= 0 { // 页码非法时使用默认值
		pageNum = 1 // 默认第 1 页
	}
	if pageSize <= 0 { // 每页条数非法时使用默认值
		pageSize = 10 // 默认每页 10 条
	}

	var rows []model.Comic // 当前页的数据行
	err := q.Order("createTime DESC"). // 按创建时间倒序
						Offset(int((pageNum - 1) * pageSize)). // 跳过前面页的数据
						Limit(int(pageSize)).                  // 限制本页条数
						Find(&rows).Error                      // 执行查询
	if err != nil { // 查询失败
		return nil, err // 返回错误
	}

	records := make([]model.ComicInfo, 0, len(rows)) // 预分配 API 响应切片
	for i := range rows { // 遍历查询结果
		if info := rows[i].ToComicInfo(); info != nil { // 实体转 API 结构（含 JSON 解析）
			records = append(records, *info) // 追加到结果列表
		}
	}
	return &model.ComicPageResult{ // 组装分页响应
		Total: total, Records: records, PageNum: pageNum, PageSize: pageSize, // 总数、记录、页码、页大小
	}, nil // 成功返回
}

// toJSON 将任意结构体序列化为 JSON 字符串（失败返回空串）
func toJSON(v interface{}) string {
	b, err := json.Marshal(v) // 序列化
	if err != nil { // 序列化失败
		return "" // 返回空字符串，避免写入非法 JSON
	}
	return string(b) // 转为 string 供 GORM 写入 JSON 列
}
