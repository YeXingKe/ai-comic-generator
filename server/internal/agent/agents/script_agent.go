package agents // 分镜脚本 Agent 包

import (
	"context"         // 上下文，控制 LLM 调用超时与取消
	"encoding/json"   // 将故事与角色序列化为 JSON 字符串
	"fmt"             // 格式化错误信息

	"github.com/ai-comic-generator/server/internal/common"       // Prompt 组装与公共常量
	"github.com/ai-comic-generator/server/internal/model"        // 漫画领域模型
	"github.com/ai-comic-generator/server/internal/pkg/llmjson"    // LLM JSON 解析工具
	"github.com/tmc/langchaingo/llms"                            // LangChainGo LLM 抽象接口
	"github.com/ai-comic-generator/server/internal/service"
)

const defaultPanelCount = 4 // 默认分镜格数（四格漫画）

// ScriptAgent 分镜脚本 Agent（流水线第 3 步，qwen-plus）
type ScriptAgent struct {
	llm llms.Model // 注入的大语言模型实例（通义千问 qwen-plus）
	agentLogService *service.AgentLogService // 智能体执行日志服务
}

// NewScriptAgent 创建分镜脚本 Agent
func NewScriptAgent(llm llms.Model) *ScriptAgent {
	return &ScriptAgent{llm: llm} // 保存 LLM 引用供 Execute 调用
}

// Execute 根据故事与角色生成分镜脚本，写入 state.Storyboard
func (a *ScriptAgent) Execute(ctx context.Context, state *model.ComicState) error {
	startTime := time.Now() // 记录开始时间，用于计算耗时
	agentLog := &model.AgentLog{
		TaskID:    state.TaskID,           // 关联文章生成任务 ID
		AgentName: common.Agent4ScriptAgent, // 智能体名称，与旧版单链路 Agent1 保持一致
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
	if state.StoryIdeation == nil { // 前置步骤（故事构思）必须已完成
		agentLog.Status = "FAILED" // 标记为失败
		agentLog.ErrorMessage = "前置步骤（故事构思）必须已完成" // 设置错误信息
		return fmt.Errorf("story ideation is empty") // 缺少故事数据则终止
	}
	if len(state.Characters) == 0 { // 前置步骤（角色设定）必须已完成
		agentLog.Status = "FAILED" // 标记为失败
		agentLog.ErrorMessage = "前置步骤（角色设定）必须已完成" // 设置错误信息
		return fmt.Errorf("characters empty") // 缺少角色数据则终止
	}

	storyJSON, _ := json.Marshal(state.StoryIdeation) // 序列化故事构思（忽略序列化错误，结构体必定可序列化）
	charJSON, _ := json.Marshal(state.Characters)      // 序列化角色列表

	prompt := common.BuildStoryboardScriptPrompt( // 组装分镜脚本 Prompt
		string(storyJSON), string(charJSON), state.Style, state.UserDescription, defaultPanelCount, // 传入故事、角色、风格、描述、格数
	)
	agentLog.Prompt = &prompt    
	content, err := llms.GenerateFromSinglePrompt(ctx, a.llm, prompt) // 调用 qwen-plus 生成分镜 JSON
	if err != nil { // LLM 调用失败
		agentLog.Status = "FAILED" // 标记为失败
		agentLog.ErrorMessage = "分镜脚本 LLM 调用失败" // 设置错误信息
		return fmt.Errorf("script llm: %w", err) // 包装 LLM 错误
	}

	var storyboard model.StoryboardResult                        // 声明分镜脚本结果接收变量
	if err := llmjson.Unmarshal(content, &storyboard); err != nil { // 从 LLM 输出中解析 JSON 对象
		agentLog.Status = "FAILED" // 标记为失败
		agentLog.ErrorMessage = "分镜脚本 JSON 解析失败" // 设置错误信息
		return fmt.Errorf("parse storyboard: %w", err) // 解析失败则终止
	}
	if len(storyboard.Panels) == 0 { // 解析成功但分镜格为空
		agentLog.Status = "FAILED" // 标记为失败
		agentLog.ErrorMessage = "分镜脚本 JSON 解析失败" // 设置错误信息
		return fmt.Errorf("storyboard panels empty") // 无有效分镜则视为步骤失败
	}

	state.Storyboard = &storyboard                      // 将分镜脚本写入流水线内存态
	state.Phase = model.ComicPhaseStoryboardScript      // 更新当前阶段为「分镜脚本」
	return nil                                          // 本步骤成功完成
}
