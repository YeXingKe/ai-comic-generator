package agent // Agent 接口定义包

import (
	"context" // 上下文，贯穿流水线各步骤
	"github.com/ai-comic-generator/server/internal/model" // 漫画编排内存态 ComicState
)

// Agent 漫画流水线中的单步智能体（步骤 1–3 用 LLM）
type Agent interface {
	Execute(ctx context.Context, state *model.ComicState) error // 执行单步逻辑，读写 state，失败返回 error
}
