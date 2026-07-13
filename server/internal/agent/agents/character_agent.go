package agents // 角色设定 Agent 包

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ai-comic-generator/server/internal/common"
	"github.com/ai-comic-generator/server/internal/model"
	"github.com/ai-comic-generator/server/internal/pkg/llmjson"
	"github.com/tmc/langchaingo/llms"
)

// CharacterAgent 角色设定 Agent（流水线第 2 步，qwen-plus）
type CharacterAgent struct {
	llm llms.Model
}

// NewCharacterAgent 创建角色设定 Agent
func NewCharacterAgent(llm llms.Model) *CharacterAgent {
	return &CharacterAgent{llm: llm}
}

// Execute 根据故事构思生成角色列表，写入 state.Characters
func (a *CharacterAgent) Execute(ctx context.Context, state *model.ComicState) error {
	if state.StoryIdeation == nil {
		return fmt.Errorf("story ideation is empty")
	}
	storyJSON, err := json.Marshal(state.StoryIdeation)
	if err != nil {
		return fmt.Errorf("marshal story: %w", err)
	}

	prompt := common.BuildCharacterDesignPrompt(string(storyJSON), state.Style)
	content, err := llms.GenerateFromSinglePrompt(ctx, a.llm, prompt)
	if err != nil {
		return fmt.Errorf("character llm: %w", err)
	}

	var characters []model.ComicCharacter
	if err := llmjson.Unmarshal(content, &characters); err != nil {
		return fmt.Errorf("parse characters: %w", err)
	}
	if len(characters) == 0 {
		return fmt.Errorf("characters empty")
	}

	state.Characters = characters
	state.Phase = model.ComicPhaseCharacterDesign
	return nil
}
