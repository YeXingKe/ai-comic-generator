package agents // 角色设定 Agent 包

import (
	"context"         // 上下文，控制 LLM 调用超时与取消
	"encoding/json"   // 将故事构思结构体序列化为 JSON 字符串
	"fmt"             // 格式化错误信息

	"github.com/ai-comic-generator/server/internal/common"       // Prompt 组装与公共常量
	"github.com/ai-comic-generator/server/internal/model"        // 漫画领域模型（ComicState、ComicCharacter）
	"github.com/ai-comic-generator/server/internal/pkg/llmjson"    // 从 LLM 原始输出中提取并解析 JSON
	"github.com/tmc/langchaingo/llms"                            // LangChainGo LLM 抽象接口
)

// CharacterAgent 角色设定 Agent（流水线第 2 步，qwen-plus）
type CharacterAgent struct {
	llm llms.Model // 注入的大语言模型实例（通义千问 qwen-plus）
}

// NewCharacterAgent 创建角色设定 Agent
func NewCharacterAgent(llm llms.Model) *CharacterAgent {
	return &CharacterAgent{llm: llm} // 保存 LLM 引用供 Execute 调用
}

// Execute 根据故事构思生成角色列表，写入 state.Characters
func (a *CharacterAgent) Execute(ctx context.Context, state *model.ComicState) error {
	if state.StoryIdeation == nil { // 前置步骤（故事构思）必须已完成
		return fmt.Errorf("story ideation is empty") // 缺少故事数据则终止本步骤
	}
	storyJSON, err := json.Marshal(state.StoryIdeation) // 将故事构思序列化为 JSON 供 Prompt 使用
	if err != nil { // 序列化失败（理论上不应发生）
		return fmt.Errorf("marshal story: %w", err) // 包装错误并返回
	}

	prompt := common.BuildCharacterDesignPrompt(string(storyJSON), state.Style) // 组装角色设定 Prompt（含风格）
	content, err := llms.GenerateFromSinglePrompt(ctx, a.llm, prompt)          // 调用 qwen-plus 生成角色 JSON
	if err != nil { // LLM 调用失败（网络、鉴权、限流等）
		return fmt.Errorf("character llm: %w", err) // 包装 LLM 错误
	}

	var characters []model.ComicCharacter                                    // 声明角色列表接收变量
	if err := llmjson.Unmarshal(content, &characters); err != nil {          // 从 LLM 输出中解析 JSON 数组
		return fmt.Errorf("parse characters: %w", err) // 解析失败（格式不符或缺少字段）
	}
	if len(characters) == 0 { // 解析成功但结果为空
		return fmt.Errorf("characters empty") // 无有效角色则视为步骤失败
	}

	state.Characters = characters                         // 将角色列表写入流水线内存态
	state.Phase = model.ComicPhaseCharacterDesign         // 更新当前阶段为「角色设定」
	return nil                                            // 本步骤成功完成
}
