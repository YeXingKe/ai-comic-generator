package agents // 故事构思 Agent 包

import (
	"context"
	"fmt"

	"github.com/ai-comic-generator/server/internal/common"
	"github.com/ai-comic-generator/server/internal/model"
	"github.com/ai-comic-generator/server/internal/pkg/llmjson"
	"github.com/tmc/langchaingo/llms"
)

// StoryAgent 故事构思 Agent（流水线第 1 步）
type StoryAgent struct {
	llm llms.Model
}

// NewStoryAgent 创建故事构思 Agent
func NewStoryAgent(llm llms.Model) *StoryAgent {
	return &StoryAgent{llm: llm}
}

// Execute 根据用户主题生成故事构思，写入 state.StoryIdeation
func (a *StoryAgent) Execute(ctx context.Context, state *model.ComicState) error {
	prompt := common.BuildStoryIdeationPrompt(state.Topic, state.Style, state.UserDescription, state.SelectedTitle)
	content, err := llms.GenerateFromSinglePrompt(ctx, a.llm, prompt)
	if err != nil {
		return fmt.Errorf("story llm: %w", err)
	}

	var result model.StoryIdeationResult
	if err := llmjson.Unmarshal(content, &result); err != nil {
		return fmt.Errorf("parse story: %w", err)
	}
	if state.SelectedTitle != "" {
		result.Title = state.SelectedTitle
	}

	state.StoryIdeation = &result
	state.Phase = model.ComicPhaseStoryIdeation
	return nil
}
