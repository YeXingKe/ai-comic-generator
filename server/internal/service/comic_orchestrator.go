package service // 业务逻辑层：漫画六步流水线编排

import (
	"context" // 贯穿各步骤的上下文（LLM/生图/上传超时控制）
	"fmt"     // 包装步骤失败错误
	"log"     // 记录每步开始日志

	"github.com/ai-comic-generator/server/internal/agent"        // Agent 接口定义
	"github.com/ai-comic-generator/server/internal/agent/agents" // 故事/角色/分镜三个 LLM Agent
	"github.com/ai-comic-generator/server/internal/model"        // 漫画流水线内存态
	"github.com/ai-comic-generator/server/internal/store"      // 持久化阶段与产物
	"github.com/tmc/langchaingo/llms"                            // LLM 实例，用于创建 Agent
)

// ComicOrchestrator 漫画六步流水线编排器
type ComicOrchestrator struct {
	comicStore     *store.ComicStore // 任务状态与产物持久化
	storyAgent     agent.Agent       // 第 1 步：故事构思（qwen-plus）
	characterAgent agent.Agent       // 第 2 步：角色设定（qwen-plus）
	scriptAgent    agent.Agent       // 第 3 步：分镜脚本（qwen-plus）
	imageService   *ImageService     // 第 4 步：画面生成（混元/占位图）
	composeService *ComposeService   // 第 5 步：排版合成（imaging+gg）
	publishService *PublishService   // 第 6 步：公众号发布
}

// NewComicOrchestrator 组装六步流水线所需全部依赖
func NewComicOrchestrator(
	llm llms.Model, // 通义千问 LLM 实例
	comicStore *store.ComicStore, // 漫画 Store
	imageSvc *ImageService, // 生图服务
	composeSvc *ComposeService, // 排版服务
	publishSvc *PublishService, // 发布服务
) *ComicOrchestrator {
	return &ComicOrchestrator{ // 注入各步骤执行器
		comicStore:     comicStore,                    // 持久化
		storyAgent:     agents.NewStoryAgent(llm),     // 故事 Agent
		characterAgent: agents.NewCharacterAgent(llm), // 角色 Agent
		scriptAgent:    agents.NewScriptAgent(llm),    // 分镜 Agent
		imageService:   imageSvc,                      // 生图
		composeService: composeSvc,                    // 排版
		publishService: publishSvc,                    // 发布
	}
}

// Run 顺序执行六步流水线，每步完成后 SyncState 到数据库
func (o *ComicOrchestrator) Run(ctx context.Context, state *model.ComicState) error {
	steps := []struct { // 定义六步执行顺序
		phase string                                              // 阶段常量名
		fn    func(context.Context, *model.ComicState) error      // 该步执行函数
	}{
		{model.ComicPhaseStoryIdeation, o.storyAgent.Execute},         // 1. 故事构思
		{model.ComicPhaseCharacterDesign, o.characterAgent.Execute},   // 2. 角色设定
		{model.ComicPhaseStoryboardScript, o.scriptAgent.Execute},     // 3. 分镜脚本
		{model.ComicPhaseImageGeneration, o.imageService.GeneratePanels}, // 4. 画面生成
		{model.ComicPhaseLayoutCompose, o.composeService.Compose},     // 5. 排版合成
		{model.ComicPhaseWechatPublish, o.publishService.Publish},   // 6. 公众号发布
	}

	for _, step := range steps { // 逐步执行
		log.Printf("comic step start: taskId=%s phase=%s", state.TaskID, step.phase) // 记录步骤开始
		if err := o.comicStore.UpdatePhase(state.TaskID, model.ComicStatusProcessing, step.phase); err != nil { // 更新 DB 为进行中
			return err // 更新失败则终止流水线
		}
		if err := step.fn(ctx, state); err != nil { // 执行当前步骤
			_ = o.comicStore.MarkFailed(state.TaskID, step.phase, err.Error()) // 标记任务失败并记录错误
			return fmt.Errorf("%s: %w", step.phase, err)                       // 包装步骤名与原始错误
		}
		if err := o.comicStore.SyncState(state); err != nil { // 将本步产物同步到数据库 JSON 列
			return err // 同步失败则终止
		}
	}
	return o.comicStore.MarkCompleted(state.TaskID) // 全部成功，标记任务完成
}
