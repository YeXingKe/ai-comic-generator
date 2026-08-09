package service

import (
	"context"
	"fmt"
	"log"

	"github.com/ai-comic-generator/server/internal/agent"
	"github.com/ai-comic-generator/server/internal/agent/agents"
	"github.com/ai-comic-generator/server/internal/common"
	"github.com/ai-comic-generator/server/internal/model"
	"github.com/ai-comic-generator/server/internal/store"
	"github.com/tmc/langchaingo/llms"
)

// ComicOrchestrator 漫画流水线编排器（标题推荐 + 五步创作，不含公众号发布）
type ComicOrchestrator struct {
	comicStore     *store.ComicStore
	titleAgent     agent.Agent
	storyAgent     agent.Agent
	characterAgent agent.Agent
	scriptAgent    agent.Agent
	imageService   *ImageService
	composeService *ComposeService
}

func NewComicOrchestrator(
	llm llms.Model,
	promptBuilder *common.PromptBuilder,
	comicStore *store.ComicStore,
	imageSvc *ImageService,
	composeSvc *ComposeService,
) *ComicOrchestrator {
	return &ComicOrchestrator{
		comicStore:     comicStore,
		titleAgent:     agents.NewTitleAgent(llm, promptBuilder),
		storyAgent:     agents.NewStoryAgent(llm, promptBuilder),
		characterAgent: agents.NewCharacterAgent(llm, promptBuilder),
		scriptAgent:    agents.NewScriptAgent(llm, promptBuilder),
		imageService:   imageSvc,
		composeService: composeSvc,
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

// RunFromStory 从故事构思起执行后续五步（用户确认标题后调用；公众号发布由用户在历史列表手动触发）
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
