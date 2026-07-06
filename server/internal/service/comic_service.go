package service // 业务逻辑层：漫画任务 CRUD 与异步触发

import (
	"context" // 后台流水线 goroutine 使用的上下文
	"log"     // 记录流水线失败日志

	"github.com/ai-comic-generator/server/internal/common" // 公共常量、业务错误
	"github.com/ai-comic-generator/server/internal/model"  // 漫画请求/实体/状态模型
	"github.com/ai-comic-generator/server/internal/store" // 漫画数据访问层
	"github.com/google/uuid" // 生成任务 UUID
)

// ComicService 漫画任务业务层
type ComicService struct {
	comicStore   *store.ComicStore   // 漫画数据库操作
	orchestrator *ComicOrchestrator  // 六步流水线编排器
}

// NewComicService 创建漫画业务服务
func NewComicService(comicStore *store.ComicStore, orchestrator *ComicOrchestrator) *ComicService {
	return &ComicService{comicStore: comicStore, orchestrator: orchestrator} // 注入 Store 与编排器
}

// Create 创建漫画任务并异步启动六步流水线，返回 taskId
func (s *ComicService) Create(userID int64, req *model.CreateComicRequest) (string, error) {
	taskID := uuid.NewString() // 生成全局唯一任务 ID
	style := req.Style         // 读取用户指定的漫画风格
	if style == "" {           // 未指定风格
		style = common.ComicStyleCartoon // 默认卡通风格
	}

	comic := &model.Comic{ // 组装数据库实体
		TaskID: taskID,                   // 任务 UUID
		UserID: userID,                   // 所属用户
		Topic:  req.Topic,                 // 创作主题
		Style:  style,                    // 漫画风格
		Status: model.ComicStatusPending,  // 初始总状态：等待
		Phase:  model.ComicPhasePending,  // 初始阶段：未开始
	}
	if req.UserDescription != nil { // 用户提供了补充描述
		comic.UserDescription = req.UserDescription // 写入可选描述字段
	}
	if err := s.comicStore.Create(comic); err != nil { // 持久化到数据库
		return "", common.ErrOperation.WithMessage("创建漫画任务失败") // 包装为业务错误
	}

	state := &model.ComicState{ // 组装流水线内存态（goroutine 内逐步填充）
		TaskID: taskID,                  // 任务 UUID
		UserID: userID,                  // 所属用户
		Topic:  req.Topic,                // 创作主题
		Style:  style,                   // 漫画风格
		Phase:  model.ComicPhasePending, // 初始阶段
	}
	if req.UserDescription != nil { // 有补充描述
		state.UserDescription = *req.UserDescription // 写入内存态（非指针）
	}

	go func() { // 异步执行六步流水线，不阻塞 HTTP 响应
		ctx := context.Background()                          // 后台任务使用独立上下文
		if err := s.orchestrator.Run(ctx, state); err != nil { // 顺序执行故事→角色→分镜→生图→排版→发布
			log.Printf("comic pipeline failed taskId=%s err=%v", taskID, err) // 失败仅记日志（DB 已 MarkFailed）
		}
	}()

	return taskID, nil // 立即返回 taskId 供前端轮询
}

// GetByTaskID 按 taskId 查询漫画详情（不做权限校验，内部或管理员用）
func (s *ComicService) GetByTaskID(taskID string) (*model.ComicInfo, error) {
	comic, err := s.comicStore.GetByTaskID(taskID) // 从数据库查询
	if err != nil { // 未找到或 DB 错误
		return nil, common.ErrNotFound // 统一为未找到
	}
	return comic.ToComicInfo(), nil // 转为 API 响应结构（解析 JSON 列）
}

// ListByPage 分页查询漫画任务列表
func (s *ComicService) ListByPage(req *model.QueryComicRequest) (*model.ComicPageResult, error) {
	if req.PageNum <= 0 { // 页码非法
		req.PageNum = common.DefaultPageNum // 默认第 1 页
	}
	if req.PageSize <= 0 { // 每页条数非法
		req.PageSize = common.DefaultPageSize // 默认 10 条
	}
	if req.PageSize > common.MaxPageSize { // 超过上限
		req.PageSize = common.MaxPageSize // 限制最大 100 条
	}
	page, err := s.comicStore.ListByPage(req) // 委托 Store 分页查询
	if err != nil { // 数据库错误
		return nil, common.ErrSystem // 返回系统错误
	}
	return page, nil // 返回分页结果
}

// GetForUser 按 taskId 查询详情，非管理员只能查看自己的任务
func (s *ComicService) GetForUser(taskID string, userID int64, isAdmin bool) (*model.ComicInfo, error) {
	comic, err := s.comicStore.GetByTaskID(taskID) // 从数据库查询
	if err != nil { // 任务不存在
		return nil, common.ErrNotFound // 返回未找到
	}
	if !isAdmin && comic.UserID != userID { // 非管理员且任务不属于当前用户
		return nil, common.ErrNoAuth // 返回无权限
	}
	return comic.ToComicInfo(), nil // 转为 API 响应结构
}
