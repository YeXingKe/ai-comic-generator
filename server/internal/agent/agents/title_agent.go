package agents

import (
	"context"
	"fmt"

	"github.com/ai-comic-generator/server/internal/common"
	"github.com/ai-comic-generator/server/internal/model"
	"github.com/ai-comic-generator/server/internal/pkg/llmjson"
	"github.com/tmc/langchaingo/llms"
)

// TitleAgent 标题推荐 Agent（流水线第 0 步，qwen-plus）
type TitleAgent struct {
	llm llms.Model
}

func NewTitleAgent(llm llms.Model) *TitleAgent {
	return &TitleAgent{llm: llm}
}

func (a *TitleAgent) Execute(ctx context.Context, state *model.ComicState) error {
	prompt := common.BuildTitleIdeationPrompt(state.Topic, state.Style, state.UserDescription)
	content, err := llms.GenerateFromSinglePrompt(ctx, a.llm, prompt)
	if err != nil {
		return fmt.Errorf("title llm: %w", err)
	}

	var result model.TitleOptionsResult
	if err := llmjson.Unmarshal(content, &result); err != nil {
		return fmt.Errorf("parse title options: %w", err)
	}
	if len(result.Options) == 0 {
		return fmt.Errorf("title options empty")
	}

	state.TitleOptions = &result
	state.Phase = model.ComicPhaseTitleSelecting
	return nil
}
