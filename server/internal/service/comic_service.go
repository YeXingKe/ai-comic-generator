package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/ai-comic-generator/server/internal/common"
	"github.com/ai-comic-generator/server/internal/model"
	"github.com/ai-comic-generator/server/internal/store"
	"github.com/google/uuid"
)

// ComicService 漫画任务业务层
type ComicService struct {
	comicStore     *store.ComicStore
	userStore      *store.UserStore
	orchestrator   *ComicOrchestrator
	publishService *PublishService
}

func NewComicService(comicStore *store.ComicStore, userStore *store.UserStore, orchestrator *ComicOrchestrator, publishSvc *PublishService) *ComicService {
	return &ComicService{comicStore: comicStore, userStore: userStore, orchestrator: orchestrator, publishService: publishSvc}
}

// Create 创建漫画任务，异步生成标题推荐后等待用户确认
func (s *ComicService) Create(userID int64, req *model.CreateComicRequest) (string, error) {
	taskID := uuid.NewString()
	style := req.Style
	if style == "" {
		style = common.ComicStyleCartoon
	}
	imageBackend := req.ImageBackend
	if imageBackend == "" {
		imageBackend = common.ImageBackendHunyuan
	}
	captionTextMode := req.CaptionTextMode
	if captionTextMode == "" {
		captionTextMode = common.CaptionTextModeTop
	}
	panelCount := req.PanelCount
	if panelCount <= 0 {
		panelCount = 4
	}
	if panelCount != 4 && panelCount != 6 && panelCount != 8 {
		return "", common.ErrParams.WithMessage("格数仅支持 4 / 6 / 8")
	}

	comic := &model.Comic{
		TaskID:          taskID,
		UserID:          userID,
		Topic:           req.Topic,
		Style:           style,
		ImageBackend:    imageBackend,
		CaptionTextMode: captionTextMode,
		PanelCount:      panelCount,
		Status:          model.ComicStatusPending,
		Phase:           model.ComicPhasePending,
	}
	if req.UserDescription != nil {
		comic.UserDescription = req.UserDescription
	}
	if err := s.comicStore.Create(comic); err != nil {
		return "", common.ErrOperation.WithMessage("创建漫画任务失败")
	}

	state := &model.ComicState{
		TaskID:          taskID,
		UserID:          userID,
		Topic:           req.Topic,
		Style:           style,
		ImageBackend:    imageBackend,
		CaptionTextMode: captionTextMode,
		PanelCount:      panelCount,
		Phase:           model.ComicPhasePending,
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

// ConfirmTitle 用户确认/编辑标题，不启动后续流水线
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

	return s.comicStore.MarkTitleConfirmed(state)
}

// StartPipeline 正式启动故事构思起的五步流水线
func (s *ComicService) StartPipeline(userID int64, req *model.StartComicRequest, isAdmin bool) error {
	comic, err := s.comicStore.GetByTaskID(req.TaskID)
	if err != nil {
		return common.ErrNotFound
	}
	if !isAdmin && comic.UserID != userID {
		return common.ErrNoAuth
	}
	if comic.Status != model.ComicStatusTitleConfirmed {
		return common.ErrOperation.WithMessage("请先确认标题后再开始生成")
	}
	if comic.Title == nil || strings.TrimSpace(*comic.Title) == "" {
		return common.ErrOperation.WithMessage("标题未设置，无法开始生成")
	}

	state := s.comicStore.BuildStateFromComic(comic)

	if err := s.comicStore.UpdatePhase(req.TaskID, model.ComicStatusProcessing, model.ComicPhaseStoryIdeation); err != nil {
		return common.ErrSystem
	}

	go func() {
		ctx := context.Background()
		if err := s.orchestrator.RunFromStory(ctx, state); err != nil {
			log.Printf("comic pipeline failed taskId=%s err=%v", req.TaskID, err)
		}
	}()

	return nil
}

// ConfirmStoryboard 确认/编辑分镜后启动生图与排版
func (s *ComicService) ConfirmStoryboard(userID int64, req *model.ConfirmStoryboardRequest, isAdmin bool) error {
	comic, err := s.comicStore.GetByTaskID(req.TaskID)
	if err != nil {
		return common.ErrNotFound
	}
	if !isAdmin && comic.UserID != userID {
		return common.ErrNoAuth
	}
	if comic.Status != model.ComicStatusAwaitingStoryboard {
		return common.ErrOperation.WithMessage("当前任务不在分镜确认阶段")
	}
	if len(req.Storyboard) == 0 {
		return common.ErrParams.WithMessage("分镜不能为空")
	}

	state := s.comicStore.BuildStateFromComic(comic)
	panels := make([]model.StoryboardPanel, len(req.Storyboard))
	copy(panels, req.Storyboard)
	for i := range panels {
		if panels[i].PanelNo <= 0 {
			panels[i].PanelNo = i + 1
		}
		panels[i].Scene = strings.TrimSpace(panels[i].Scene)
		panels[i].Narration = strings.TrimSpace(panels[i].Narration)
		if panels[i].Scene == "" {
			return common.ErrParams.WithMessage(fmt.Sprintf("第 %d 格场景描述不能为空", panels[i].PanelNo))
		}
	}
	pageCount := 1
	if state.Storyboard != nil && state.Storyboard.PageCount > 0 {
		pageCount = state.Storyboard.PageCount
	}
	state.Storyboard = &model.StoryboardResult{PageCount: pageCount, Panels: panels}
	if err := s.comicStore.SyncState(state); err != nil {
		return common.ErrSystem
	}
	if err := s.comicStore.UpdatePhase(req.TaskID, model.ComicStatusProcessing, model.ComicPhaseImageGeneration); err != nil {
		return common.ErrSystem
	}

	go func() {
		ctx := context.Background()
		if err := s.orchestrator.RunFromImages(ctx, state); err != nil {
			log.Printf("comic image pipeline failed taskId=%s err=%v", req.TaskID, err)
		}
	}()
	return nil
}

// RetryFailed 失败任务从当前步骤重试
func (s *ComicService) RetryFailed(userID int64, req *model.RetryComicRequest, isAdmin bool) error {
	comic, err := s.comicStore.GetByTaskID(req.TaskID)
	if err != nil {
		return common.ErrNotFound
	}
	if !isAdmin && comic.UserID != userID {
		return common.ErrNoAuth
	}
	if comic.Status != model.ComicStatusFailed {
		return common.ErrOperation.WithMessage("仅失败任务可重试")
	}

	fromPhase := comic.Phase
	if fromPhase == "" || fromPhase == model.ComicPhasePending {
		fromPhase = model.ComicPhaseTitleGeneration
	}

	state := s.comicStore.BuildStateFromComic(comic)
	if err := s.comicStore.ClearFailure(req.TaskID, model.ComicStatusProcessing, fromPhase); err != nil {
		return common.ErrSystem
	}

	go func() {
		ctx := context.Background()
		if err := s.orchestrator.RetryFromPhase(ctx, state, fromPhase); err != nil {
			log.Printf("comic retry failed taskId=%s phase=%s err=%v", req.TaskID, fromPhase, err)
		}
	}()
	return nil
}

// RegeneratePanel 修改单格场景/台词后重新生图并重排版
func (s *ComicService) RegeneratePanel(userID int64, req *model.RegeneratePanelRequest, isAdmin bool) (*model.ComicInfo, error) {
	comic, err := s.comicStore.GetByTaskID(req.TaskID)
	if err != nil {
		return nil, common.ErrNotFound
	}
	if !isAdmin && comic.UserID != userID {
		return nil, common.ErrNoAuth
	}
	if comic.Status != model.ComicStatusCompleted && comic.Status != model.ComicStatusAwaitingStoryboard {
		return nil, common.ErrOperation.WithMessage("当前状态不支持单格重绘")
	}
	if comic.Status == model.ComicStatusAwaitingStoryboard {
		return nil, common.ErrOperation.WithMessage("请先在分镜确认中编辑并确认分镜")
	}

	state := s.comicStore.BuildStateFromComic(comic)
	if state.Storyboard == nil || len(state.Storyboard.Panels) == 0 {
		return nil, common.ErrOperation.WithMessage("分镜脚本不存在")
	}

	found := false
	for i := range state.Storyboard.Panels {
		p := &state.Storyboard.Panels[i]
		if p.PanelNo != req.PanelNo {
			continue
		}
		found = true
		if req.Scene != nil {
			p.Scene = strings.TrimSpace(*req.Scene)
		}
		if req.Dialogue != nil {
			p.Dialogue = req.Dialogue
		}
		if req.Narration != nil {
			p.Narration = strings.TrimSpace(*req.Narration)
		}
		if p.Scene == "" {
			return nil, common.ErrParams.WithMessage("场景描述不能为空")
		}
		break
	}
	if !found {
		return nil, common.ErrParams.WithMessage("分镜格不存在")
	}

	ctx := context.Background()
	if err := s.orchestrator.imageService.GenerateSinglePanel(ctx, state, req.PanelNo); err != nil {
		return nil, common.ErrOperation.WithMessage("单格重绘失败")
	}
	if err := s.orchestrator.composeService.Compose(ctx, state); err != nil {
		return nil, common.ErrOperation.WithMessage("排版合成失败")
	}
	if err := s.comicStore.SyncState(state); err != nil {
		return nil, common.ErrSystem
	}
	_ = s.comicStore.MarkCompleted(req.TaskID)

	return s.GetForUser(req.TaskID, userID, isAdmin)
}

// Publish 将已完成作品发布至微信公众号（历史列表手动触发）
func (s *ComicService) Publish(userID int64, req *model.PublishComicRequest, isAdmin bool) (*model.PublishResult, error) {
	comic, err := s.comicStore.GetByTaskID(req.TaskID)
	if err != nil {
		return nil, common.ErrNotFound
	}
	if !isAdmin && comic.UserID != userID {
		return nil, common.ErrNoAuth
	}
	if comic.Status != model.ComicStatusCompleted {
		return nil, common.ErrOperation.WithMessage("仅已完成的作品可发布")
	}

	state := s.comicStore.BuildStateFromComic(comic)
	if state.ComposedLayout == nil {
		return nil, common.ErrOperation.WithMessage("作品尚未完成排版合成，无法发布")
	}

	ctx := context.Background()
	pubErr := s.publishService.Publish(ctx, state)
	if state.PublishResult != nil {
		if saveErr := s.comicStore.SavePublishResult(req.TaskID, state.PublishResult); saveErr != nil {
			log.Printf("comic publish result save failed taskId=%s err=%v", req.TaskID, saveErr)
			return nil, common.ErrSystem
		}
	}
	if pubErr != nil {
		log.Printf("comic publish failed taskId=%s err=%v", req.TaskID, pubErr)
		return state.PublishResult, common.ErrOperation.WithMessage("发布失败，请稍后重试")
	}
	return state.PublishResult, nil
}

func (s *ComicService) GetByTaskID(taskID string) (*model.ComicInfo, error) {
	comic, err := s.comicStore.GetByTaskID(taskID)
	if err != nil {
		return nil, common.ErrNotFound
	}
	info := comic.ToComicInfo()
	s.attachUsers([]*model.ComicInfo{info})
	return info, nil
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
	infos := make([]*model.ComicInfo, 0, len(page.Records))
	for i := range page.Records {
		infos = append(infos, &page.Records[i])
	}
	s.attachUsers(infos)
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
	info := comic.ToComicInfo()
	s.attachUsers([]*model.ComicInfo{info})
	return info, nil
}

// attachUsers 批量联查创建者账号/昵称并回填到 ComicInfo（不落库）
func (s *ComicService) attachUsers(infos []*model.ComicInfo) {
	if len(infos) == 0 || s.userStore == nil {
		return
	}
	idSet := make(map[int64]struct{}, len(infos))
	ids := make([]int64, 0, len(infos))
	for _, info := range infos {
		if info == nil || info.UserID <= 0 {
			continue
		}
		if _, ok := idSet[info.UserID]; ok {
			continue
		}
		idSet[info.UserID] = struct{}{}
		ids = append(ids, info.UserID)
	}
	if len(ids) == 0 {
		return
	}
	users, err := s.userStore.ListByIDs(ids)
	if err != nil {
		log.Printf("comic attachUsers: list users failed: %v", err)
		return
	}
	byID := make(map[int64]*model.User, len(users))
	for i := range users {
		byID[users[i].ID] = &users[i]
	}
	for _, info := range infos {
		if info == nil {
			continue
		}
		u, ok := byID[info.UserID]
		if !ok {
			continue
		}
		info.UserAccount = u.UserAccount
		info.UserName = u.UserName
	}
}
