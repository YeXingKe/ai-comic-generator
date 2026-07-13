package agents // 分镜脚本 Agent 包

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ai-comic-generator/server/internal/common"
	"github.com/ai-comic-generator/server/internal/model"
	"github.com/ai-comic-generator/server/internal/pkg/llmjson"
	"github.com/tmc/langchaingo/llms"
)

const defaultPanelCount = 4

// ScriptAgent 分镜脚本 Agent（流水线第 3 步，qwen-plus）
type ScriptAgent struct {
	llm llms.Model
}

// NewScriptAgent 创建分镜脚本 Agent
func NewScriptAgent(llm llms.Model) *ScriptAgent {
	return &ScriptAgent{llm: llm}
}

// Execute 根据故事与角色生成分镜脚本，写入 state.Storyboard
func (a *ScriptAgent) Execute(ctx context.Context, state *model.ComicState) error {
	if state.StoryIdeation == nil {
		return fmt.Errorf("story ideation is empty")
	}
	if len(state.Characters) == 0 {
		return fmt.Errorf("characters empty")
	}

	storyJSON, _ := json.Marshal(state.StoryIdeation)
	charJSON, _ := json.Marshal(state.Characters)

	prompt := common.BuildStoryboardScriptPrompt(
		string(storyJSON), string(charJSON), state.Style, state.UserDescription, defaultPanelCount,
	)
	content, err := llms.GenerateFromSinglePrompt(ctx, a.llm, prompt)
	if err != nil {
		return fmt.Errorf("script llm: %w", err)
	}

	var storyboard model.StoryboardResult
	if err := llmjson.Unmarshal(content, &storyboard); err != nil {
		return fmt.Errorf("parse storyboard: %w", err)
	}
	if len(storyboard.Panels) == 0 {
		return fmt.Errorf("storyboard panels empty")
	}

	state.Storyboard = &storyboard
	state.Phase = model.ComicPhaseStoryboardScript
	return nil
}
