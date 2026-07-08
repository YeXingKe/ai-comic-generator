package agents

import (
	"context"
	"fmt"

	"github.com/ai-comic-generator/server/internal/common"
	"github.com/ai-comic-generator/server/internal/model"
	"github.com/ai-comic-generator/server/internal/pkg/llmjson"
	"github.com/ai-comic-generator/server/internal/service"
	"github.com/tmc/langchaingo/llms"
)

// TitleAgent 标题推荐 Agent（流水线第 0 步，qwen-plus）
type TitleAgent struct {
	llm llms.Model
	agentLogService *service.AgentLogService // 智能体执行日志服务
}

func NewTitleAgent(llm llms.Model) *TitleAgent {
	return &TitleAgent{llm: llm}
}

func (a *TitleAgent) Execute(ctx context.Context, state *model.ComicState) error {
	startTime := time.Now() // 记录开始时间，用于计算耗时
	agentLog := &model.AgentLog{
		TaskID:    state.TaskID,           // 关联文章生成任务 ID
		AgentName: common.Agent1TitleAgent, // 智能体名称，与旧版单链路 Agent1 保持一致
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
	prompt := common.BuildTitleIdeationPrompt(state.Topic, state.Style, state.UserDescription)
	agentLog.Prompt = &prompt    
	content, err := llms.GenerateFromSinglePrompt(ctx, a.llm, prompt)
	if err != nil {
		agentLog.Status = "FAILED" // 标记为失败
		agentLog.ErrorMessage = "标题推荐 LLM 调用失败" // 设置错误信息
		return fmt.Errorf("title llm: %w", err)
	}

	var result model.TitleOptionsResult
	if err := llmjson.Unmarshal(content, &result); err != nil {
		agentLog.Status = "FAILED" // 标记为失败
		agentLog.ErrorMessage = "标题推荐 JSON 解析失败" // 设置错误信息
		return fmt.Errorf("parse title options: %w", err)
	}
	if len(result.Options) == 0 {
		agentLog.Status = "FAILED" // 标记为失败
		agentLog.ErrorMessage = "标题推荐 JSON 解析失败" // 设置错误信息
		return fmt.Errorf("title options empty")
	}

	state.TitleOptions = &result
	state.Phase = model.ComicPhaseTitleSelecting
	return nil
}
