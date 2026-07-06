package service

import (
	"context"
	"log"
	"strings"

	"github.com/ai-comic-generator/server/internal/common"
	"github.com/ai-comic-generator/server/internal/model"
	"github.com/ai-comic-generator/server/internal/store"
	"github.com/google/uuid"
)

// ComicService 漫画任务业务层
type ComicService struct {
	comicStore   *store.ComicStore
	orchestrator *ComicOrchestrator
}

func NewComicService(comicStore *store.ComicStore, orchestrator *ComicOrchestrator) *ComicService {
	return &ComicService{comicStore: comicStore, orchestrator: orchestrator}
}

// Create 创建漫画任务，异步生成标题推荐后等待用户确认
func (s *ComicService) Create(userID int64, req *model.CreateComicRequest) (string, error) {
	taskID := uuid.NewString()
	style := req.Style
	if style == "" {
		style = common.ComicStyleCartoon
	}

	comic := &model.Comic{
		TaskID: taskID,
		UserID: userID,
		Topic:  req.Topic,
		Style:  style,
		Status: model.ComicStatusPending,
		Phase:  model.ComicPhasePending,
	}
	if req.UserDescription != nil {
		comic.UserDescription = req.UserDescription
	}
	if err := s.comicStore.Create(comic); err != nil {
		return "", common.ErrOperation.WithMessage("创建漫画任务失败")
	}

	state := &model.ComicState{
		TaskID: taskID,
		UserID: userID,
		Topic:  req.Topic,
		Style:  style,
		Phase:  model.ComicPhasePending,
	}
	if req.UserDescription != nil {
		state.UserDescription = *req.UserDescription
	}

	go func() {
		ctx := context.Background()
		if err := s.orchestrator.RunTitles(ctx, state); err != nil {
			log.Printf("comic title generation failed taskId=%s err=%v", taskID, err)
		}
	}()

	return taskID, nil
}

// ConfirmTitle 用户确认/编辑标题后，继续执行后续六步流水线
func (s *ComicService) ConfirmTitle(userID int64, req *model.ConfirmTitleRequest, isAdmin bool) error {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return common.ErrParams.WithMessage("标题不能为空")
	}

	comic, err := s.comicStore.GetByTaskID(req.TaskID)
	if err != nil {
		return common.ErrNotFound
	}
	if !isAdmin && comic.UserID != userID {
		return common.ErrNoAuth
	}
	if comic.Status != model.ComicStatusAwaitingConfirm || comic.Phase != model.ComicPhaseTitleSelecting {
		return common.ErrOperation.WithMessage("当前任务不在标题选择阶段")
	}

	state := s.comicStore.BuildStateFromComic(comic)
	state.SelectedTitle = title

	if err := s.comicStore.UpdatePhase(req.TaskID, model.ComicStatusProcessing, model.ComicPhaseStoryIdeation); err != nil {
		return common.ErrSystem
	}
	_ = s.comicStore.SyncState(state)

	go func() {
		ctx := context.Background()
		if err := s.orchestrator.RunFromStory(ctx, state); err != nil {
			log.Printf("comic pipeline failed taskId=%s err=%v", req.TaskID, err)
		}
	}()

	return nil
}

func (s *ComicService) GetByTaskID(taskID string) (*model.ComicInfo, error) {
	comic, err := s.comicStore.GetByTaskID(taskID)
	if err != nil {
		return nil, common.ErrNotFound
	}
	return comic.ToComicInfo(), nil
}

func (s *ComicService) ListByPage(req *model.QueryComicRequest) (*model.ComicPageResult, error) {
	if req.PageNum <= 0 {
		req.PageNum = common.DefaultPageNum
	}
	if req.PageSize <= 0 {
		req.PageSize = common.DefaultPageSize
	}
	if req.PageSize > common.MaxPageSize {
		req.PageSize = common.MaxPageSize
	}
	page, err := s.comicStore.ListByPage(req)
	if err != nil {
		return nil, common.ErrSystem
	}
	return page, nil
}

func (s *ComicService) GetForUser(taskID string, userID int64, isAdmin bool) (*model.ComicInfo, error) {
	comic, err := s.comicStore.GetByTaskID(taskID)
	if err != nil {
		return nil, common.ErrNotFound
	}
	if !isAdmin && comic.UserID != userID {
		return nil, common.ErrNoAuth
	}
	return comic.ToComicInfo(), nil
}
