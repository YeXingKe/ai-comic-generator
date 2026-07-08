package agents // 故事构思 Agent 包

import (
	"context" // 上下文，控制 LLM 调用超时与取消
	"fmt"     // 格式化错误信息

	"github.com/tmc/langchaingo/llms"                         // LangChainGo LLM 抽象接口
	"github.com/ai-comic-generator/server/internal/common"    // Prompt 组装函数
	"github.com/ai-comic-generator/server/internal/model"     // 漫画领域模型
	"github.com/ai-comic-generator/server/internal/pkg/llmjson" // LLM JSON 解析工具
	"github.com/ai-comic-generator/server/internal/service"
)

// StoryAgent 故事构思 Agent（流水线第 1 步）
type StoryAgent struct {
	llm llms.Model // 注入的大语言模型实例（通义千问 qwen-plus）
	agentLogService *service.AgentLogService // 智能体执行日志服务
}

// NewStoryAgent 创建故事构思 Agent
func NewStoryAgent(llm llms.Model) *StoryAgent {
	return &StoryAgent{llm: llm} // 保存 LLM 引用供 Execute 调用
}

// Execute 根据用户主题生成故事构思，写入 state.StoryIdeation
func (a *StoryAgent) Execute(ctx context.Context, state *model.ComicState) error {
	startTime := time.Now() // 记录开始时间，用于计算耗时
	agentLog := &model.AgentLog{
		TaskID:    state.TaskID,           // 关联文章生成任务 ID
		AgentName: common.Agent2StoryAgent, // 智能体名称，与旧版单链路 Agent1 保持一致
		StartTime: startTime,              // 开始时间
		Status:    "RUNNING",                // 运行中
	}
	// 使用 defer 确保无论成功或失败，日志都会在函数退出前异步保存
	defer func() {
		endTime := time.Now()                               // 结束时间
		agentLog.EndTime = &endTime                         // 写入日志结束时间
		duration := int(time.Since(startTime).Milliseconds()) // 耗时（毫秒）
		agentLog.DurationMs = &duration                     // 写入耗时
		a.agentLogService.SaveLogAsync(agentLog)          // 异步落库，不阻塞主流程
	}()

	prompt := common.BuildStoryIdeationPrompt(state.Topic, state.Style, state.UserDescription, state.SelectedTitle)
	agentLog.Prompt = &prompt    
	content, err := llms.GenerateFromSinglePrompt(ctx, a.llm, prompt) // 调用 qwen-plus 生成故事 JSON
	if err != nil { // LLM 调用失败
		agentLog.Status = "FAILED" // 标记为失败
		agentLog.ErrorMessage = "故事构思 LLM 调用失败" // 设置错误信息
		return fmt.Errorf("story llm: %w", err) // 包装 LLM 错误
	}

	var result model.StoryIdeationResult
	if err := llmjson.Unmarshal(content, &result); err != nil {
		agentLog.Status = "FAILED" // 标记为失败
		agentLog.ErrorMessage = "故事构思 JSON 解析失败" // 设置错误信息
		return fmt.Errorf("parse story: %w", err)
	}
	if state.SelectedTitle != "" {
		result.Title = state.SelectedTitle
	}

	state.StoryIdeation = &result
	state.Phase = model.ComicPhaseStoryIdeation      // 更新当前阶段为「故事构思」
	return nil                                       // 本步骤成功完成
}
