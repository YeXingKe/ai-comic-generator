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

type pipelineStep struct {
	phase string
	fn    func(context.Context, *model.ComicState) error
}

// ComicOrchestrator 漫画流水线编排器（标题推荐 + 创作步骤，不含公众号发布）
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

func (o *ComicOrchestrator) storySteps() []pipelineStep {
	return []pipelineStep{
		{model.ComicPhaseStoryIdeation, o.storyAgent.Execute},
		{model.ComicPhaseCharacterDesign, o.executeCharacterWithLookbook},
		{model.ComicPhaseStoryboardScript, o.scriptAgent.Execute},
	}
}

func (o *ComicOrchestrator) executeCharacterWithLookbook(ctx context.Context, state *model.ComicState) error {
	if err := o.characterAgent.Execute(ctx, state); err != nil {
		return err
	}
	common.NormalizeCharacterAnchors(state.Characters)
	// 定妆照失败不阻断流水线：保留文字锚点，继续分镜
	if err := o.imageService.GenerateCharacterAvatars(ctx, state); err != nil {
		log.Printf("character lookbook failed taskId=%s err=%v (continue with text anchors)", state.TaskID, err)
	}
	return nil
}

func (o *ComicOrchestrator) imageSteps() []pipelineStep {
	return []pipelineStep{
		{model.ComicPhaseImageGeneration, o.imageService.GeneratePanels},
		{model.ComicPhaseLayoutCompose, o.composeService.Compose},
	}
}

func (o *ComicOrchestrator) runSteps(ctx context.Context, state *model.ComicState, steps []pipelineStep) error {
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
	return nil
}

func stepsFrom(steps []pipelineStep, fromPhase string) []pipelineStep {
	if fromPhase == "" {
		return steps
	}
	for i, step := range steps {
		if step.phase == fromPhase {
			return steps[i:]
		}
	}
	return steps
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

// RunFromStory 故事构思 → 角色 → 分镜，完成后等待用户确认分镜
func (o *ComicOrchestrator) RunFromStory(ctx context.Context, state *model.ComicState) error {
	if err := o.runSteps(ctx, state, o.storySteps()); err != nil {
		return err
	}
	return o.comicStore.MarkAwaitingStoryboardConfirm(state)
}

// RunFromImages 画面生成 → 排版合成 → 完成
func (o *ComicOrchestrator) RunFromImages(ctx context.Context, state *model.ComicState) error {
	if err := o.runSteps(ctx, state, o.imageSteps()); err != nil {
		return err
	}
	return o.comicStore.MarkCompleted(state.TaskID)
}

// RetryFromPhase 从失败阶段继续（标题 / 故事段 / 生图段）
func (o *ComicOrchestrator) RetryFromPhase(ctx context.Context, state *model.ComicState, fromPhase string) error {
	switch fromPhase {
	case model.ComicPhaseTitleGeneration, model.ComicPhaseTitleSelecting, model.ComicPhasePending:
		return o.RunTitles(ctx, state)
	case model.ComicPhaseStoryIdeation, model.ComicPhaseCharacterDesign, model.ComicPhaseStoryboardScript:
		if err := o.runSteps(ctx, state, stepsFrom(o.storySteps(), fromPhase)); err != nil {
			return err
		}
		return o.comicStore.MarkAwaitingStoryboardConfirm(state)
	case model.ComicPhaseImageGeneration, model.ComicPhaseLayoutCompose:
		if err := o.runSteps(ctx, state, stepsFrom(o.imageSteps(), fromPhase)); err != nil {
			return err
		}
		return o.comicStore.MarkCompleted(state.TaskID)
	default:
		return fmt.Errorf("unsupported retry phase: %s", fromPhase)
	}
}
