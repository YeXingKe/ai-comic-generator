package service

import (
	"context"
	"fmt"
	"log"

	"github.com/ai-comic-generator/server/internal/agent"
	"github.com/ai-comic-generator/server/internal/agent/agents"
	"github.com/ai-comic-generator/server/internal/model"
	"github.com/ai-comic-generator/server/internal/store"
	"github.com/tmc/langchaingo/llms"
)

// ComicOrchestrator 漫画流水线编排器（标题推荐 + 六步）
type ComicOrchestrator struct {
	comicStore     *store.ComicStore
	titleAgent     agent.Agent
	storyAgent     agent.Agent
	characterAgent agent.Agent
	scriptAgent    agent.Agent
	imageService   *ImageService
	composeService *ComposeService
	publishService *PublishService
}

func NewComicOrchestrator(
	llm llms.Model,
	comicStore *store.ComicStore,
	imageSvc *ImageService,
	composeSvc *ComposeService,
	publishSvc *PublishService,
) *ComicOrchestrator {
	return &ComicOrchestrator{
		comicStore:     comicStore,
		titleAgent:     agents.NewTitleAgent(llm),
		storyAgent:     agents.NewStoryAgent(llm),
		characterAgent: agents.NewCharacterAgent(llm),
		scriptAgent:    agents.NewScriptAgent(llm),
		imageService:   imageSvc,
		composeService: composeSvc,
		publishService: publishSvc,
	}
}

// RunTitles 仅执行标题推荐，完成后等待用户确认
func (o *ComicOrchestrator) RunTitles(ctx context.Context, state *model.ComicState) error {
	log.Printf("comic step start: taskId=%s phase=%s", state.TaskID, model.ComicPhaseTitleGeneration)
	if err := o.comicStore.UpdatePhase(state.TaskID, model.ComicStatusProcessing, model.ComicPhaseTitleGeneration); err != nil {
		return err
	}
	if err := o.titleAgent.Execute(ctx, state); err != nil {
		_ = o.comicStore.MarkFailed(state.TaskID, model.ComicPhaseTitleGeneration, err.Error())
		return fmt.Errorf("%s: %w", model.ComicPhaseTitleGeneration, err)
	}
	return o.comicStore.MarkAwaitingTitleConfirm(state)
}

// RunFromStory 从故事构思起执行后续六步（用户确认标题后调用）
func (o *ComicOrchestrator) RunFromStory(ctx context.Context, state *model.ComicState) error {
	steps := []struct {
		phase string
		fn    func(context.Context, *model.ComicState) error
	}{
		{model.ComicPhaseStoryIdeation, o.storyAgent.Execute},
		{model.ComicPhaseCharacterDesign, o.characterAgent.Execute},
		{model.ComicPhaseStoryboardScript, o.scriptAgent.Execute},
		{model.ComicPhaseImageGeneration, o.imageService.GeneratePanels},
		{model.ComicPhaseLayoutCompose, o.composeService.Compose},
		{model.ComicPhaseWechatPublish, o.publishService.Publish},
	}

	for _, step := range steps {
		log.Printf("comic step start: taskId=%s phase=%s", state.TaskID, step.phase)
		if err := o.comicStore.UpdatePhase(state.TaskID, model.ComicStatusProcessing, step.phase); err != nil {
			return err
		}
		if err := step.fn(ctx, state); err != nil {
			_ = o.comicStore.MarkFailed(state.TaskID, step.phase, err.Error())
			return fmt.Errorf("%s: %w", step.phase, err)
		}
		if err := o.comicStore.SyncState(state); err != nil {
			return err
		}
	}
	return o.comicStore.MarkCompleted(state.TaskID)
}
