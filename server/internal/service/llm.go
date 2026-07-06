package service // 业务逻辑层：LLM 客户端初始化

import (
	"fmt" // 格式化配置缺失等错误信息
	"log" // 启动时打印 LLM 初始化日志

	"github.com/ai-comic-generator/server/internal/config" // 读取 DashScope 配置
	"github.com/tmc/langchaingo/llms"                      // LangChainGo LLM 抽象接口
	"github.com/tmc/langchaingo/llms/openai"               // OpenAI 兼容客户端（用于通义千问）
)

// NewLLM 创建通义千问客户端（qwen-plus，OpenAI 兼容）
func NewLLM(cfg *config.Config) (llms.Model, error) {
	ds := cfg.AI.DashScope                                     // 取出 DashScope 配置段
	if ds.APIKey == "" || ds.APIKey == "sk-xxx" {              // API Key 未配置或为占位符
		return nil, fmt.Errorf("dashscope api_key 未配置") // 返回错误，阻止启动漫画模块
	}
	log.Printf("init dashscope llm: model=%s base=%s", ds.Model, ds.BaseURL) // 记录模型与 BaseURL
	return openai.New( // 通过 OpenAI 兼容模式连接通义千问
		openai.WithToken(ds.APIKey),   // 设置 API Key
		openai.WithModel(ds.Model),    // 设置模型名（如 qwen-plus）
		openai.WithBaseURL(ds.BaseURL), // 设置兼容接口地址
	)
}
