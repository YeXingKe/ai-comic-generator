package agents // 故事构思 Agent 包

import (
	"context" // 上下文，控制 LLM 调用超时与取消
	"fmt"     // 格式化错误信息

	"github.com/tmc/langchaingo/llms"                         // LangChainGo LLM 抽象接口
	"github.com/ai-comic-generator/server/internal/common"    // Prompt 组装函数
	"github.com/ai-comic-generator/server/internal/model"     // 漫画领域模型
	"github.com/ai-comic-generator/server/internal/pkg/llmjson" // LLM JSON 解析工具
)

// StoryAgent 故事构思 Agent（流水线第 1 步）
type StoryAgent struct {
	llm llms.Model // 注入的大语言模型实例（通义千问 qwen-plus）
}

// NewStoryAgent 创建故事构思 Agent
func NewStoryAgent(llm llms.Model) *StoryAgent {
	return &StoryAgent{llm: llm} // 保存 LLM 引用供 Execute 调用
}

// Execute 根据用户主题生成故事构思，写入 state.StoryIdeation
func (a *StoryAgent) Execute(ctx context.Context, state *model.ComicState) error {
	prompt := common.BuildStoryIdeationPrompt(state.Topic, state.Style, state.UserDescription) // 组装故事构思 Prompt

	content, err := llms.GenerateFromSinglePrompt(ctx, a.llm, prompt) // 调用 qwen-plus 生成故事 JSON
	if err != nil { // LLM 调用失败
		return fmt.Errorf("story llm: %w", err) // 包装 LLM 错误
	}

	var result model.StoryIdeationResult                         // 声明故事构思结果接收变量
	if err := llmjson.Unmarshal(content, &result); err != nil { // 从 LLM 输出中解析 JSON 对象
		return fmt.Errorf("parse story: %w", err) // 解析失败则终止
	}

	state.StoryIdeation = &result                    // 将故事构思写入流水线内存态
	state.Phase = model.ComicPhaseStoryIdeation      // 更新当前阶段为「故事构思」
	return nil                                       // 本步骤成功完成
}
